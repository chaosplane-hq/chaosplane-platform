package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

var (
	ErrNoWork        = errors.New("no work available")
	ErrNotClaimable  = errors.New("experiment not claimable by this agent")
	claimLeaseWindow = 60 * time.Second
)

type AgentWorkService struct {
	pool *database.Pool
}

func NewAgentWorkService(pool *database.Pool) *AgentWorkService {
	return &AgentWorkService{pool: pool}
}

type AgentWorkItem struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	ExperimentType  string          `json:"experimentType"`
	Target          json.RawMessage `json:"target"`
	Action          json.RawMessage `json:"action"`
	Steps           json.RawMessage `json:"steps,omitempty"`
	SteadyState     json.RawMessage `json:"steadyState,omitempty"`
	Rollback        json.RawMessage `json:"rollback,omitempty"`
	AbortConditions json.RawMessage `json:"abortConditions,omitempty"`
	DurationSeconds int             `json:"durationSeconds"`
	DesiredState    string          `json:"desiredState"`
	Generation      int64           `json:"generation"`
	ClaimExpiresAt  time.Time       `json:"claimExpiresAt"`
}

type AgentStatusReport struct {
	Status             string          `json:"status" binding:"required"`
	ObservedGeneration *int64          `json:"observedGeneration,omitempty"`
	Phase              string          `json:"phase,omitempty"`
	Result             json.RawMessage `json:"result,omitempty"`
}

type AgentStatusAck struct {
	Acknowledged bool   `json:"acknowledged"`
	DesiredState string `json:"desiredState"`
	Generation   int64  `json:"generation"`
}

var agentReportableStatus = map[string]bool{
	"running":     true,
	"paused":      true,
	"completed":   true,
	"failed":      true,
	"aborted":     true,
	"rolled_back": true,
}

// ClaimWork atomically claims the next scheduled experiment for the agent's
// environment using FOR UPDATE SKIP LOCKED, so concurrent agent replicas never
// double-execute the same row. Expired leases on still-running rows are
// reclaimable for crash recovery.
func (s *AgentWorkService) ClaimWork(ctx context.Context, environmentID, agentInstance string) (*AgentWorkItem, error) {
	var w AgentWorkItem
	err := s.pool.Conn(ctx).QueryRow(ctx, `
		WITH claimable AS (
			SELECT id FROM experiments
			WHERE environment_id = $1::uuid
			  AND deleted_at IS NULL
			  AND (
			    (status = 'scheduled')
			    OR (status = 'running' AND claim_expires_at IS NOT NULL AND claim_expires_at < now())
			  )
			ORDER BY created_at
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		UPDATE experiments e
		SET status = 'running',
		    claimed_by = $2,
		    claim_expires_at = now() + make_interval(secs => $3),
		    run_started_at = COALESCE(e.run_started_at, now()),
		    last_agent_report_at = now()
		FROM claimable
		WHERE e.id = claimable.id
		RETURNING e.id::text, e.name, e.experiment_type, e.target, e.action, e.steps,
		          e.steady_state, e.rollback, e.abort_conditions, e.duration_seconds,
		          e.desired_state, e.generation, e.claim_expires_at
	`, environmentID, agentInstance, claimLeaseWindow.Seconds()).Scan(
		&w.ID, &w.Name, &w.ExperimentType, &w.Target, &w.Action, &w.Steps,
		&w.SteadyState, &w.Rollback, &w.AbortConditions, &w.DurationSeconds,
		&w.DesiredState, &w.Generation, &w.ClaimExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoWork
		}
		return nil, fmt.Errorf("claim work: %w", err)
	}
	return &w, nil
}

// ReportStatus records an agent's status update for an experiment it holds.
// The agent may only set agent-owned states, and only for a row in its own
// environment that it currently leases, preventing a stale replica from
// clobbering a reclaimed run or another environment's experiments.
func (s *AgentWorkService) ReportStatus(ctx context.Context, environmentID, experimentID, agentInstance string, req *AgentStatusReport) (*AgentStatusAck, error) {
	if !agentReportableStatus[req.Status] {
		return nil, fmt.Errorf("invalid agent status %q", req.Status)
	}

	terminal := req.Status == "completed" || req.Status == "failed" ||
		req.Status == "aborted" || req.Status == "rolled_back"

	var newExpiry interface{}
	var endedAt interface{}
	if terminal {
		newExpiry = nil
		endedAt = time.Now()
	} else {
		newExpiry = time.Now().Add(claimLeaseWindow)
		endedAt = nil
	}

	cmd, err := s.pool.Conn(ctx).Exec(ctx, `
		UPDATE experiments
		SET status = $4,
		    observed_generation = COALESCE($5, observed_generation),
		    claim_expires_at = $6,
		    run_ended_at = COALESCE($7, run_ended_at),
		    last_agent_report_at = now()
		WHERE id = $1::uuid
		  AND environment_id = $2::uuid
		  AND claimed_by = $3
		  AND deleted_at IS NULL
	`, experimentID, environmentID, agentInstance, req.Status, req.ObservedGeneration, newExpiry, endedAt)
	if err != nil {
		return nil, fmt.Errorf("report status: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return nil, ErrNotClaimable
	}

	if terminal {
		if err := s.recordResult(ctx, environmentID, experimentID, req.Status, req.Result); err != nil {
			return nil, err
		}
	}

	var ack AgentStatusAck
	ack.Acknowledged = true
	err = s.pool.Conn(ctx).QueryRow(ctx, `
		SELECT desired_state, generation FROM experiments
		WHERE id = $1::uuid AND environment_id = $2::uuid
	`, experimentID, environmentID).Scan(&ack.DesiredState, &ack.Generation)
	if err != nil {
		return nil, fmt.Errorf("read desired state: %w", err)
	}
	return &ack, nil
}

func (s *AgentWorkService) recordResult(ctx context.Context, environmentID, experimentID, status string, result json.RawMessage) error {
	resultStatus := status
	if resultStatus == "rolled_back" {
		resultStatus = "aborted"
	}

	var r struct {
		SteadyStateMet *bool           `json:"steadyStateMet"`
		ImpactSummary  json.RawMessage `json:"impactSummary"`
		Metrics        json.RawMessage `json:"metrics"`
		ErrorMessage   *string         `json:"errorMessage"`
	}
	if len(result) > 0 {
		_ = json.Unmarshal(result, &r)
	}
	impact := r.ImpactSummary
	if len(impact) == 0 {
		impact = json.RawMessage(`{}`)
	}
	metrics := r.Metrics
	if len(metrics) == 0 {
		metrics = json.RawMessage(`{}`)
	}

	_, err := s.pool.Conn(ctx).Exec(ctx, `
		INSERT INTO experiment_results
			(tenant_id, experiment_id, run_number, status, finished_at,
			 steady_state_met, impact_summary, metrics, error_message)
		SELECT e.tenant_id, e.id,
		       COALESCE((SELECT max(run_number) FROM experiment_results WHERE experiment_id = e.id), 0) + 1,
		       $3, now(), $4, $5::jsonb, $6::jsonb, $7
		FROM experiments e
		WHERE e.id = $1::uuid AND e.environment_id = $2::uuid
	`, experimentID, environmentID, resultStatus, r.SteadyStateMet, impact, metrics, r.ErrorMessage)
	if err != nil {
		return fmt.Errorf("record result: %w", err)
	}
	return nil
}
