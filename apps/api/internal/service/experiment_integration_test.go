package service

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

// These integration tests run only when TEST_DATABASE_URL points at a database
// with the full migration set applied; they exercise the real RLS tenant tx so
// persistence and agent claiming are covered end-to-end, not just in unit form.
func integrationPool(t *testing.T) *database.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping DB integration test")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("connect test db: %v", err)
	}
	t.Cleanup(pool.Close)
	return &database.Pool{App: pool}
}

// seedTenant creates the minimal tenant/user/workspace graph an experiment
// needs, returning an ActorContext. Rows are cleaned up after the test.
func seedTenant(t *testing.T, pool *database.Pool) ActorContext {
	t.Helper()
	ctx := context.Background()
	tenantID := uuid.NewString()
	userID := uuid.NewString()

	conn := pool.App
	if _, err := conn.Exec(ctx, `INSERT INTO tenants (id, name, slug) VALUES ($1::uuid, $2, $3)`,
		tenantID, "itest-"+tenantID[:8], "itest-"+tenantID[:8]); err != nil {
		t.Fatalf("seed tenant: %v", err)
	}
	if _, err := conn.Exec(ctx, `INSERT INTO users (id, email, password_hash, name) VALUES ($1::uuid, $2, $3, $4)`,
		userID, userID+"@itest.local", "x", "itest"); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if _, err := conn.Exec(ctx, `INSERT INTO user_tenants (user_id, tenant_id) VALUES ($1::uuid, $2::uuid)`,
		userID, tenantID); err != nil {
		t.Fatalf("seed membership: %v", err)
	}
	var orgID string
	if err := conn.QueryRow(ctx, `INSERT INTO organizations (tenant_id, name, slug) VALUES ($1::uuid, 'Org', 'org') RETURNING id::text`,
		tenantID).Scan(&orgID); err != nil {
		t.Fatalf("seed organization: %v", err)
	}
	if _, err := conn.Exec(ctx, `INSERT INTO workspaces (tenant_id, organization_id, name, slug) VALUES ($1::uuid, $2::uuid, 'WS', 'ws')`,
		tenantID, orgID); err != nil {
		t.Fatalf("seed workspace: %v", err)
	}

	t.Cleanup(func() {
		_, _ = conn.Exec(ctx, `DELETE FROM tenants WHERE id = $1::uuid`, tenantID)
		_, _ = conn.Exec(ctx, `DELETE FROM users WHERE id = $1::uuid`, userID)
	})

	return ActorContext{UserID: userID, TenantID: tenantID}
}

// withTenantTx mimics the TenantContext middleware: a transaction-scoped GUC so
// RLS sees the tenant. Each call commits, mirroring one request.
func withTenantTx(t *testing.T, pool *database.Pool, tenantID string, fn func(ctx context.Context)) {
	t.Helper()
	ctx := context.Background()
	tx, err := pool.App.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID); err != nil {
		t.Fatalf("set tenant guc: %v", err)
	}
	fn(database.WithTx(ctx, tx))
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit tx: %v", err)
	}
}

func TestIntegrationCreateSingleActionPersists(t *testing.T) {
	pool := integrationPool(t)
	actor := seedTenant(t, pool)
	svc := NewExperimentService(pool)

	var created *ExperimentResponse
	withTenantTx(t, pool, actor.TenantID, func(ctx context.Context) {
		var err error
		created, err = svc.Create(ctx, actor, &CreateExperimentRequest{
			Name:      "single-itest",
			Namespace: "default",
			Action:    ActionRequest{Type: "pod-kill"},
			Target:    TargetRequest{Namespace: "default"},
			Duration:  "30s",
		})
		if err != nil {
			t.Fatalf("create single: %v", err)
		}
	})

	if created.ExperimentType != "single" {
		t.Fatalf("expected experimentType=single, got %q", created.ExperimentType)
	}
	if created.Action.Type != "pod-kill" {
		t.Fatalf("expected action pod-kill, got %q", created.Action.Type)
	}

	withTenantTx(t, pool, actor.TenantID, func(ctx context.Context) {
		got, err := svc.Get(ctx, actor, created.ID)
		if err != nil {
			t.Fatalf("get single: %v", err)
		}
		if got.Action.Type != "pod-kill" || len(got.Steps) != 0 {
			t.Fatalf("single round-trip mismatch: %+v", got)
		}
	})
}

func TestIntegrationCreateWorkflowPersistsAndClaims(t *testing.T) {
	pool := integrationPool(t)
	actor := seedTenant(t, pool)
	svc := NewExperimentService(pool)
	agentSvc := NewAgentWorkService(pool)

	steps := []ScenarioStep{
		{Name: "kill", Action: ActionRequest{Type: "pod-kill"}, Target: TargetRequest{Namespace: "default"}, Duration: "30s"},
		{Name: "delay", DependsOn: []string{"kill"}, Action: ActionRequest{Type: "network-delay"}, Target: TargetRequest{Namespace: "default"}, Duration: "1m"},
	}

	var created *ExperimentResponse
	withTenantTx(t, pool, actor.TenantID, func(ctx context.Context) {
		var err error
		created, err = svc.Create(ctx, actor, &CreateExperimentRequest{
			Name:      "wf-itest",
			Namespace: "default",
			Steps:     steps,
			Duration:  "2m",
		})
		if err != nil {
			t.Fatalf("create workflow: %v", err)
		}
	})

	if created.ExperimentType != "workflow" {
		t.Fatalf("expected experimentType=workflow, got %q", created.ExperimentType)
	}
	if len(created.Steps) != 2 {
		t.Fatalf("expected 2 persisted steps, got %d", len(created.Steps))
	}

	var envID string
	withTenantTx(t, pool, actor.TenantID, func(ctx context.Context) {
		if err := pool.Conn(ctx).QueryRow(ctx,
			`SELECT environment_id::text FROM experiments WHERE id = $1::uuid`, created.ID).Scan(&envID); err != nil {
			t.Fatalf("lookup env: %v", err)
		}
	})

	var work *AgentWorkItem
	withTenantTx(t, pool, actor.TenantID, func(ctx context.Context) {
		var err error
		work, err = agentSvc.ClaimWork(ctx, envID, "itest-agent")
		if err != nil {
			t.Fatalf("claim work: %v", err)
		}
	})

	if work.ID != created.ID {
		t.Fatalf("claimed wrong experiment: %s vs %s", work.ID, created.ID)
	}
	if work.ExperimentType != "workflow" {
		t.Fatalf("claimed item not a workflow: %q", work.ExperimentType)
	}
	var claimedSteps []ScenarioStep
	if err := json.Unmarshal(work.Steps, &claimedSteps); err != nil {
		t.Fatalf("unmarshal claimed steps: %v (raw=%s)", err, string(work.Steps))
	}
	if len(claimedSteps) != 2 || claimedSteps[1].Name != "delay" || claimedSteps[1].DependsOn[0] != "kill" {
		t.Fatalf("claimed steps lost DAG structure: %+v", claimedSteps)
	}
}

func TestIntegrationCreateUnknownFaultRejectedNoRow(t *testing.T) {
	pool := integrationPool(t)
	actor := seedTenant(t, pool)
	svc := NewExperimentService(pool)

	withTenantTx(t, pool, actor.TenantID, func(ctx context.Context) {
		_, err := svc.Create(ctx, actor, &CreateExperimentRequest{
			Name:   "bad-itest",
			Action: ActionRequest{Type: "made-up"},
			Target: TargetRequest{Namespace: "default"},
		})
		var verr *ValidationError
		if !asValidationError(err, &verr) {
			t.Fatalf("expected *ValidationError, got %v", err)
		}
	})

	withTenantTx(t, pool, actor.TenantID, func(ctx context.Context) {
		var count int
		if err := pool.Conn(ctx).QueryRow(ctx,
			`SELECT count(*) FROM experiments WHERE tenant_id = $1::uuid AND name = 'bad-itest'`, actor.TenantID).Scan(&count); err != nil {
			if err == pgx.ErrNoRows {
				return
			}
			t.Fatalf("count: %v", err)
		}
		if count != 0 {
			t.Fatalf("rejected experiment must not persist, found %d rows", count)
		}
	})
}
