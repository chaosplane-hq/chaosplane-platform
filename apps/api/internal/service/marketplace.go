package service

import (
	"context"
	"fmt"
	"time"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

type MarketplaceService struct {
	pool *database.Pool
}

func NewMarketplaceService(pool *database.Pool) *MarketplaceService {
	return &MarketplaceService{pool: pool}
}

type MarketplacePlugin struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
	Description *string   `json:"description,omitempty"`
	Author      string    `json:"author"`
	Version     string    `json:"version"`
	Category    string    `json:"category"`
	Downloads   int64     `json:"downloads"`
	Rating      float64   `json:"rating"`
	Verified    bool      `json:"verified"`
	PublishedAt time.Time `json:"publishedAt"`
}

type MarketplaceListResponse struct {
	Items []MarketplacePlugin `json:"items"`
}

type InstallPluginRequest struct {
	PluginID string `json:"pluginId" binding:"required,uuid"`
}

func (s *MarketplaceService) List(ctx context.Context, category string, limit int) (*MarketplaceListResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	query := `SELECT id::text, name, display_name, description, author, version, category, downloads, rating, verified, published_at FROM marketplace_plugins`
	args := []any{}
	if category != "" {
		query += " WHERE category = $1"
		args = append(args, category)
		query += " ORDER BY downloads DESC LIMIT $2"
		args = append(args, limit)
	} else {
		query += " ORDER BY downloads DESC LIMIT $1"
		args = append(args, limit)
	}
	rows, err := s.pool.App.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list marketplace plugins: %w", err)
	}
	defer rows.Close()
	items := []MarketplacePlugin{}
	for rows.Next() {
		var p MarketplacePlugin
		if err := rows.Scan(&p.ID, &p.Name, &p.DisplayName, &p.Description, &p.Author, &p.Version, &p.Category, &p.Downloads, &p.Rating, &p.Verified, &p.PublishedAt); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return &MarketplaceListResponse{Items: items}, rows.Err()
}

func (s *MarketplaceService) Install(ctx context.Context, actor ActorContext, req *InstallPluginRequest) error {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return err
	}
	if _, err := s.pool.App.Exec(ctx, `
		INSERT INTO marketplace_installs (tenant_id, plugin_id) VALUES ($1::uuid, $2::uuid) ON CONFLICT DO NOTHING
	`, actor.TenantID, req.PluginID); err != nil {
		return fmt.Errorf("install plugin: %w", err)
	}
	_, _ = s.pool.App.Exec(ctx, `UPDATE marketplace_plugins SET downloads = downloads + 1 WHERE id = $1::uuid`, req.PluginID)
	return nil
}

func (s *MarketplaceService) Uninstall(ctx context.Context, actor ActorContext, pluginID string) error {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return err
	}
	cmd, err := s.pool.App.Exec(ctx, `DELETE FROM marketplace_installs WHERE tenant_id = $1::uuid AND plugin_id = $2::uuid`, actor.TenantID, pluginID)
	if err != nil {
		return fmt.Errorf("uninstall plugin: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrHierarchyNotFound
	}
	return nil
}
