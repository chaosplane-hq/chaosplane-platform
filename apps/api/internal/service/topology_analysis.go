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

type TopologyAnalysisService struct {
	pool *database.Pool
}

func NewTopologyAnalysisService(pool *database.Pool) *TopologyAnalysisService {
	return &TopologyAnalysisService{pool: pool}
}

type ServiceDependency struct {
	ID              string    `json:"id"`
	SourceKind      string    `json:"sourceKind"`
	SourceName      string    `json:"sourceName"`
	SourceNamespace string    `json:"sourceNamespace"`
	TargetKind      string    `json:"targetKind"`
	TargetName      string    `json:"targetName"`
	TargetNamespace string    `json:"targetNamespace"`
	Protocol        *string   `json:"protocol,omitempty"`
	Port            *int      `json:"port,omitempty"`
	LastSeenAt      time.Time `json:"lastSeenAt"`
}

type TopologyDrift struct {
	ID                string          `json:"id"`
	DriftType         string          `json:"driftType"`
	ResourceKind      string          `json:"resourceKind"`
	ResourceName      string          `json:"resourceName"`
	ResourceNamespace string          `json:"resourceNamespace"`
	PreviousState     json.RawMessage `json:"previousState,omitempty"`
	CurrentState      json.RawMessage `json:"currentState,omitempty"`
	DetectedAt        time.Time       `json:"detectedAt"`
	AcknowledgedAt    *time.Time      `json:"acknowledgedAt,omitempty"`
}

type EnvironmentMetric struct {
	MetricName  string          `json:"metricName"`
	MetricValue float64         `json:"metricValue"`
	Labels      json.RawMessage `json:"labels"`
	CollectedAt time.Time       `json:"collectedAt"`
}

type DependencyMapResponse struct {
	Dependencies []ServiceDependency `json:"dependencies"`
	NodeCount    int                 `json:"nodeCount"`
	EdgeCount    int                 `json:"edgeCount"`
}

type DriftListResponse struct {
	Items      []TopologyDrift `json:"items"`
	TotalCount int             `json:"totalCount"`
}

type MetricsResponse struct {
	Items []EnvironmentMetric `json:"items"`
}

type SubmitDependenciesRequest struct {
	EnvironmentID string            `json:"environmentId" binding:"required,uuid"`
	Dependencies  []DependencyInput `json:"dependencies" binding:"required"`
}

type DependencyInput struct {
	SourceKind      string  `json:"sourceKind" binding:"required"`
	SourceName      string  `json:"sourceName" binding:"required"`
	SourceNamespace string  `json:"sourceNamespace" binding:"required"`
	TargetKind      string  `json:"targetKind" binding:"required"`
	TargetName      string  `json:"targetName" binding:"required"`
	TargetNamespace string  `json:"targetNamespace" binding:"required"`
	Protocol        *string `json:"protocol,omitempty"`
	Port            *int    `json:"port,omitempty"`
}

type SubmitMetricsRequest struct {
	EnvironmentID string        `json:"environmentId" binding:"required,uuid"`
	Metrics       []MetricInput `json:"metrics" binding:"required"`
}

type MetricInput struct {
	Name   string          `json:"name" binding:"required"`
	Value  float64         `json:"value" binding:"required"`
	Labels json.RawMessage `json:"labels,omitempty"`
}

type SubmitDriftRequest struct {
	EnvironmentID     string          `json:"environmentId" binding:"required,uuid"`
	DriftType         string          `json:"driftType" binding:"required"`
	ResourceKind      string          `json:"resourceKind" binding:"required"`
	ResourceName      string          `json:"resourceName" binding:"required"`
	ResourceNamespace string          `json:"resourceNamespace" binding:"required"`
	PreviousState     json.RawMessage `json:"previousState,omitempty"`
	CurrentState      json.RawMessage `json:"currentState,omitempty"`
}

func (s *TopologyAnalysisService) GetDependencyMap(ctx context.Context, actor ActorContext, environmentID string) (*DependencyMapResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}

	rows, err := s.pool.App.Query(ctx, `
		SELECT id::text, source_kind, source_name, source_namespace,
		       target_kind, target_name, target_namespace, protocol, port, last_seen_at
		FROM service_dependencies
		WHERE environment_id = $1::uuid AND tenant_id = $2::uuid
		ORDER BY last_seen_at DESC
	`, environmentID, actor.TenantID)
	if err != nil {
		return nil, fmt.Errorf("get dependency map: %w", err)
	}
	defer rows.Close()

	deps := []ServiceDependency{}
	nodes := map[string]struct{}{}
	for rows.Next() {
		var d ServiceDependency
		if err := rows.Scan(&d.ID, &d.SourceKind, &d.SourceName, &d.SourceNamespace,
			&d.TargetKind, &d.TargetName, &d.TargetNamespace, &d.Protocol, &d.Port, &d.LastSeenAt); err != nil {
			return nil, fmt.Errorf("scan dependency: %w", err)
		}
		deps = append(deps, d)
		nodes[d.SourceNamespace+"/"+d.SourceKind+"/"+d.SourceName] = struct{}{}
		nodes[d.TargetNamespace+"/"+d.TargetKind+"/"+d.TargetName] = struct{}{}
	}

	return &DependencyMapResponse{Dependencies: deps, NodeCount: len(nodes), EdgeCount: len(deps)}, rows.Err()
}

func (s *TopologyAnalysisService) GetDrifts(ctx context.Context, actor ActorContext, environmentID string, unackedOnly bool, limit int) (*DriftListResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	query := `
		SELECT id::text, drift_type, resource_kind, resource_name, resource_namespace,
		       previous_state, current_state, detected_at, acknowledged_at
		FROM topology_drifts
		WHERE environment_id = $1::uuid AND tenant_id = $2::uuid
	`
	args := []any{environmentID, actor.TenantID}
	if unackedOnly {
		query += " AND acknowledged_at IS NULL"
	}
	query += " ORDER BY detected_at DESC LIMIT $3"
	args = append(args, limit)

	rows, err := s.pool.App.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get drifts: %w", err)
	}
	defer rows.Close()

	items := []TopologyDrift{}
	for rows.Next() {
		var d TopologyDrift
		if err := rows.Scan(&d.ID, &d.DriftType, &d.ResourceKind, &d.ResourceName, &d.ResourceNamespace,
			&d.PreviousState, &d.CurrentState, &d.DetectedAt, &d.AcknowledgedAt); err != nil {
			return nil, fmt.Errorf("scan drift: %w", err)
		}
		items = append(items, d)
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM topology_drifts WHERE environment_id = $1::uuid AND tenant_id = $2::uuid`
	if unackedOnly {
		countQuery += " AND acknowledged_at IS NULL"
	}
	_ = s.pool.App.QueryRow(ctx, countQuery, environmentID, actor.TenantID).Scan(&total)

	return &DriftListResponse{Items: items, TotalCount: total}, rows.Err()
}

func (s *TopologyAnalysisService) AcknowledgeDrift(ctx context.Context, actor ActorContext, driftID string) error {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return err
	}
	cmd, err := s.pool.App.Exec(ctx, `
		UPDATE topology_drifts
		SET acknowledged_at = now(), acknowledged_by = $3::uuid
		WHERE id = $1::uuid AND tenant_id = $2::uuid AND acknowledged_at IS NULL
	`, driftID, actor.TenantID, actor.UserID)
	if err != nil {
		return fmt.Errorf("acknowledge drift: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrHierarchyNotFound
	}
	return nil
}

func (s *TopologyAnalysisService) GetMetrics(ctx context.Context, actor ActorContext, environmentID, metricName string, limit int) (*MetricsResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	query := `
		SELECT metric_name, metric_value, labels, collected_at
		FROM environment_metrics
		WHERE environment_id = $1::uuid AND tenant_id = $2::uuid
	`
	args := []any{environmentID, actor.TenantID}
	if metricName != "" {
		query += " AND metric_name = $3"
		args = append(args, metricName)
		query += fmt.Sprintf(" ORDER BY collected_at DESC LIMIT $%d", len(args)+1)
	} else {
		query += fmt.Sprintf(" ORDER BY collected_at DESC LIMIT $%d", len(args)+1)
	}
	args = append(args, limit)

	rows, err := s.pool.App.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get metrics: %w", err)
	}
	defer rows.Close()

	items := []EnvironmentMetric{}
	for rows.Next() {
		var m EnvironmentMetric
		if err := rows.Scan(&m.MetricName, &m.MetricValue, &m.Labels, &m.CollectedAt); err != nil {
			return nil, fmt.Errorf("scan metric: %w", err)
		}
		items = append(items, m)
	}
	return &MetricsResponse{Items: items}, rows.Err()
}

func (s *TopologyAnalysisService) SubmitDependencies(ctx context.Context, tenantID string, req *SubmitDependenciesRequest) (int, error) {
	count := 0
	for _, dep := range req.Dependencies {
		labels := "{}"
		if dep.Protocol != nil || dep.Port != nil {
			_, err := s.pool.App.Exec(ctx, `
				INSERT INTO service_dependencies (tenant_id, environment_id, source_kind, source_name, source_namespace,
					target_kind, target_name, target_namespace, protocol, port)
				VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9, $10)
				ON CONFLICT (environment_id, source_kind, source_name, source_namespace, target_kind, target_name, target_namespace)
				DO UPDATE SET protocol = COALESCE($9, service_dependencies.protocol),
					port = COALESCE($10, service_dependencies.port),
					last_seen_at = now()
			`, tenantID, req.EnvironmentID, dep.SourceKind, dep.SourceName, dep.SourceNamespace,
				dep.TargetKind, dep.TargetName, dep.TargetNamespace, dep.Protocol, dep.Port)
			if err != nil {
				return count, fmt.Errorf("upsert dependency: %w", err)
			}
		} else {
			_, err := s.pool.App.Exec(ctx, `
				INSERT INTO service_dependencies (tenant_id, environment_id, source_kind, source_name, source_namespace,
					target_kind, target_name, target_namespace)
				VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8)
				ON CONFLICT (environment_id, source_kind, source_name, source_namespace, target_kind, target_name, target_namespace)
				DO UPDATE SET last_seen_at = now()
			`, tenantID, req.EnvironmentID, dep.SourceKind, dep.SourceName, dep.SourceNamespace,
				dep.TargetKind, dep.TargetName, dep.TargetNamespace)
			if err != nil {
				return count, fmt.Errorf("upsert dependency: %w", err)
			}
		}
		_ = labels
		count++
	}
	return count, nil
}

func (s *TopologyAnalysisService) SubmitMetrics(ctx context.Context, tenantID string, req *SubmitMetricsRequest) (int, error) {
	count := 0
	for _, m := range req.Metrics {
		labels := "{}"
		if len(m.Labels) > 0 {
			labels = string(m.Labels)
		}
		_, err := s.pool.App.Exec(ctx, `
			INSERT INTO environment_metrics (tenant_id, environment_id, metric_name, metric_value, labels)
			VALUES ($1::uuid, $2::uuid, $3, $4, $5::jsonb)
		`, tenantID, req.EnvironmentID, m.Name, m.Value, labels)
		if err != nil {
			return count, fmt.Errorf("insert metric: %w", err)
		}
		count++
	}
	return count, nil
}

func (s *TopologyAnalysisService) SubmitDrift(ctx context.Context, tenantID string, req *SubmitDriftRequest) (*TopologyDrift, error) {
	prev := "null"
	if len(req.PreviousState) > 0 {
		prev = string(req.PreviousState)
	}
	curr := "null"
	if len(req.CurrentState) > 0 {
		curr = string(req.CurrentState)
	}

	var drift TopologyDrift
	err := s.pool.App.QueryRow(ctx, `
		INSERT INTO topology_drifts (tenant_id, environment_id, drift_type, resource_kind, resource_name, resource_namespace, previous_state, current_state)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7::jsonb, $8::jsonb)
		RETURNING id::text, drift_type, resource_kind, resource_name, resource_namespace, previous_state, current_state, detected_at, acknowledged_at
	`, tenantID, req.EnvironmentID, req.DriftType, req.ResourceKind, req.ResourceName, req.ResourceNamespace, prev, curr).Scan(
		&drift.ID, &drift.DriftType, &drift.ResourceKind, &drift.ResourceName, &drift.ResourceNamespace,
		&drift.PreviousState, &drift.CurrentState, &drift.DetectedAt, &drift.AcknowledgedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("submit drift: %w", err)
	}
	return &drift, nil
}

var _ = errors.Is
var _ = pgx.ErrNoRows
