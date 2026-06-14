CREATE TABLE cui_markings (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    resource_type   text NOT NULL,
    resource_id     text NOT NULL,
    cui_category    text NOT NULL CHECK (cui_category IN ('CUI','CUI//SP-CTI','CUI//SP-EXPT','CUI//SP-PRVCY')),
    marking_applied_by uuid REFERENCES users(id),
    created_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, resource_type, resource_id)
);

CREATE INDEX idx_cui_markings_tenant ON cui_markings(tenant_id);
CREATE INDEX idx_cui_markings_resource ON cui_markings(resource_type, resource_id);
