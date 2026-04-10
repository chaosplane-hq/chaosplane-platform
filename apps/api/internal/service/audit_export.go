package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

type AuditExportService struct {
	pool *database.Pool
}

func NewAuditExportService(pool *database.Pool) *AuditExportService {
	return &AuditExportService{pool: pool}
}

type AuditExport struct {
	ID              string          `json:"id"`
	Destination     string          `json:"destination"`
	Config          json.RawMessage `json:"config"`
	Status          string          `json:"status"`
	RecordsExported int64           `json:"recordsExported"`
	ErrorMessage    *string         `json:"errorMessage,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
}

type AuditExportListResponse struct {
	Items []AuditExport `json:"items"`
}

type CreateAuditExportRequest struct {
	Destination string          `json:"destination" binding:"required"`
	Config      json.RawMessage `json:"config" binding:"required"`
}

func (s *AuditExportService) List(ctx context.Context, actor ActorContext) (*AuditExportListResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	rows, err := s.pool.App.Query(ctx, `
		SELECT id::text, destination, config, status, records_exported, error_message, created_at
		FROM audit_log_exports WHERE tenant_id = $1::uuid ORDER BY created_at DESC LIMIT 50
	`, actor.TenantID)
	if err != nil {
		return nil, fmt.Errorf("list audit exports: %w", err)
	}
	defer rows.Close()
	items := []AuditExport{}
	for rows.Next() {
		var e AuditExport
		if err := rows.Scan(&e.ID, &e.Destination, &e.Config, &e.Status, &e.RecordsExported, &e.ErrorMessage, &e.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	return &AuditExportListResponse{Items: items}, rows.Err()
}

func (s *AuditExportService) Create(ctx context.Context, actor ActorContext, req *CreateAuditExportRequest) (*AuditExport, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	var e AuditExport
	err := s.pool.App.QueryRow(ctx, `
		INSERT INTO audit_log_exports (tenant_id, destination, config)
		VALUES ($1::uuid, $2, $3::jsonb)
		RETURNING id::text, destination, config, status, records_exported, error_message, created_at
	`, actor.TenantID, req.Destination, string(req.Config)).Scan(
		&e.ID, &e.Destination, &e.Config, &e.Status, &e.RecordsExported, &e.ErrorMessage, &e.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create audit export: %w", err)
	}
	return &e, nil
}
