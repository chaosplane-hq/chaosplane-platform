package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/config"
)

// Querier is the common query surface shared by *pgxpool.Pool and pgx.Tx,
// so call sites can run either against the pool (auto-commit) or against a
// request-scoped transaction carrying the RLS tenant GUC.
type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type txContextKey struct{}

// WithTx stores a request-scoped transaction on the context. The tenant
// middleware sets this so all queries in the request run on the same
// connection where app.current_tenant_id was configured.
func WithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

// Pool holds both the application and superadmin database connection pools.
// The App pool enforces RLS policies; the Superadmin pool bypasses them
// and all queries through it must be audit logged (ADR-008).
type Pool struct {
	App        *pgxpool.Pool // Normal app pool (RLS enforced)
	Superadmin *pgxpool.Pool // Superadmin pool (separate role, audit logged)
}

// Conn returns the request-scoped transaction if the tenant middleware ran,
// otherwise the App pool. RLS-protected tables error when the tenant GUC is
// unset, so non-tenant paths (login, health) fail closed rather than leak.
func (p *Pool) Conn(ctx context.Context) Querier {
	if tx, ok := ctx.Value(txContextKey{}).(pgx.Tx); ok && tx != nil {
		return tx
	}
	return p.App
}

// NewPool creates both database connection pools from config.
func NewPool(ctx context.Context, cfg *config.Config) (*Pool, error) {
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	appPool, err := newConfiguredPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create app pool: %w", err)
	}

	var superadminPool *pgxpool.Pool
	if cfg.SuperadminDBURL != "" {
		superadminPool, err = newConfiguredPool(ctx, cfg.SuperadminDBURL)
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

// newConfiguredPool keeps warm connections (MinConns) so requests never pay the
// multi-second cost of establishing a new TLS connection to RDS Proxy on a cold
// pool, which was making login and other first-hit queries take 10-30s.
func newConfiguredPool(ctx context.Context, url string) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	poolCfg.MinConns = 2
	poolCfg.MaxConns = 10
	poolCfg.MaxConnLifetime = 30 * time.Minute
	poolCfg.MaxConnIdleTime = 5 * time.Minute
	poolCfg.HealthCheckPeriod = 1 * time.Minute

	return pgxpool.NewWithConfig(ctx, poolCfg)
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
