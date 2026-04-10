package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/config"
)

// Pool holds both the application and superadmin database connection pools.
// The App pool enforces RLS policies; the Superadmin pool bypasses them
// and all queries through it must be audit logged (ADR-008).
type Pool struct {
	App        *pgxpool.Pool // Normal app pool (RLS enforced)
	Superadmin *pgxpool.Pool // Superadmin pool (separate role, audit logged)
}

// NewPool creates both database connection pools from config.
func NewPool(ctx context.Context, cfg *config.Config) (*Pool, error) {
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	appPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create app pool: %w", err)
	}

	var superadminPool *pgxpool.Pool
	if cfg.SuperadminDBURL != "" {
		superadminPool, err = pgxpool.New(ctx, cfg.SuperadminDBURL)
		if err != nil {
			appPool.Close()
			return nil, fmt.Errorf("failed to create superadmin pool: %w", err)
		}
	}

	return &Pool{
		App:        appPool,
		Superadmin: superadminPool,
	}, nil
}

// Close closes both connection pools.
func (p *Pool) Close() {
	if p.App != nil {
		p.App.Close()
	}
	if p.Superadmin != nil {
		p.Superadmin.Close()
	}
}

// Ping checks connectivity on the app pool.
func (p *Pool) Ping(ctx context.Context) error {
	return p.App.Ping(ctx)
}
