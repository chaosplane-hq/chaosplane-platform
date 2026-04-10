CREATE TABLE saml_providers (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name            text NOT NULL,
    entity_id       text NOT NULL,
    sso_url         text NOT NULL,
    certificate     text NOT NULL,
    metadata_url    text,
    enabled         boolean NOT NULL DEFAULT true,
    jit_provisioning boolean NOT NULL DEFAULT false,
    default_role    text NOT NULL DEFAULT 'viewer',
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, entity_id)
);

CREATE INDEX idx_saml_providers_tenant ON saml_providers(tenant_id) WHERE enabled = true;

CREATE TRIGGER trg_saml_providers_updated_at
    BEFORE UPDATE ON saml_providers
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE scim_tokens (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    token_hash      text NOT NULL UNIQUE,
    revoked_at      timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_scim_tokens_tenant ON scim_tokens(tenant_id) WHERE revoked_at IS NULL;

CREATE TABLE mfa_recovery_codes (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash       text NOT NULL,
    used_at         timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_mfa_recovery_user ON mfa_recovery_codes(user_id) WHERE used_at IS NULL;

CREATE TABLE abac_policies (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name            text NOT NULL,
    description     text,
    effect          text NOT NULL CHECK (effect IN ('allow','deny')),
    subjects        jsonb NOT NULL DEFAULT '{}',
    resources       jsonb NOT NULL DEFAULT '{}',
    actions         jsonb NOT NULL DEFAULT '[]',
    conditions      jsonb NOT NULL DEFAULT '{}',
    priority        int NOT NULL DEFAULT 0,
    enabled         boolean NOT NULL DEFAULT true,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_abac_policies_tenant ON abac_policies(tenant_id) WHERE enabled = true;

CREATE TRIGGER trg_abac_policies_updated_at
    BEFORE UPDATE ON abac_policies
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE active_sessions (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    ip_address      inet,
    user_agent      text,
    last_activity   timestamptz NOT NULL DEFAULT now(),
    created_at      timestamptz NOT NULL DEFAULT now(),
    revoked_at      timestamptz
);

CREATE INDEX idx_active_sessions_user ON active_sessions(user_id) WHERE revoked_at IS NULL;
CREATE INDEX idx_active_sessions_tenant ON active_sessions(tenant_id) WHERE revoked_at IS NULL;

CREATE TABLE account_deletion_requests (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason          text,
    grace_period_ends timestamptz NOT NULL,
    cancelled_at    timestamptz,
    executed_at     timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_deletion_requests_user ON account_deletion_requests(user_id);
CREATE INDEX idx_deletion_requests_pending ON account_deletion_requests(grace_period_ends) WHERE executed_at IS NULL AND cancelled_at IS NULL;

CREATE TABLE email_change_requests (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    new_email       text NOT NULL,
    token_hash      text NOT NULL,
    expires_at      timestamptz NOT NULL,
    confirmed_at    timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_email_change_user ON email_change_requests(user_id);

CREATE TABLE audit_log_exports (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    destination     text NOT NULL CHECK (destination IN ('s3','minio','splunk','datadog')),
    config          jsonb NOT NULL DEFAULT '{}',
    status          text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','running','completed','failed')),
    started_at      timestamptz,
    completed_at    timestamptz,
    records_exported bigint DEFAULT 0,
    error_message   text,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_exports_tenant ON audit_log_exports(tenant_id);
