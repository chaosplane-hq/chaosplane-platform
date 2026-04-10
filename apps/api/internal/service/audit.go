package service

import (
	"context"
	"fmt"
	"time"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

type AuditService struct {
	pool *database.Pool
}

func NewAuditService(pool *database.Pool) *AuditService {
	return &AuditService{pool: pool}
}

type AuditLogEntry struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenantId"`
	UserID         *string   `json:"userId,omitempty"`
	Action         string    `json:"action"`
	ResourceType   string    `json:"resourceType"`
	ResourceID     *string   `json:"resourceId,omitempty"`
	IPAddress      *string   `json:"ipAddress,omitempty"`
	RequestMethod  *string   `json:"requestMethod,omitempty"`
	RequestPath    *string   `json:"requestPath,omitempty"`
	ResponseStatus *int      `json:"responseStatus,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

type ListAuditLogsResponse struct {
	Items      []AuditLogEntry `json:"items"`
	TotalCount int             `json:"totalCount"`
}

type ListAuditLogsRequest struct {
	Action       string
	ResourceType string
	UserID       string
	Limit        int
	Offset       int
}

func (s *AuditService) List(ctx context.Context, actor ActorContext, req *ListAuditLogsRequest) (*ListAuditLogsResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}

	query := `
		SELECT id::text, tenant_id::text, user_id::text, action, resource_type, resource_id,
		       host(ip_address), request_method, request_path, response_status, created_at
		FROM audit_logs
		WHERE tenant_id = $1::uuid
	`
	args := []any{actor.TenantID}
	argIdx := 2

	if req.Action != "" {
		query += fmt.Sprintf(" AND action = $%d", argIdx)
		args = append(args, req.Action)
		argIdx++
	}
	if req.ResourceType != "" {
		query += fmt.Sprintf(" AND resource_type = $%d", argIdx)
		args = append(args, req.ResourceType)
		argIdx++
	}
	if req.UserID != "" {
		query += fmt.Sprintf(" AND user_id = $%d::uuid", argIdx)
		args = append(args, req.UserID)
		argIdx++
	}

	var totalCount int
	countQuery := "SELECT COUNT(*) FROM (" + query + ") sub"
	if err := s.pool.App.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("count audit logs: %w", err)
	}

	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, req.Limit, req.Offset)

	rows, err := s.pool.App.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list audit logs: %w", err)
	}
	defer rows.Close()

	items := []AuditLogEntry{}
	for rows.Next() {
		var entry AuditLogEntry
		if err := rows.Scan(
			&entry.ID, &entry.TenantID, &entry.UserID, &entry.Action, &entry.ResourceType, &entry.ResourceID,
			&entry.IPAddress, &entry.RequestMethod, &entry.RequestPath, &entry.ResponseStatus, &entry.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		items = append(items, entry)
	}

	return &ListAuditLogsResponse{Items: items, TotalCount: totalCount}, rows.Err()
}
