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

	if _, err := tx.Exec(ctx, "SET LOCAL app.current_tenant_id = '"+TenantID+"'"); err != nil {
		return fmt.Errorf("demo seed: set tenant context: %w", err)
	}

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
	if err := seedResultAnalyses(ctx, tx, now); err != nil {
		return err
	}
	if err := seedResilienceTrend(ctx, tx, now); err != nil {
		return err
	}
	if err := seedFederatedClusters(ctx, tx); err != nil {
		return err
	}
	if err := seedSubscription(ctx, tx); err != nil {
		return err
	}
	if err := seedGamedays(ctx, tx, now); err != nil {
		return err
	}
	if err := seedWorkflowTemplates(ctx, tx, now); err != nil {
		return err
	}
	if err := seedPolicies(ctx, tx, now); err != nil {
		return err
	}
	if err := seedNotificationChannels(ctx, tx, now); err != nil {
		return err
	}
	if err := seedPredictions(ctx, tx, now); err != nil {
		return err
	}
	if err := seedMarketplace(ctx, tx, now); err != nil {
		return err
	}
	if err := seedCICDIntegrations(ctx, tx, now); err != nil {
		return err
	}
	if err := seedTopologyDrifts(ctx, tx, now); err != nil {
		return err
	}
	if err := seedTopologyMetrics(ctx, tx, now); err != nil {
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

func seedResultAnalyses(ctx context.Context, tx pgx.Tx, now time.Time) error {
	type analysis struct {
		id, name, summary, impact, recs, severity string
		services                                  string
	}
	items := []analysis{
		{"d0000000-0000-4000-a000-000000000701", "pod-kill-payment-svc",
			"Payment service recovered within 12 seconds after pod termination. Kubernetes recreated the pod and readiness probe passed quickly.",
			"Brief 502 errors observed on payment endpoints during the 12s recovery window. 3 in-flight transactions were retried successfully by the client.",
			"Add a PodDisruptionBudget with minAvailable=1. Consider running at least 2 replicas to eliminate single-point-of-failure.",
			"medium", `["payment-svc","api-gateway"]`},
		{"d0000000-0000-4000-a000-000000000702", "network-delay-api-gateway",
			"API gateway handled 200ms latency injection gracefully. Circuit breakers activated at 150ms threshold for downstream calls.",
			"P99 latency increased from 45ms to 260ms during the experiment. No errors or timeouts observed. Circuit breaker prevented cascade.",
			"Current timeout settings are well-tuned. Consider adding retry budgets to prevent amplification under sustained latency.",
			"low", `["api-gateway","order-svc","user-svc"]`},
		{"d0000000-0000-4000-a000-000000000703", "cpu-stress-order-svc",
			"Order service maintained functionality under 80% CPU stress. Response times degraded by 3x but no requests failed.",
			"P50 latency went from 20ms to 65ms. CPU throttling engaged as expected. HPA did not scale because threshold is set at 90%.",
			"Lower HPA CPU threshold to 70% for faster scale-out. Add CPU requests to prevent co-located pod starvation.",
			"low", `["order-svc"]`},
		{"d0000000-0000-4000-a000-000000000704", "memory-stress-cache",
			"Cache layer pod was OOMKilled after 40 seconds of memory stress. Redis data was lost and required cold cache rebuild.",
			"Cache miss rate spiked to 100% for 45 seconds during rebuild. Downstream services experienced 4x latency increase hitting the database directly.",
			"Increase memory limits for cache-layer. Implement Redis persistence (AOF/RDB) to survive restarts. Add eviction policies.",
			"high", `["cache-layer","order-svc","payment-svc","search-svc"]`},
		{"d0000000-0000-4000-a000-000000000705", "dns-error-payment-svc",
			"Payment service correctly fell back to cached DNS entries for 15 seconds, then began failing when cache expired.",
			"After DNS cache expiry at T+15s, payment processing failed for 10 seconds until DNS recovered. 23 transactions queued and retried.",
			"Increase DNS cache TTL in payment-svc. Implement circuit breaker for DNS-dependent external calls. Add fallback IP resolution.",
			"medium", `["payment-svc"]`},
	}

	for _, a := range items {
		_, err := tx.Exec(ctx, `
			INSERT INTO experiment_results_analysis (id, tenant_id, experiment_name, environment_id, summary, impact_analysis, recommendations, severity_assessment, affected_services, metrics_impact, analyzed_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, '{}'::jsonb, $10)
			ON CONFLICT (id) DO NOTHING`,
			a.id, TenantID, a.name, EnvironmentID, a.summary, a.impact, a.recs, a.severity, a.services, now.Add(-time.Duration(len(items))*time.Hour))
		if err != nil {
			return err
		}
	}
	return nil
}

func seedResilienceTrend(ctx context.Context, tx pgx.Tx, now time.Time) error {
	type score struct {
		id                      string
		grade                   string
		overall, avail, ft, rec float64
		daysAgo                 int
	}
	scores := []score{
		{"d0000000-0000-4000-a000-000000000302", "C", 72, 78, 68, 72, 4},
		{"d0000000-0000-4000-a000-000000000303", "C", 74, 80, 70, 74, 3},
		{"d0000000-0000-4000-a000-000000000304", "B", 76, 82, 71, 77, 2},
		{"d0000000-0000-4000-a000-000000000305", "B", 77, 84, 72, 79, 1},
	}
	for _, s := range scores {
		_, err := tx.Exec(ctx, `
			INSERT INTO resilience_scores (id, tenant_id, environment_id, overall_grade, overall_score, availability, fault_tolerance, recoverability, details, calculated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, '{}'::jsonb, $9)
			ON CONFLICT (id) DO NOTHING`,
			s.id, TenantID, EnvironmentID, s.grade, s.overall, s.avail, s.ft, s.rec, now.Add(-time.Duration(s.daysAgo)*24*time.Hour))
		if err != nil {
			return err
		}
	}
	return nil
}

func seedFederatedClusters(ctx context.Context, tx pgx.Tx) error {
	type cluster struct {
		name, region, endpoint, status string
		agentID                        *string
	}
	envID := EnvironmentID
	clusters := []cluster{
		{"prod-ap-northeast-2", "ap-northeast-2", "https://eks.ap-northeast-2.amazonaws.com/chaosplane-prod", "connected", &envID},
		{"staging-us-west-2", "us-west-2", "https://eks.us-west-2.amazonaws.com/chaosplane-staging", "connected", nil},
		{"dev-eu-west-1", "eu-west-1", "https://eks.eu-west-1.amazonaws.com/chaosplane-dev", "disconnected", nil},
	}
	for _, c := range clusters {
		_, err := tx.Exec(ctx, `
			INSERT INTO federation_clusters (tenant_id, name, region, provider, api_endpoint, status, agent_id, metadata)
			VALUES ($1, $2, $3, 'aws', $4, $5, $6, '{}'::jsonb)
			ON CONFLICT (tenant_id, name) DO NOTHING`,
			TenantID, c.name, c.region, c.endpoint, c.status, c.agentID)
		if err != nil {
			return err
		}
	}
	return nil
}

func seedSubscription(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO subscriptions (tenant_id, plan, status, gateway, current_period_start, current_period_end, trial_ends_at)
		VALUES ($1, 'business', 'active', 'none', now() - interval '25 days', now() + interval '5 days', now() - interval '14 days')
		ON CONFLICT (tenant_id) DO UPDATE SET plan = 'business', status = 'active'`,
		TenantID)
	return err
}

func seedGamedays(ctx context.Context, tx pgx.Tx, now time.Time) error {
	type gd struct {
		id, title, desc, status   string
		scheduled, started, ended *time.Time
	}
	t := func(d time.Duration) *time.Time { v := now.Add(d); return &v }
	gamedays := []gd{
		{"d0000000-0000-4000-a000-000000000801", "Q2 Resilience Drill", "Quarterly resilience exercise targeting payment and order processing services", "completed", nil, t(-72 * time.Hour), t(-68 * time.Hour)},
		{"d0000000-0000-4000-a000-000000000802", "Network Partition Drill", "Test service mesh behavior during inter-AZ network partition", "scheduled", t(24 * time.Hour), nil, nil},
		{"d0000000-0000-4000-a000-000000000803", "Black Friday Prep", "Load testing and chaos validation before peak traffic season", "planning", nil, nil, nil},
	}
	for _, g := range gamedays {
		_, err := tx.Exec(ctx, `
			INSERT INTO gamedays (id, tenant_id, environment_id, title, description, status, scheduled_at, started_at, ended_at, created_by, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (id) DO NOTHING`,
			g.id, TenantID, EnvironmentID, g.title, g.desc, g.status, g.scheduled, g.started, g.ended, UserID, now.Add(-7*24*time.Hour))
		if err != nil {
			return err
		}
	}
	return nil
}

func seedWorkflowTemplates(ctx context.Context, tx pgx.Tx, now time.Time) error {
	type tmpl struct {
		id, name, desc, spec string
	}
	templates := []tmpl{
		{"d0000000-0000-4000-a000-000000000901", "Sequential Pod Kill", "Terminates pods one service at a time to validate independent recovery",
			`{"steps":[{"name":"kill-payment","type":"pod-kill","target":"payment-svc","duration":"30s"},{"name":"kill-order","type":"pod-kill","target":"order-svc","duration":"30s","dependsOn":["kill-payment"]}]}`},
		{"d0000000-0000-4000-a000-000000000902", "Network Chaos Suite", "Parallel network faults across multiple services",
			`{"steps":[{"name":"delay-gateway","type":"network-delay","target":"api-gateway","duration":"60s"},{"name":"loss-order","type":"network-loss","target":"order-svc","duration":"60s"},{"name":"partition-db","type":"network-partition","target":"postgres-primary","duration":"30s"}]}`},
		{"d0000000-0000-4000-a000-000000000903", "Full Stack Resilience", "Combined pod, network, and resource chaos for comprehensive testing",
			`{"steps":[{"name":"stress-cpu","type":"cpu-stress","target":"order-svc","duration":"120s"},{"name":"kill-cache","type":"pod-kill","target":"cache-layer","duration":"30s","dependsOn":["stress-cpu"]},{"name":"delay-all","type":"network-delay","target":"*","duration":"60s","dependsOn":["kill-cache"]}]}`},
	}
	for _, t := range templates {
		_, err := tx.Exec(ctx, `
			INSERT INTO workflow_templates (id, tenant_id, name, description, category, is_public, spec, created_by, created_at)
			VALUES ($1, $2, $3, $4, 'custom', false, $5::jsonb, $6, $7)
			ON CONFLICT (id) DO NOTHING`,
			t.id, TenantID, t.name, t.desc, t.spec, UserID, now.Add(-14*24*time.Hour))
		if err != nil {
			return err
		}
	}
	return nil
}

func seedPolicies(ctx context.Context, tx pgx.Tx, now time.Time) error {
	type pol struct {
		id, name, desc, enforcement string
		maxConc, maxTgt             int
		blocked                     string
	}
	policies := []pol{
		{"d0000000-0000-4000-a000-000000000a01", "Production Safety", "Limits concurrent experiments and protects critical system namespaces", "enforce", 2, 3, "{kube-system,chaosplane-system}"},
		{"d0000000-0000-4000-a000-000000000a02", "Business Hours Only", "Audit-mode policy for tracking off-hours experiments", "audit", 5, 10, "{}"},
		{"d0000000-0000-4000-a000-000000000a03", "Critical Service Protection", "Strict limits for payment-related services", "enforce", 1, 1, "{payment-system}"},
	}
	for _, p := range policies {
		_, err := tx.Exec(ctx, `
			INSERT INTO blast_radius_policies (id, tenant_id, environment_id, name, description, enforcement, max_concurrent, max_targets, allowed_namespaces, blocked_namespaces, created_by, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, '{}', $9, $10, $11)
			ON CONFLICT (id) DO NOTHING`,
			p.id, TenantID, EnvironmentID, p.name, p.desc, p.enforcement, p.maxConc, p.maxTgt, p.blocked, UserID, now.Add(-30*24*time.Hour))
		if err != nil {
			return err
		}
	}
	return nil
}

func seedNotificationChannels(ctx context.Context, tx pgx.Tx, now time.Time) error {
	type ch struct {
		id, typ, name, config string
	}
	channels := []ch{
		{"d0000000-0000-4000-a000-000000000b01", "slack", "Platform Engineering Alerts", `{"webhook_url":"https://hooks.slack.com/services/DEMO/CHANNEL/TOKEN","channel":"#chaos-alerts"}`},
		{"d0000000-0000-4000-a000-000000000b02", "email", "SRE Team", `{"recipients":["sre@company.com","oncall@company.com"]}`},
	}
	for _, c := range channels {
		_, err := tx.Exec(ctx, `
			INSERT INTO notification_channels (id, tenant_id, type, name, config, enabled, created_at)
			VALUES ($1, $2, $3, $4, $5::jsonb, true, $6)
			ON CONFLICT (id) DO NOTHING`,
			c.id, TenantID, c.typ, c.name, c.config, now.Add(-60*24*time.Hour))
		if err != nil {
			return err
		}
	}
	return nil
}

func seedPredictions(ctx context.Context, tx pgx.Tx, now time.Time) error {
	type pred struct {
		id, ptype, severity, title, desc, action, status string
		confidence                                       float64
	}
	preds := []pred{
		{"d0000000-0000-4000-a000-000000000c01", "capacity", "high", "Memory pressure on cache-layer",
			"Redis memory usage trending toward limit. At current growth rate, OOM expected within 48 hours.",
			"Scale cache-layer memory limits from 1Gi to 2Gi or enable eviction policy", "open", 0.87},
		{"d0000000-0000-4000-a000-000000000c02", "failure_risk", "medium", "Increased error rate on payment-svc",
			"Payment service error rate has increased 3x in the last 6 hours. Pattern matches pre-failure signature from previous incidents.",
			"Investigate payment-svc logs for upstream dependency issues", "acknowledged", 0.73},
		{"d0000000-0000-4000-a000-000000000c03", "performance_degradation", "low", "Performance degradation predicted for order-svc",
			"Order service P99 latency trending upward. Database connection pool saturation detected at 78%.",
			"Increase connection pool size or optimize slow queries", "open", 0.65},
	}
	for _, p := range preds {
		_, err := tx.Exec(ctx, `
			INSERT INTO predictive_analyses (id, tenant_id, environment_id, prediction_type, severity, title, description, confidence, recommended_action, status, predicted_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (id) DO NOTHING`,
			p.id, TenantID, EnvironmentID, p.ptype, p.severity, p.title, p.desc, p.confidence, p.action, p.status, now.Add(-6*time.Hour))
		if err != nil {
			return err
		}
	}
	return nil
}

func seedMarketplace(ctx context.Context, tx pgx.Tx, now time.Time) error {
	type plugin struct {
		id, name, display, desc, author, version, category string
		downloads                                          int
		rating                                             float64
		verified                                           bool
	}
	plugins := []plugin{
		{"d0000000-0000-4000-a000-000000000d01", "aws-fault-injection", "AWS Fault Injection", "Native AWS fault injection actions including EC2, RDS, and ECS chaos", "ChaosPlane", "2.1.0", "chaos_action", 12500, 4.8, true},
		{"d0000000-0000-4000-a000-000000000d02", "datadog-monitor", "Datadog Monitor", "Real-time observability integration with Datadog APM and metrics", "Datadog", "1.5.2", "monitoring", 8200, 4.6, true},
		{"d0000000-0000-4000-a000-000000000d03", "pagerduty-integration", "PagerDuty Integration", "Automatic incident creation and experiment status sync", "PagerDuty", "3.0.1", "integration", 6100, 4.5, true},
		{"d0000000-0000-4000-a000-000000000d04", "kubernetes-network-policies", "K8s Network Policies", "Advanced network chaos using Kubernetes NetworkPolicy resources", "Community", "0.9.4", "chaos_action", 3400, 4.2, false},
		{"d0000000-0000-4000-a000-000000000d05", "grafana-dashboard", "Grafana Dashboard", "Pre-built Grafana dashboards for chaos experiment visualization", "Grafana Labs", "1.2.0", "monitoring", 9800, 4.7, true},
	}
	for _, p := range plugins {
		_, err := tx.Exec(ctx, `
			INSERT INTO marketplace_plugins (id, name, display_name, description, author, version, category, downloads, rating, verified, published_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (name) DO NOTHING`,
			p.id, p.name, p.display, p.desc, p.author, p.version, p.category, p.downloads, p.rating, p.verified, now.Add(-90*24*time.Hour))
		if err != nil {
			return err
		}
	}
	installs := []string{"d0000000-0000-4000-a000-000000000d01", "d0000000-0000-4000-a000-000000000d02"}
	for _, pid := range installs {
		_, err := tx.Exec(ctx, `
			INSERT INTO marketplace_installs (tenant_id, plugin_id, installed_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (tenant_id, plugin_id) DO NOTHING`,
			TenantID, pid, now.Add(-30*24*time.Hour))
		if err != nil {
			return err
		}
	}
	return nil
}

func seedCICDIntegrations(ctx context.Context, tx pgx.Tx, now time.Time) error {
	type ci struct {
		id, provider, name, config string
	}
	integrations := []ci{
		{"d0000000-0000-4000-a000-000000000e01", "github_actions", "Production Deploy Pipeline", `{"repo":"chaosplane-hq/chaosplane-platform","workflow":"deploy.yml","branch":"main"}`},
		{"d0000000-0000-4000-a000-000000000e02", "custom", "ArgoCD Sync Hook", `{"url":"https://argocd.internal/api/v1/hooks","events":["sync_succeeded","sync_failed"]}`},
	}
	for _, c := range integrations {
		_, err := tx.Exec(ctx, `
			INSERT INTO cicd_integrations (id, tenant_id, provider, name, config, enabled, last_triggered, created_at)
			VALUES ($1, $2, $3, $4, $5::jsonb, true, $6, $7)
			ON CONFLICT (id) DO NOTHING`,
			c.id, TenantID, c.provider, c.name, c.config, now.Add(-time.Hour), now.Add(-60*24*time.Hour))
		if err != nil {
			return err
		}
	}
	return nil
}

func seedTopologyDrifts(ctx context.Context, tx pgx.Tx, now time.Time) error {
	type drift struct {
		id, driftType, kind, name string
		prev, curr                string
		acked                     bool
	}
	drifts := []drift{
		{"d0000000-0000-4000-a000-000000000f01", "scaled", "Deployment", "order-svc", `{"replicas":2}`, `{"replicas":3}`, false},
		{"d0000000-0000-4000-a000-000000000f02", "added", "Deployment", "feature-flag-svc", `null`, `{"replicas":1,"image":"feature-flag:v0.1"}`, false},
		{"d0000000-0000-4000-a000-000000000f03", "modified", "Deployment", "api-gateway", `{"image":"api-gw:1.2.0"}`, `{"image":"api-gw:1.3.0"}`, true},
	}
	for _, d := range drifts {
		var ackedAt *time.Time
		var ackedBy *string
		if d.acked {
			t := now.Add(-10 * time.Minute)
			ackedAt = &t
			uid := UserID
			ackedBy = &uid
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO topology_drifts (id, tenant_id, environment_id, drift_type, resource_kind, resource_name, resource_namespace, previous_state, current_state, detected_at, acknowledged_at, acknowledged_by)
			VALUES ($1, $2, $3, $4, $5, $6, 'default', $7::jsonb, $8::jsonb, $9, $10, $11)
			ON CONFLICT (id) DO NOTHING`,
			d.id, TenantID, EnvironmentID, d.driftType, d.kind, d.name, d.prev, d.curr, now.Add(-2*time.Hour), ackedAt, ackedBy)
		if err != nil {
			return err
		}
	}
	return nil
}

func seedTopologyMetrics(ctx context.Context, tx pgx.Tx, now time.Time) error {
	type metric struct {
		svc    string
		cpu    float64
		memory float64
	}
	metrics := []metric{
		{"api-gateway", 45.2, 312},
		{"order-svc", 67.8, 456},
		{"payment-svc", 23.4, 198},
		{"cache-layer", 15.6, 890},
		{"postgres-primary", 38.9, 1024},
		{"message-queue", 12.3, 256},
	}
	for _, m := range metrics {
		labels := fmt.Sprintf(`{"service":"%s","namespace":"default"}`, m.svc)
		_, err := tx.Exec(ctx, `
			INSERT INTO environment_metrics (tenant_id, environment_id, metric_name, metric_value, labels, collected_at)
			VALUES ($1, $2, 'cpu_usage_percent', $3, $4::jsonb, $5)
			ON CONFLICT DO NOTHING`,
			TenantID, EnvironmentID, m.cpu, labels, now)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO environment_metrics (tenant_id, environment_id, metric_name, metric_value, labels, collected_at)
			VALUES ($1, $2, 'memory_usage_mb', $3, $4::jsonb, $5)
			ON CONFLICT DO NOTHING`,
			TenantID, EnvironmentID, m.memory, labels, now)
		if err != nil {
			return err
		}
	}
	return nil
}
