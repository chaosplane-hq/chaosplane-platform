package demo

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

type Refresher struct {
	pool *database.Pool
}

func NewRefresher(pool *database.Pool) *Refresher {
	return &Refresher{pool: pool}
}

func (r *Refresher) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(4 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.refresh(ctx)
			}
		}
	}()
}

func (r *Refresher) db() *pgxpool.Pool {
	if r.pool.Superadmin != nil {
		return r.pool.Superadmin
	}
	return r.pool.App
}

func (r *Refresher) refresh(ctx context.Context) {
	queries := []string{
		`UPDATE experiments SET created_at = created_at + (now() - updated_at), updated_at = now() WHERE tenant_id = $1`,
		`UPDATE topology_snapshots SET collected_at = now() WHERE tenant_id = $1`,
		`UPDATE vulnerability_findings SET detected_at = now() - (random() * interval '24 hours') WHERE tenant_id = $1`,
	}

	for _, q := range queries {
		if _, err := r.db().Exec(ctx, q, TenantID); err != nil {
			slog.Error("demo refresher failed", "error", err, "query", q)
		}
	}
}
