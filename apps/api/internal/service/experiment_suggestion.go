package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

type ExperimentSuggestionService struct {
	pool *database.Pool
}

func NewExperimentSuggestionService(pool *database.Pool) *ExperimentSuggestionService {
	return &ExperimentSuggestionService{pool: pool}
}

type ExperimentSuggestion struct {
	ID              string          `json:"id"`
	EnvironmentID   string          `json:"environmentId"`
	FindingID       *string         `json:"findingId,omitempty"`
	Source          string          `json:"source"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	ActionType      string          `json:"actionType"`
	TargetNamespace string          `json:"targetNamespace"`
	TargetName      string          `json:"targetName"`
	Duration        string          `json:"duration"`
	Parameters      json.RawMessage `json:"parameters"`
	Confidence      float64         `json:"confidence"`
	CreatedAt       time.Time       `json:"createdAt"`
}

type SuggestionListResponse struct {
	Items []ExperimentSuggestion `json:"items"`
}

type GenerateSuggestionsRequest struct {
	EnvironmentID string `json:"environmentId" binding:"required,uuid"`
}

type GenerateSuggestionsResponse struct {
	Generated int `json:"generated"`
}

func (s *ExperimentSuggestionService) List(ctx context.Context, actor ActorContext, environmentID string) (*SuggestionListResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}

	rows, err := s.pool.Conn(ctx).Query(ctx, `
		SELECT id::text, environment_id::text, finding_id::text, source, title, description,
		       action_type, target_namespace, target_name, duration, parameters, confidence, created_at
		FROM experiment_suggestions
		WHERE environment_id = $1::uuid AND tenant_id = $2::uuid
		ORDER BY confidence DESC, created_at DESC
	`, environmentID, actor.TenantID)
	if err != nil {
		return nil, fmt.Errorf("list suggestions: %w", err)
	}
	defer rows.Close()

	items := []ExperimentSuggestion{}
	for rows.Next() {
		var s ExperimentSuggestion
		if err := rows.Scan(&s.ID, &s.EnvironmentID, &s.FindingID, &s.Source, &s.Title, &s.Description,
			&s.ActionType, &s.TargetNamespace, &s.TargetName, &s.Duration, &s.Parameters, &s.Confidence, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan suggestion: %w", err)
		}
		items = append(items, s)
	}
	return &SuggestionListResponse{Items: items}, rows.Err()
}

func (s *ExperimentSuggestionService) Generate(ctx context.Context, tenantID string, req *GenerateSuggestionsRequest) (*GenerateSuggestionsResponse, error) {
	vulnFindings, err := s.getOpenFindings(ctx, tenantID, req.EnvironmentID)
	if err != nil {
		return nil, err
	}

	bestPractices := s.getBestPracticeSuggestions(req.EnvironmentID)

	count := 0
	for _, f := range vulnFindings {
		if f.SuggestedExperiment == nil || len(f.SuggestedExperiment) == 0 {
			continue
		}
		var exp struct {
			Action    string `json:"action"`
			Namespace string `json:"namespace"`
			Target    string `json:"target"`
			Duration  string `json:"duration"`
		}
		if json.Unmarshal(f.SuggestedExperiment, &exp) != nil {
			continue
		}

		_, err := s.pool.Conn(ctx).Exec(ctx, `
			INSERT INTO experiment_suggestions (tenant_id, environment_id, finding_id, source, title, description,
				action_type, target_namespace, target_name, duration, parameters, confidence)
			VALUES ($1::uuid, $2::uuid, $3::uuid, 'vulnerability', $4, $5, $6, $7, $8, $9, $10::jsonb, $11)
			ON CONFLICT DO NOTHING
		`, tenantID, req.EnvironmentID, f.ID,
			fmt.Sprintf("Test resilience: %s", f.Title),
			fmt.Sprintf("Auto-generated from vulnerability finding: %s", f.Description),
			exp.Action, exp.Namespace, exp.Target, exp.Duration,
			string(f.SuggestedExperiment), 0.8)
		if err == nil {
			count++
		}
	}

	for _, bp := range bestPractices {
		_, err := s.pool.Conn(ctx).Exec(ctx, `
			INSERT INTO experiment_suggestions (tenant_id, environment_id, source, title, description,
				action_type, target_namespace, target_name, duration, parameters, confidence)
			VALUES ($1::uuid, $2::uuid, 'best_practice', $3, $4, $5, $6, $7, $8, $9::jsonb, $10)
			ON CONFLICT DO NOTHING
		`, tenantID, req.EnvironmentID,
			bp.Title, bp.Description, bp.ActionType, bp.TargetNamespace, bp.TargetName, bp.Duration,
			string(bp.Parameters), bp.Confidence)
		if err == nil {
			count++
		}
	}

	return &GenerateSuggestionsResponse{Generated: count}, nil
}

func (s *ExperimentSuggestionService) Delete(ctx context.Context, actor ActorContext, suggestionID string) error {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return err
	}
	cmd, err := s.pool.Conn(ctx).Exec(ctx, `
		DELETE FROM experiment_suggestions WHERE id = $1::uuid AND tenant_id = $2::uuid
	`, suggestionID, actor.TenantID)
	if err != nil {
		return fmt.Errorf("delete suggestion: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrHierarchyNotFound
	}
	return nil
}

type vulnFindingForSuggestion struct {
	ID                  string
	Title               string
	Description         string
	SuggestedExperiment json.RawMessage
}

func (s *ExperimentSuggestionService) getOpenFindings(ctx context.Context, tenantID, environmentID string) ([]vulnFindingForSuggestion, error) {
	rows, err := s.pool.Conn(ctx).Query(ctx, `
		SELECT id::text, title, description, suggested_experiment
		FROM vulnerability_findings
		WHERE environment_id = $1::uuid AND tenant_id = $2::uuid AND status = 'open' AND suggested_experiment IS NOT NULL
	`, environmentID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get open findings for suggestions: %w", err)
	}
	defer rows.Close()

	var items []vulnFindingForSuggestion
	for rows.Next() {
		var f vulnFindingForSuggestion
		if err := rows.Scan(&f.ID, &f.Title, &f.Description, &f.SuggestedExperiment); err != nil {
			return nil, err
		}
		items = append(items, f)
	}
	return items, rows.Err()
}

func (s *ExperimentSuggestionService) getBestPracticeSuggestions(environmentID string) []ExperimentSuggestion {
	return []ExperimentSuggestion{
		{
			Source: "best_practice", Title: "Random pod termination",
			Description: "Verify that your services recover gracefully from random pod failures.",
			ActionType:  "pod-kill", TargetNamespace: "default", TargetName: "*",
			Duration:   "60s",
			Parameters: json.RawMessage(`{"mode":"random-max-percent","value":"25"}`),
			Confidence: 0.9,
		},
		{
			Source: "best_practice", Title: "Network latency injection",
			Description: "Test service behavior under degraded network conditions.",
			ActionType:  "network-delay", TargetNamespace: "default", TargetName: "*",
			Duration:   "120s",
			Parameters: json.RawMessage(`{"latency":"200ms","jitter":"50ms"}`),
			Confidence: 0.85,
		},
		{
			Source: "best_practice", Title: "DNS failure simulation",
			Description: strings.ReplaceAll("Verify DNS failure handling and fallback mechanisms.", "'", ""),
			ActionType:  "pod-dns-error", TargetNamespace: "default", TargetName: "*",
			Duration:   "30s",
			Parameters: json.RawMessage(`{"domains":"*.external.com"}`),
			Confidence: 0.75,
		},
	}
}
