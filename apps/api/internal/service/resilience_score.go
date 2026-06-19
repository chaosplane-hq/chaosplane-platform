package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

type ResilienceScoreService struct {
	pool *database.Pool
}

func NewResilienceScoreService(pool *database.Pool) *ResilienceScoreService {
	return &ResilienceScoreService{pool: pool}
}

type ResilienceScore struct {
	ID             string          `json:"id"`
	EnvironmentID  string          `json:"environmentId"`
	OverallGrade   string          `json:"overallGrade"`
	OverallScore   float64         `json:"overallScore"`
	Availability   float64         `json:"availability"`
	FaultTolerance float64         `json:"faultTolerance"`
	Recoverability float64         `json:"recoverability"`
	Details        json.RawMessage `json:"details"`
	CalculatedAt   time.Time       `json:"calculatedAt"`
}

type ResilienceScoreResponse struct {
	Current *ResilienceScore  `json:"current"`
	History []ResilienceScore `json:"history"`
}

type CalculateScoreRequest struct {
	EnvironmentID string `json:"environmentId" binding:"required,uuid"`
}

func (s *ResilienceScoreService) Get(ctx context.Context, actor ActorContext, environmentID string) (*ResilienceScoreResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}

	rows, err := s.pool.Conn(ctx).Query(ctx, `
		SELECT id::text, environment_id::text, overall_grade, overall_score, availability, fault_tolerance, recoverability, details, calculated_at
		FROM resilience_scores
		WHERE environment_id = $1::uuid AND tenant_id = $2::uuid
		ORDER BY calculated_at DESC LIMIT 30
	`, environmentID, actor.TenantID)
	if err != nil {
		return nil, fmt.Errorf("get resilience scores: %w", err)
	}
	defer rows.Close()

	history := []ResilienceScore{}
	for rows.Next() {
		var s ResilienceScore
		if err := rows.Scan(&s.ID, &s.EnvironmentID, &s.OverallGrade, &s.OverallScore, &s.Availability, &s.FaultTolerance, &s.Recoverability, &s.Details, &s.CalculatedAt); err != nil {
			return nil, err
		}
		history = append(history, s)
	}

	var current *ResilienceScore
	if len(history) > 0 {
		current = &history[0]
	}
	return &ResilienceScoreResponse{Current: current, History: history}, rows.Err()
}

func (s *ResilienceScoreService) Calculate(ctx context.Context, tenantID string, req *CalculateScoreRequest) (*ResilienceScore, error) {
	availability := s.calculateAvailability(ctx, tenantID, req.EnvironmentID)
	faultTolerance := s.calculateFaultTolerance(ctx, tenantID, req.EnvironmentID)
	recoverability := s.calculateRecoverability(ctx, tenantID, req.EnvironmentID)

	overall := (availability*0.4 + faultTolerance*0.35 + recoverability*0.25)
	grade := scoreToGrade(overall)

	details, _ := json.Marshal(map[string]interface{}{
		"availabilityWeight":   0.4,
		"faultToleranceWeight": 0.35,
		"recoverabilityWeight": 0.25,
	})

	var score ResilienceScore
	err := s.pool.Conn(ctx).QueryRow(ctx, `
		INSERT INTO resilience_scores (tenant_id, environment_id, overall_grade, overall_score, availability, fault_tolerance, recoverability, details)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8::jsonb)
		RETURNING id::text, environment_id::text, overall_grade, overall_score, availability, fault_tolerance, recoverability, details, calculated_at
	`, tenantID, req.EnvironmentID, grade, overall, availability, faultTolerance, recoverability, string(details)).Scan(
		&score.ID, &score.EnvironmentID, &score.OverallGrade, &score.OverallScore,
		&score.Availability, &score.FaultTolerance, &score.Recoverability, &score.Details, &score.CalculatedAt)
	if err != nil {
		return nil, fmt.Errorf("store resilience score: %w", err)
	}
	return &score, nil
}

func (s *ResilienceScoreService) calculateAvailability(ctx context.Context, tenantID, envID string) float64 {
	var replicaScore float64
	var total, multiReplica int
	rows, err := s.pool.Conn(ctx).Query(ctx, `
		SELECT replicas FROM (
			SELECT (d->>'replicas')::int as replicas
			FROM topology_snapshots, jsonb_array_elements(deployments) d
			WHERE environment_id = $1::uuid AND tenant_id = $2::uuid
			ORDER BY collected_at DESC LIMIT 1
		) sub
	`, envID, tenantID)
	if err != nil {
		slog.Error("query availability replicas", "error", err)
		return -1
	}
	defer rows.Close()
	for rows.Next() {
		var r int
		if rows.Scan(&r) == nil {
			total++
			if r >= 2 {
				multiReplica++
			}
		}
	}
	if total > 0 {
		replicaScore = float64(multiReplica) / float64(total) * 100
	}
	return math.Min(replicaScore, 100)
}

func (s *ResilienceScoreService) calculateFaultTolerance(ctx context.Context, tenantID, envID string) float64 {
	var openVulns int
	if err := s.pool.Conn(ctx).QueryRow(ctx, `
		SELECT COUNT(*) FROM vulnerability_findings
		WHERE environment_id = $1::uuid AND tenant_id = $2::uuid AND status = 'open'
	`, envID, tenantID).Scan(&openVulns); err != nil {
		slog.Error("query fault tolerance vulns", "error", err)
		return -1
	}

	score := 100.0 - float64(openVulns)*10
	return math.Max(score, 0)
}

func (s *ResilienceScoreService) calculateRecoverability(ctx context.Context, tenantID, envID string) float64 {
	var completedExperiments int
	if err := s.pool.Conn(ctx).QueryRow(ctx, `
		SELECT COUNT(*) FROM experiment_results_analysis
		WHERE environment_id = $1::uuid AND tenant_id = $2::uuid
	`, envID, tenantID).Scan(&completedExperiments); err != nil {
		slog.Error("query recoverability experiments", "error", err)
		return -1
	}

	score := math.Min(float64(completedExperiments)*20, 100)
	return score
}

func scoreToGrade(score float64) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}
