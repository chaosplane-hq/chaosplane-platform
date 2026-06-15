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

type PredictiveAnalysisService struct {
	pool *database.Pool
	llm  *LLMClient
}

func NewPredictiveAnalysisService(pool *database.Pool, llm *LLMClient) *PredictiveAnalysisService {
	return &PredictiveAnalysisService{pool: pool, llm: llm}
}

type PredictiveAnalysis struct {
	ID                string          `json:"id"`
	EnvironmentID     string          `json:"environmentId"`
	PredictionType    string          `json:"predictionType"`
	Severity          string          `json:"severity"`
	Title             string          `json:"title"`
	Description       string          `json:"description"`
	Confidence        float64         `json:"confidence"`
	RecommendedAction *string         `json:"recommendedAction,omitempty"`
	AutoRemediation   json.RawMessage `json:"autoRemediation,omitempty"`
	Status            string          `json:"status"`
	PredictedAt       time.Time       `json:"predictedAt"`
}

type PredictiveListResponse struct {
	Items []PredictiveAnalysis `json:"items"`
}

type RunPredictionRequest struct {
	EnvironmentID string `json:"environmentId" binding:"required,uuid"`
}

type RunPredictionResponse struct {
	PredictionsCreated int `json:"predictionsCreated"`
}

func (s *PredictiveAnalysisService) List(ctx context.Context, actor ActorContext, environmentID string, status string) (*PredictiveListResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	query := `
		SELECT id::text, environment_id::text, prediction_type, severity, title, description,
		       confidence, recommended_action, auto_remediation, status, predicted_at
		FROM predictive_analyses WHERE tenant_id = $1::uuid
	`
	args := []any{actor.TenantID}
	argIdx := 2
	if environmentID != "" {
		query += fmt.Sprintf(" AND environment_id = $%d::uuid", argIdx)
		args = append(args, environmentID)
		argIdx++
	}
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	query += " ORDER BY predicted_at DESC LIMIT 50"
	rows, err := s.pool.Conn(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list predictions: %w", err)
	}
	defer rows.Close()
	items := []PredictiveAnalysis{}
	for rows.Next() {
		var p PredictiveAnalysis
		if err := rows.Scan(&p.ID, &p.EnvironmentID, &p.PredictionType, &p.Severity, &p.Title, &p.Description,
			&p.Confidence, &p.RecommendedAction, &p.AutoRemediation, &p.Status, &p.PredictedAt); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return &PredictiveListResponse{Items: items}, rows.Err()
}

func (s *PredictiveAnalysisService) Run(ctx context.Context, tenantID string, req *RunPredictionRequest) (*RunPredictionResponse, error) {
	predictions := s.generatePredictions(ctx, tenantID, req.EnvironmentID)
	count := 0
	for _, p := range predictions {
		remediation := "null"
		if len(p.AutoRemediation) > 0 {
			remediation = string(p.AutoRemediation)
		}
		_, err := s.pool.Conn(ctx).Exec(ctx, `
			INSERT INTO predictive_analyses (tenant_id, environment_id, prediction_type, severity, title, description, confidence, recommended_action, auto_remediation)
			VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9::jsonb)
		`, tenantID, req.EnvironmentID, p.PredictionType, p.Severity, p.Title, p.Description, p.Confidence, p.RecommendedAction, remediation)
		if err == nil {
			count++
		}
	}
	return &RunPredictionResponse{PredictionsCreated: count}, nil
}

func (s *PredictiveAnalysisService) UpdateStatus(ctx context.Context, actor ActorContext, predictionID, newStatus string) error {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return err
	}
	cmd, err := s.pool.Conn(ctx).Exec(ctx, `UPDATE predictive_analyses SET status = $3 WHERE id = $1::uuid AND tenant_id = $2::uuid`, predictionID, actor.TenantID, newStatus)
	if err != nil {
		return fmt.Errorf("update prediction status: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrHierarchyNotFound
	}
	return nil
}

func (s *PredictiveAnalysisService) generatePredictions(ctx context.Context, tenantID, envID string) []PredictiveAnalysis {
	var openVulns int
	_ = s.pool.Conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM vulnerability_findings WHERE environment_id = $1::uuid AND tenant_id = $2::uuid AND status = 'open' AND severity IN ('critical','high')`, envID, tenantID).Scan(&openVulns)

	var predictions []PredictiveAnalysis
	if openVulns > 3 {
		action := "Address critical and high severity vulnerabilities immediately to reduce failure risk."
		predictions = append(predictions, PredictiveAnalysis{
			PredictionType: "failure_risk", Severity: "high",
			Title: "Elevated failure risk detected", Description: fmt.Sprintf("Environment has %d unresolved critical/high vulnerabilities, indicating elevated failure risk.", openVulns),
			Confidence: 0.85, RecommendedAction: &action,
		})
	}
	return predictions
}

var _ = errors.Is
var _ = pgx.ErrNoRows
