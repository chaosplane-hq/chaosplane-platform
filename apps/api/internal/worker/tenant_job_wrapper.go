package worker

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/config"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

// TenantJobWrapper ensures all async jobs execute with the correct tenant
// context set on the DB connection (ADR-007).
type TenantJobWrapper struct {
	pool *database.Pool
	cfg  *config.Config
}

// NewTenantJobWrapper creates a new TenantJobWrapper.
func NewTenantJobWrapper(pool *database.Pool, cfg *config.Config) *TenantJobWrapper {
	return &TenantJobWrapper{pool: pool, cfg: cfg}
}

// Wrap returns a function that sets SET LOCAL app.current_tenant_id before
// calling the provided handler. Panics in dev mode if tenantID is empty.
func (w *TenantJobWrapper) Wrap(tenantID string, handler func(ctx context.Context) error) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		if tenantID == "" {
			if w.cfg.Environment == "development" {
				panic("TenantJobWrapper: tenantID must not be empty")
			}
			return fmt.Errorf("TenantJobWrapper: tenantID must not be empty")
		}

		conn, err := w.pool.App.Acquire(ctx)
		if err != nil {
			return fmt.Errorf("TenantJobWrapper: failed to acquire connection: %w", err)
		}
		defer conn.Release()

		// ADR-001: SET LOCAL scopes the tenant_id to the current transaction
		_, err = conn.Exec(ctx, "SET LOCAL app.current_tenant_id = $1", tenantID)
		if err != nil {
			return fmt.Errorf("TenantJobWrapper: failed to set tenant context: %w", err)
		}

		slog.Info("tenant job started", "tenant_id", tenantID)
		return handler(ctx)
	}
}
