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

type ResultAnalysisService struct {
	pool *database.Pool
}

func NewResultAnalysisService(pool *database.Pool) *ResultAnalysisService {
	return &ResultAnalysisService{pool: pool}
}

type ResultAnalysis struct {
	ID                 string          `json:"id"`
	ExperimentName     string          `json:"experimentName"`
	EnvironmentID      *string         `json:"environmentId,omitempty"`
	Summary            string          `json:"summary"`
	ImpactAnalysis     *string         `json:"impactAnalysis,omitempty"`
	Recommendations    *string         `json:"recommendations,omitempty"`
	SeverityAssessment *string         `json:"severityAssessment,omitempty"`
	AffectedServices   json.RawMessage `json:"affectedServices"`
	MetricsImpact      json.RawMessage `json:"metricsImpact"`
	AnalyzedAt         time.Time       `json:"analyzedAt"`
}

type ResultAnalysisListResponse struct {
	Items []ResultAnalysis `json:"items"`
}

type AnalyzeResultRequest struct {
	ExperimentName string  `json:"experimentName" binding:"required"`
	EnvironmentID  *string `json:"environmentId,omitempty"`
}

func (s *ResultAnalysisService) List(ctx context.Context, actor ActorContext, environmentID string) (*ResultAnalysisListResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}

	query := `
		SELECT id::text, experiment_name, environment_id::text, summary, impact_analysis,
		       recommendations, severity_assessment, affected_services, metrics_impact, analyzed_at
		FROM experiment_results_analysis
		WHERE tenant_id = $1::uuid
	`
	args := []any{actor.TenantID}
	if environmentID != "" {
		query += " AND environment_id = $2::uuid"
		args = append(args, environmentID)
	}
	query += " ORDER BY analyzed_at DESC LIMIT 50"

	rows, err := s.pool.App.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list result analyses: %w", err)
	}
	defer rows.Close()

	items := []ResultAnalysis{}
	for rows.Next() {
		var r ResultAnalysis
		if err := rows.Scan(&r.ID, &r.ExperimentName, &r.EnvironmentID, &r.Summary, &r.ImpactAnalysis,
			&r.Recommendations, &r.SeverityAssessment, &r.AffectedServices, &r.MetricsImpact, &r.AnalyzedAt); err != nil {
			return nil, fmt.Errorf("scan result analysis: %w", err)
		}
		items = append(items, r)
	}
	return &ResultAnalysisListResponse{Items: items}, rows.Err()
}

func (s *ResultAnalysisService) Get(ctx context.Context, actor ActorContext, analysisID string) (*ResultAnalysis, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}

	var r ResultAnalysis
	err := s.pool.App.QueryRow(ctx, `
		SELECT id::text, experiment_name, environment_id::text, summary, impact_analysis,
		       recommendations, severity_assessment, affected_services, metrics_impact, analyzed_at
		FROM experiment_results_analysis
		WHERE id = $1::uuid AND tenant_id = $2::uuid
	`, analysisID, actor.TenantID).Scan(&r.ID, &r.ExperimentName, &r.EnvironmentID, &r.Summary, &r.ImpactAnalysis,
		&r.Recommendations, &r.SeverityAssessment, &r.AffectedServices, &r.MetricsImpact, &r.AnalyzedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrHierarchyNotFound
		}
		return nil, fmt.Errorf("get result analysis: %w", err)
	}
	return &r, nil
}

func (s *ResultAnalysisService) Analyze(ctx context.Context, tenantID string, req *AnalyzeResultRequest) (*ResultAnalysis, error) {
	summary := fmt.Sprintf("Experiment '%s' completed. Analysis pending LLM integration.", req.ExperimentName)
	impact := "Impact analysis will be generated when LLM API is configured."
	recommendations := "Recommendations will be generated when LLM API is configured."
	severity := "low"
	affectedServices := "[]"
	metricsImpact := "{}"

	var r ResultAnalysis
	err := s.pool.App.QueryRow(ctx, `
		INSERT INTO experiment_results_analysis (tenant_id, experiment_name, environment_id, summary,
			impact_analysis, recommendations, severity_assessment, affected_services, metrics_impact)
		VALUES ($1::uuid, $2, $3::uuid, $4, $5, $6, $7, $8::jsonb, $9::jsonb)
		RETURNING id::text, experiment_name, environment_id::text, summary, impact_analysis,
		          recommendations, severity_assessment, affected_services, metrics_impact, analyzed_at
	`, tenantID, req.ExperimentName, req.EnvironmentID, summary, impact, recommendations, severity,
		affectedServices, metricsImpact).Scan(
		&r.ID, &r.ExperimentName, &r.EnvironmentID, &r.Summary, &r.ImpactAnalysis,
		&r.Recommendations, &r.SeverityAssessment, &r.AffectedServices, &r.MetricsImpact, &r.AnalyzedAt)
	if err != nil {
		return nil, fmt.Errorf("create result analysis: %w", err)
	}
	return &r, nil
}
