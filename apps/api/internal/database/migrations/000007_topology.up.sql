CREATE TABLE topology_snapshots (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    environment_id  uuid NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    nodes           jsonb NOT NULL DEFAULT '[]',
    namespaces      jsonb NOT NULL DEFAULT '[]',
    services        jsonb NOT NULL DEFAULT '[]',
    deployments     jsonb NOT NULL DEFAULT '[]',
    pods            jsonb NOT NULL DEFAULT '[]',
    collected_at    timestamptz NOT NULL DEFAULT now(),
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_topology_snapshots_env ON topology_snapshots(environment_id);
CREATE INDEX idx_topology_snapshots_tenant ON topology_snapshots(tenant_id);
CREATE INDEX idx_topology_snapshots_collected ON topology_snapshots(collected_at DESC);
