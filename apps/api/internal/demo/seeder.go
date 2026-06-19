package demo

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

func Seed(ctx context.Context, pool *database.Pool) error {
	db := pool.Superadmin
	if db == nil {
		db = pool.App
	}
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("demo seed: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now()

	if err := seedTenant(ctx, tx); err != nil {
		return err
	}
	if err := seedUser(ctx, tx); err != nil {
		return err
	}
	if err := seedUserTenant(ctx, tx); err != nil {
		return err
	}
	if err := seedOrganization(ctx, tx); err != nil {
		return err
	}
	if err := seedWorkspace(ctx, tx); err != nil {
		return err
	}
	if err := seedTeam(ctx, tx); err != nil {
		return err
	}
	if err := seedTeamMember(ctx, tx); err != nil {
		return err
	}
	if err := seedProject(ctx, tx); err != nil {
		return err
	}
	if err := seedEnvironment(ctx, tx); err != nil {
		return err
	}
	if err := seedExperiments(ctx, tx, now); err != nil {
		return err
	}
	if err := seedTopologySnapshot(ctx, tx, now); err != nil {
		return err
	}
	if err := seedServiceDependencies(ctx, tx, now); err != nil {
		return err
	}
	if err := seedResilienceScore(ctx, tx, now); err != nil {
		return err
	}
	if err := seedVulnerabilityFindings(ctx, tx, now); err != nil {
		return err
	}
	if err := seedExperimentSuggestions(ctx, tx, now); err != nil {
		return err
	}
	if err := seedOnboardingProgress(ctx, tx, now); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func seedTenant(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO tenants (id, name, slug, plan, status)
		VALUES ($1, $2, 'demo', 'business', 'active')
		ON CONFLICT (id) DO NOTHING`,
		TenantID, "ChaosPlane Demo")
	return err
}

func seedUser(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO users (id, email, name, password_hash, status, email_verified)
		VALUES ($1, $2, $3, $4, 'active', true)
		ON CONFLICT (id) DO UPDATE SET password_hash = EXCLUDED.password_hash`,
		UserID, UserEmail, UserName, UserPassword)
	return err
}

func seedUserTenant(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO user_tenants (user_id, tenant_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, tenant_id) DO NOTHING`,
		UserID, TenantID)
	return err
}

func seedOrganization(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO organizations (id, tenant_id, name, slug)
		VALUES ($1, $2, $3, 'demo-org')
		ON CONFLICT (id) DO NOTHING`,
		OrgID, TenantID, OrgName)
	return err
}

func seedWorkspace(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO workspaces (id, tenant_id, organization_id, name, slug)
		VALUES ($1, $2, $3, $4, 'production')
		ON CONFLICT (id) DO NOTHING`,
		WorkspaceID, TenantID, OrgID, WorkspaceName)
	return err
}

func seedTeam(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO teams (id, tenant_id, workspace_id, name, slug)
		VALUES ($1, $2, $3, $4, 'platform-eng')
		ON CONFLICT (id) DO NOTHING`,
		TeamID, TenantID, WorkspaceID, TeamName)
	return err
}

func seedTeamMember(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO team_members (team_id, user_id, role)
		VALUES ($1, $2, 'lead')
		ON CONFLICT (team_id, user_id) DO NOTHING`,
		TeamID, UserID)
	return err
}

func seedProject(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO projects (id, tenant_id, workspace_id, name, slug)
		VALUES ($1, $2, $3, $4, 'ecommerce')
		ON CONFLICT (id) DO NOTHING`,
		ProjectID, TenantID, WorkspaceID, ProjectName)
	return err
}

func seedEnvironment(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO environments (id, tenant_id, project_id, name, slug, type, agent_status, last_heartbeat)
		VALUES ($1, $2, $3, $4, 'production', $5, 'connected', now())
		ON CONFLICT (id) DO NOTHING`,
		EnvironmentID, TenantID, ProjectID, EnvName, EnvType)
	return err
}

func seedExperiments(ctx context.Context, tx pgx.Tx, now time.Time) error {
	type exp struct {
		id     string
		name   string
		action string
		target string
		dur    int
		status string
		start  *time.Time
		end    *time.Time
		errMsg *string
	}

	t := func(d time.Duration) *time.Time { v := now.Add(-d); return &v }
	s := func(v string) *string { return &v }

	experiments := []exp{
		{
			id:     "d0000000-0000-4000-a000-000000000101",
			name:   "pod-kill-payment-svc",
			action: `{"type":"pod-kill","parameters":{"count":1,"grace_period":"0s"}}`,
			target: `{"namespace":"default","mode":"one","labelSelector":"app=payment-svc"}`,
			dur:    300, status: "completed", start: t(2 * time.Hour), end: t(115 * time.Minute),
		},
		{
			id:     "d0000000-0000-4000-a000-000000000102",
			name:   "network-delay-api-gateway",
			action: `{"type":"network-delay","parameters":{"latency":"200ms","jitter":"50ms","interface":"eth0"}}`,
			target: `{"namespace":"default","mode":"all","labelSelector":"app=api-gateway"}`,
			dur:    300, status: "completed", start: t(90 * time.Minute), end: t(85 * time.Minute),
		},
		{
			id:     "d0000000-0000-4000-a000-000000000103",
			name:   "cpu-stress-order-svc",
			action: `{"type":"cpu-stress","parameters":{"workers":2,"load":80}}`,
			target: `{"namespace":"default","mode":"one","labelSelector":"app=order-svc"}`,
			dur:    300, status: "completed", start: t(60 * time.Minute), end: t(55 * time.Minute),
		},
		{
			id:     "d0000000-0000-4000-a000-000000000104",
			name:   "memory-stress-cache",
			action: `{"type":"memory-stress","parameters":{"workers":1,"size":"512Mi"}}`,
			target: `{"namespace":"default","mode":"one","labelSelector":"app=cache-layer"}`,
			dur:    300, status: "failed", start: t(45 * time.Minute), end: t(40 * time.Minute),
			errMsg: s("OOMKilled: container exceeded memory limit"),
		},
		{
			id:     "d0000000-0000-4000-a000-000000000105",
			name:   "dns-error-payment-svc",
			action: `{"type":"dns-error","parameters":{"hostnames":["payment-db.internal"]}}`,
			target: `{"namespace":"default","mode":"all","labelSelector":"app=payment-svc"}`,
			dur:    180, status: "completed", start: t(30 * time.Minute), end: t(25 * time.Minute),
		},
		{
			id:     "d0000000-0000-4000-a000-000000000106",
			name:   "node-drain-worker-1",
			action: `{"type":"node-drain","parameters":{"node":"ip-10-0-1-42","timeout":"60s"}}`,
			target: `{"namespace":"default","mode":"one","labelSelector":"kubernetes.io/hostname=ip-10-0-1-42"}`,
			dur:    600, status: "running", start: t(5 * time.Minute),
		},
		{
			id:     "d0000000-0000-4000-a000-000000000107",
			name:   "network-partition-db",
			action: `{"type":"network-partition","parameters":{"direction":"both","external_targets":["postgres-primary"]}}`,
			target: `{"namespace":"default","mode":"all","labelSelector":"app=order-svc"}`,
			dur:    120, status: "running", start: t(3 * time.Minute),
		},
		{
			id:     "d0000000-0000-4000-a000-000000000108",
			name:   "pod-kill-inventory-svc",
			action: `{"type":"pod-kill","parameters":{"count":2,"grace_period":"30s"}}`,
			target: `{"namespace":"default","mode":"fixed","labelSelector":"app=inventory-svc"}`,
			dur:    300, status: "scheduled", start: nil,
		},
	}

	for _, e := range experiments {
		createdAt := now.Add(-1 * time.Minute)
		if e.start != nil {
			createdAt = *e.start
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO experiments (id, tenant_id, environment_id, name, experiment_type, target, action, duration_seconds, status, run_started_at, run_ended_at, created_by, created_at, updated_at)
			VALUES ($1, $2, $3, $4, 'single', $5::jsonb, $6::jsonb, $7, $8, $9, $10, $11, $12, $12)
			ON CONFLICT (id) DO NOTHING`,
			e.id, TenantID, EnvironmentID, e.name, e.target, e.action, e.dur, e.status, e.start, e.end, UserID, createdAt)
		if err != nil {
			return err
		}

		if e.start != nil {
			resultStatus := e.status
			if resultStatus == "scheduled" {
				continue
			}
			_, err = tx.Exec(ctx, `
				INSERT INTO experiment_results (id, tenant_id, experiment_id, run_number, status, started_at, finished_at, error_message)
				VALUES (gen_random_uuid(), $1, $2, 1, $3, $4, $5, $6)
				ON CONFLICT (experiment_id, run_number) DO NOTHING`,
				TenantID, e.id, resultStatus, e.start, e.end, e.errMsg)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func seedTopologySnapshot(ctx context.Context, tx pgx.Tx, now time.Time) error {
	services := `[
		{"name":"api-gateway","namespace":"default","kind":"Deployment","replicas":3,"ready":3},
		{"name":"auth-svc","namespace":"default","kind":"Deployment","replicas":2,"ready":2},
		{"name":"user-svc","namespace":"default","kind":"Deployment","replicas":2,"ready":2},
		{"name":"order-svc","namespace":"default","kind":"Deployment","replicas":3,"ready":3},
		{"name":"payment-svc","namespace":"default","kind":"Deployment","replicas":2,"ready":2},
		{"name":"notification-svc","namespace":"default","kind":"Deployment","replicas":2,"ready":2},
		{"name":"inventory-svc","namespace":"default","kind":"Deployment","replicas":2,"ready":2},
		{"name":"search-svc","namespace":"default","kind":"Deployment","replicas":2,"ready":2},
		{"name":"analytics-svc","namespace":"default","kind":"Deployment","replicas":2,"ready":2},
		{"name":"cache-layer","namespace":"default","kind":"StatefulSet","replicas":3,"ready":3,"image":"redis:7.2"},
		{"name":"message-queue","namespace":"default","kind":"StatefulSet","replicas":3,"ready":3,"image":"rabbitmq:3.12"},
		{"name":"postgres-primary","namespace":"default","kind":"StatefulSet","replicas":1,"ready":1,"image":"postgres:16"}
	]`

	pods := `[
		{"name":"api-gateway-7f8b9c-x2k4l","namespace":"default","node":"ip-10-0-1-42","status":"Running","ready":true},
		{"name":"api-gateway-7f8b9c-m9n3p","namespace":"default","node":"ip-10-0-1-43","status":"Running","ready":true},
		{"name":"api-gateway-7f8b9c-q5r7s","namespace":"default","node":"ip-10-0-1-44","status":"Running","ready":true},
		{"name":"auth-svc-5d4e6f-a1b2c","namespace":"default","node":"ip-10-0-1-42","status":"Running","ready":true},
		{"name":"auth-svc-5d4e6f-d3e4f","namespace":"default","node":"ip-10-0-1-43","status":"Running","ready":true},
		{"name":"user-svc-8g9h0i-k5l6m","namespace":"default","node":"ip-10-0-1-43","status":"Running","ready":true},
		{"name":"user-svc-8g9h0i-n7o8p","namespace":"default","node":"ip-10-0-1-44","status":"Running","ready":true},
		{"name":"order-svc-2j3k4l-q9r0s","namespace":"default","node":"ip-10-0-1-42","status":"Running","ready":true},
		{"name":"order-svc-2j3k4l-t1u2v","namespace":"default","node":"ip-10-0-1-43","status":"Running","ready":true},
		{"name":"order-svc-2j3k4l-w3x4y","namespace":"default","node":"ip-10-0-1-44","status":"Running","ready":true},
		{"name":"payment-svc-6m7n8o-z5a6b","namespace":"default","node":"ip-10-0-1-42","status":"Running","ready":true},
		{"name":"payment-svc-6m7n8o-c7d8e","namespace":"default","node":"ip-10-0-1-43","status":"Running","ready":true},
		{"name":"notification-svc-1p2q3r-f9g0h","namespace":"default","node":"ip-10-0-1-44","status":"Running","ready":true},
		{"name":"notification-svc-1p2q3r-i1j2k","namespace":"default","node":"ip-10-0-1-42","status":"Running","ready":true},
		{"name":"inventory-svc-4s5t6u-l3m4n","namespace":"default","node":"ip-10-0-1-43","status":"Running","ready":true},
		{"name":"inventory-svc-4s5t6u-o5p6q","namespace":"default","node":"ip-10-0-1-44","status":"Running","ready":true},
		{"name":"search-svc-7v8w9x-r7s8t","namespace":"default","node":"ip-10-0-1-42","status":"Running","ready":true},
		{"name":"search-svc-7v8w9x-u9v0w","namespace":"default","node":"ip-10-0-1-43","status":"Running","ready":true},
		{"name":"analytics-svc-0y1z2a-x1y2z","namespace":"default","node":"ip-10-0-1-44","status":"Running","ready":true},
		{"name":"analytics-svc-0y1z2a-a3b4c","namespace":"default","node":"ip-10-0-1-42","status":"Running","ready":true},
		{"name":"cache-layer-0","namespace":"default","node":"ip-10-0-1-42","status":"Running","ready":true},
		{"name":"cache-layer-1","namespace":"default","node":"ip-10-0-1-43","status":"Running","ready":true},
		{"name":"cache-layer-2","namespace":"default","node":"ip-10-0-1-44","status":"Running","ready":true},
		{"name":"message-queue-0","namespace":"default","node":"ip-10-0-1-42","status":"Running","ready":true},
		{"name":"message-queue-1","namespace":"default","node":"ip-10-0-1-43","status":"Running","ready":true},
		{"name":"message-queue-2","namespace":"default","node":"ip-10-0-1-44","status":"Running","ready":true},
		{"name":"postgres-primary-0","namespace":"default","node":"ip-10-0-1-42","status":"Running","ready":true}
	]`

	deployments := `[
		{"name":"api-gateway","namespace":"default","replicas":3,"ready":3,"strategy":"RollingUpdate"},
		{"name":"auth-svc","namespace":"default","replicas":2,"ready":2,"strategy":"RollingUpdate"},
		{"name":"user-svc","namespace":"default","replicas":2,"ready":2,"strategy":"RollingUpdate"},
		{"name":"order-svc","namespace":"default","replicas":3,"ready":3,"strategy":"RollingUpdate"},
		{"name":"payment-svc","namespace":"default","replicas":2,"ready":2,"strategy":"RollingUpdate"},
		{"name":"notification-svc","namespace":"default","replicas":2,"ready":2,"strategy":"RollingUpdate"},
		{"name":"inventory-svc","namespace":"default","replicas":2,"ready":2,"strategy":"RollingUpdate"},
		{"name":"search-svc","namespace":"default","replicas":2,"ready":2,"strategy":"RollingUpdate"},
		{"name":"analytics-svc","namespace":"default","replicas":2,"ready":2,"strategy":"RollingUpdate"}
	]`

	nodes := `[
		{"name":"ip-10-0-1-42","status":"Ready","roles":["worker"],"capacity":{"cpu":"8","memory":"32Gi","pods":"110"},"allocatable":{"cpu":"7800m","memory":"30Gi","pods":"110"}},
		{"name":"ip-10-0-1-43","status":"Ready","roles":["worker"],"capacity":{"cpu":"8","memory":"32Gi","pods":"110"},"allocatable":{"cpu":"7800m","memory":"30Gi","pods":"110"}},
		{"name":"ip-10-0-1-44","status":"Ready","roles":["worker"],"capacity":{"cpu":"8","memory":"32Gi","pods":"110"},"allocatable":{"cpu":"7800m","memory":"30Gi","pods":"110"}}
	]`

	namespaces := `["default","monitoring","chaosplane-system"]`

	_, err := tx.Exec(ctx, `
		INSERT INTO topology_snapshots (id, tenant_id, environment_id, services, pods, deployments, nodes, namespaces, collected_at, created_at)
		VALUES ('d0000000-0000-4000-a000-000000000201', $1, $2, $3::jsonb, $4::jsonb, $5::jsonb, $6::jsonb, $7::jsonb, $8, $8)
		ON CONFLICT (id) DO NOTHING`,
		TenantID, EnvironmentID, services, pods, deployments, nodes, namespaces, now)
	return err
}

func seedServiceDependencies(ctx context.Context, tx pgx.Tx, now time.Time) error {
	type dep struct {
		srcKind, srcName, srcNs string
		tgtKind, tgtName, tgtNs string
		protocol                string
		port                    int
	}

	deps := []dep{
		{"Deployment", "api-gateway", "default", "Deployment", "auth-svc", "default", "HTTP", 8080},
		{"Deployment", "api-gateway", "default", "Deployment", "user-svc", "default", "HTTP", 8080},
		{"Deployment", "api-gateway", "default", "Deployment", "order-svc", "default", "HTTP", 8080},
		{"Deployment", "api-gateway", "default", "Deployment", "payment-svc", "default", "HTTP", 8080},
		{"Deployment", "api-gateway", "default", "Deployment", "inventory-svc", "default", "HTTP", 8080},
		{"Deployment", "api-gateway", "default", "Deployment", "search-svc", "default", "HTTP", 8080},
		{"Deployment", "order-svc", "default", "Deployment", "payment-svc", "default", "gRPC", 9090},
		{"Deployment", "order-svc", "default", "Deployment", "inventory-svc", "default", "gRPC", 9090},
		{"Deployment", "order-svc", "default", "StatefulSet", "message-queue", "default", "AMQP", 5672},
		{"Deployment", "payment-svc", "default", "StatefulSet", "postgres-primary", "default", "TCP", 5432},
		{"Deployment", "user-svc", "default", "StatefulSet", "postgres-primary", "default", "TCP", 5432},
		{"Deployment", "order-svc", "default", "StatefulSet", "postgres-primary", "default", "TCP", 5432},
		{"Deployment", "inventory-svc", "default", "StatefulSet", "postgres-primary", "default", "TCP", 5432},
		{"Deployment", "auth-svc", "default", "StatefulSet", "cache-layer", "default", "TCP", 6379},
		{"Deployment", "api-gateway", "default", "StatefulSet", "cache-layer", "default", "TCP", 6379},
		{"Deployment", "search-svc", "default", "StatefulSet", "cache-layer", "default", "TCP", 6379},
		{"Deployment", "notification-svc", "default", "StatefulSet", "message-queue", "default", "AMQP", 5672},
		{"Deployment", "analytics-svc", "default", "StatefulSet", "message-queue", "default", "AMQP", 5672},
		{"Deployment", "analytics-svc", "default", "StatefulSet", "postgres-primary", "default", "TCP", 5432},
	}

	for _, d := range deps {
		_, err := tx.Exec(ctx, `
			INSERT INTO service_dependencies (tenant_id, environment_id, source_kind, source_name, source_namespace, target_kind, target_name, target_namespace, protocol, port, discovered_at, last_seen_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11)
			ON CONFLICT (environment_id, source_kind, source_name, source_namespace, target_kind, target_name, target_namespace) DO UPDATE SET last_seen_at = EXCLUDED.last_seen_at`,
			TenantID, EnvironmentID, d.srcKind, d.srcName, d.srcNs, d.tgtKind, d.tgtName, d.tgtNs, d.protocol, d.port, now)
		if err != nil {
			return err
		}
	}
	return nil
}

func seedResilienceScore(ctx context.Context, tx pgx.Tx, now time.Time) error {
	details := `{"observability":75,"categories":{"pod_resilience":82,"network_resilience":70,"node_resilience":78,"data_resilience":80}}`
	_, err := tx.Exec(ctx, `
		INSERT INTO resilience_scores (id, tenant_id, environment_id, overall_grade, overall_score, availability, fault_tolerance, recoverability, details, calculated_at)
		VALUES ('d0000000-0000-4000-a000-000000000301', $1, $2, 'B', 78, 85, 72, 80, $3::jsonb, $4)
		ON CONFLICT (id) DO NOTHING`,
		TenantID, EnvironmentID, details, now)
	return err
}

func seedVulnerabilityFindings(ctx context.Context, tx pgx.Tx, now time.Time) error {
	type finding struct {
		id       string
		category string
		severity string
		title    string
		desc     string
		kind     string
		name     string
		remed    string
	}

	findings := []finding{
		{
			id: "d0000000-0000-4000-a000-000000000401", category: "spof", severity: "critical",
			title: "Single replica for payment-svc", desc: "payment-svc has no replica redundancy in production. A single pod failure will cause downtime.",
			kind: "Deployment", name: "payment-svc", remed: "Scale payment-svc to at least 2 replicas and add a PodDisruptionBudget.",
		},
		{
			id: "d0000000-0000-4000-a000-000000000402", category: "healthcheck", severity: "high",
			title: "Missing liveness probe on analytics-svc", desc: "analytics-svc containers have no liveness probe configured. Hung processes will not be restarted.",
			kind: "Deployment", name: "analytics-svc", remed: "Add a liveness probe with appropriate thresholds to the analytics-svc deployment.",
		},
		{
			id: "d0000000-0000-4000-a000-000000000403", category: "resource_limits", severity: "high",
			title: "No memory limits on search-svc", desc: "search-svc has no memory limits set. Unbounded memory usage can trigger node-level OOM.",
			kind: "Deployment", name: "search-svc", remed: "Set memory requests and limits on all search-svc containers.",
		},
		{
			id: "d0000000-0000-4000-a000-000000000404", category: "pdb", severity: "medium",
			title: "No PodDisruptionBudget for order-svc", desc: "order-svc lacks a PDB. Voluntary disruptions (node drain, upgrades) could evict all pods simultaneously.",
			kind: "Deployment", name: "order-svc", remed: "Create a PodDisruptionBudget with minAvailable of at least 1 for order-svc.",
		},
		{
			id: "d0000000-0000-4000-a000-000000000405", category: "networking", severity: "medium",
			title: "No NetworkPolicy for cache-layer", desc: "cache-layer (Redis) accepts connections from any pod in the namespace. Lateral movement risk.",
			kind: "StatefulSet", name: "cache-layer", remed: "Create a NetworkPolicy restricting ingress to only services that require cache access.",
		},
	}

	for _, f := range findings {
		_, err := tx.Exec(ctx, `
			INSERT INTO vulnerability_findings (id, tenant_id, environment_id, category, severity, title, description, resource_kind, resource_name, resource_namespace, remediation, status, detected_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'default', $10, 'open', $11)
			ON CONFLICT (id) DO NOTHING`,
			f.id, TenantID, EnvironmentID, f.category, f.severity, f.title, f.desc, f.kind, f.name, f.remed, now.Add(-24*time.Hour))
		if err != nil {
			return err
		}
	}
	return nil
}

func seedExperimentSuggestions(ctx context.Context, tx pgx.Tx, now time.Time) error {
	type suggestion struct {
		id        string
		findingID *string
		source    string
		title     string
		desc      string
		action    string
		ns        string
		target    string
		duration  string
	}

	fid := func(v string) *string { return &v }

	suggestions := []suggestion{
		{
			id: "d0000000-0000-4000-a000-000000000501", findingID: fid("d0000000-0000-4000-a000-000000000401"),
			source: "vulnerability", title: "Kill payment-svc pod to verify SPOF",
			desc:   "Terminate the single payment-svc pod to confirm the single-point-of-failure finding and measure recovery time.",
			action: "pod-kill", ns: "default", target: "payment-svc", duration: "30s",
		},
		{
			id: "d0000000-0000-4000-a000-000000000502", findingID: fid("d0000000-0000-4000-a000-000000000404"),
			source: "vulnerability", title: "Drain node hosting order-svc",
			desc:   "Drain a node to test order-svc behavior without a PodDisruptionBudget under voluntary disruption.",
			action: "node-drain", ns: "default", target: "order-svc", duration: "120s",
		},
		{
			id: "d0000000-0000-4000-a000-000000000503", findingID: nil,
			source: "best_practice", title: "Network partition between order-svc and postgres",
			desc:   "Simulate network partition between order-svc and the database to validate circuit breaker and retry logic.",
			action: "network-partition", ns: "default", target: "order-svc", duration: "60s",
		},
	}

	for _, s := range suggestions {
		_, err := tx.Exec(ctx, `
			INSERT INTO experiment_suggestions (id, tenant_id, environment_id, finding_id, source, title, description, action_type, target_namespace, target_name, duration, confidence, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 0.85, $12)
			ON CONFLICT (id) DO NOTHING`,
			s.id, TenantID, EnvironmentID, s.findingID, s.source, s.title, s.desc, s.action, s.ns, s.target, s.duration, now)
		if err != nil {
			return err
		}
	}
	return nil
}

func seedOnboardingProgress(ctx context.Context, tx pgx.Tx, now time.Time) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO onboarding_progress (id, user_id, tenant_id, step_org_created, step_workspace_created, step_team_created, step_member_invited, step_cluster_connected, step_first_experiment, step_result_viewed, completed_at, created_at, updated_at)
		VALUES ('d0000000-0000-4000-a000-000000000601', $1, $2, true, true, true, true, true, true, true, $3, $3, $3)
		ON CONFLICT (user_id, tenant_id) DO NOTHING`,
		UserID, TenantID, now)
	return err
}
