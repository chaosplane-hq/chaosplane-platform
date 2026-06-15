CREATE TABLE blast_radius_policies (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           uuid NOT NULL REFERENCES tenants(id),
    environment_id      uuid REFERENCES environments(id),
    name                text NOT NULL,
    description         text,
    enforcement         text NOT NULL DEFAULT 'audit'
                        CHECK (enforcement IN ('enforce','audit','disabled')),
    max_concurrent      int,
    max_targets         int,
    allowed_namespaces  text[] NOT NULL DEFAULT '{}',
    blocked_namespaces  text[] NOT NULL DEFAULT '{}',
    created_by          uuid NOT NULL REFERENCES users(id),
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    deleted_at          timestamptz
);

CREATE INDEX idx_blast_radius_policies_tenant ON blast_radius_policies(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_blast_radius_policies_env ON blast_radius_policies(environment_id) WHERE deleted_at IS NULL;

CREATE TRIGGER trg_blast_radius_policies_updated_at
    BEFORE UPDATE ON blast_radius_policies
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE blast_radius_policies ENABLE ROW LEVEL SECURITY;
ALTER TABLE blast_radius_policies FORCE ROW LEVEL SECURITY;
CREATE POLICY blast_radius_policies_tenant_isolation ON blast_radius_policies
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
