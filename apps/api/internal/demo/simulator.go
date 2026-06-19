package demo

import (
	"context"
	"log/slog"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/config"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

type Simulator struct {
	pool     *database.Pool
	inFlight sync.Map
}

func NewSimulator(pool *database.Pool) *Simulator {
	return &Simulator{pool: pool}
}

func (s *Simulator) db() *pgxpool.Pool {
	if s.pool.Superadmin != nil {
		return s.pool.Superadmin
	}
	return s.pool.App
}

func (s *Simulator) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.poll(ctx)
			}
		}
	}()
}

func (s *Simulator) poll(ctx context.Context) {
	rows, err := s.db().Query(ctx,
		`SELECT id::text, name FROM experiments WHERE tenant_id = $1 AND status = 'scheduled'`,
		TenantID,
	)
	if err != nil {
		slog.Error("demo: failed to poll scheduled experiments", "error", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			slog.Error("demo: failed to scan experiment row", "error", err)
			continue
		}
		if _, loaded := s.inFlight.LoadOrStore(id, true); loaded {
			continue
		}
		go s.simulateExperiment(ctx, id, name)
	}
}

func (s *Simulator) simulateExperiment(ctx context.Context, id, name string) {
	defer s.inFlight.Delete(id)

	select {
	case <-time.After(5 * time.Second):
	case <-ctx.Done():
		return
	}

	_, err := s.db().Exec(ctx,
		`UPDATE experiments SET status = 'running', run_started_at = now(), updated_at = now() WHERE id = $1`,
		id,
	)
	if err != nil {
		slog.Error("demo: failed to transition to running", "id", id, "error", err)
		return
	}
	slog.Info("demo: experiment transitioned", "id", id, "name", name, "status", "running")

	delay := 25*time.Second + time.Duration(rand.Int63n(int64(20*time.Second)))
	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return
	}

	_, err = s.db().Exec(ctx,
		`UPDATE experiments SET status = 'completed', run_ended_at = now(), updated_at = now() WHERE id = $1`,
		id,
	)
	if err != nil {
		slog.Error("demo: failed to transition to completed", "id", id, "error", err)
		return
	}
	slog.Info("demo: experiment transitioned", "id", id, "name", name, "status", "completed")

	_, err = s.db().Exec(ctx,
		`INSERT INTO experiment_results (experiment_id, tenant_id, steady_state_met, impact_summary) VALUES ($1, $2, true, 'Experiment completed successfully. No service degradation detected.')`,
		id, TenantID,
	)
	if err != nil {
		slog.Error("demo: failed to insert experiment result", "id", id, "error", err)
	}
}

func TestConnectionHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Demo {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"connected":    true,
			"agentVersion": "1.2.0-demo",
			"message":      "Demo cluster connected",
		})
	}
}
