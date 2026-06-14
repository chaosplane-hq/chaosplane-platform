CREATE TABLE service_dependencies (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    environment_id  uuid NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    source_kind     text NOT NULL,
    source_name     text NOT NULL,
    source_namespace text NOT NULL,
    target_kind     text NOT NULL,
    target_name     text NOT NULL,
    target_namespace text NOT NULL,
    protocol        text,
    port            int,
    discovered_at   timestamptz NOT NULL DEFAULT now(),
    last_seen_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE (environment_id, source_kind, source_name, source_namespace, target_kind, target_name, target_namespace)
);

CREATE INDEX idx_service_deps_env ON service_dependencies(environment_id);
CREATE INDEX idx_service_deps_tenant ON service_dependencies(tenant_id);

CREATE TABLE topology_drifts (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    environment_id  uuid NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    drift_type      text NOT NULL CHECK (drift_type IN ('added','removed','modified','scaled')),
    resource_kind   text NOT NULL,
    resource_name   text NOT NULL,
    resource_namespace text NOT NULL,
    previous_state  jsonb,
    current_state   jsonb,
    detected_at     timestamptz NOT NULL DEFAULT now(),
    acknowledged_at timestamptz,
    acknowledged_by uuid REFERENCES users(id)
);

CREATE INDEX idx_topology_drifts_env ON topology_drifts(environment_id);
CREATE INDEX idx_topology_drifts_tenant ON topology_drifts(tenant_id);
CREATE INDEX idx_topology_drifts_unacked ON topology_drifts(environment_id) WHERE acknowledged_at IS NULL;

CREATE TABLE environment_metrics (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    environment_id  uuid NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    metric_name     text NOT NULL,
    metric_value    double precision NOT NULL,
    labels          jsonb NOT NULL DEFAULT '{}',
    collected_at    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_env_metrics_env ON environment_metrics(environment_id, metric_name);
CREATE INDEX idx_env_metrics_collected ON environment_metrics(collected_at DESC);

CREATE TABLE vulnerability_findings (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    environment_id  uuid NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    category        text NOT NULL CHECK (category IN ('spof','resource_limits','healthcheck','pdb','networking','storage','security')),
    severity        text NOT NULL CHECK (severity IN ('critical','high','medium','low','info')),
    title           text NOT NULL,
    description     text NOT NULL,
    resource_kind   text NOT NULL,
    resource_name   text NOT NULL,
    resource_namespace text NOT NULL,
    remediation     text,
    suggested_experiment jsonb,
    status          text NOT NULL DEFAULT 'open' CHECK (status IN ('open','acknowledged','resolved','ignored')),
    detected_at     timestamptz NOT NULL DEFAULT now(),
    resolved_at     timestamptz
);

CREATE INDEX idx_vuln_findings_env ON vulnerability_findings(environment_id);
CREATE INDEX idx_vuln_findings_tenant ON vulnerability_findings(tenant_id);
CREATE INDEX idx_vuln_findings_status ON vulnerability_findings(status) WHERE status = 'open';
CREATE INDEX idx_vuln_findings_severity ON vulnerability_findings(severity);
