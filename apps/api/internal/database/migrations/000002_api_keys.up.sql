CREATE TABLE api_keys (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name            text NOT NULL,
    key_hash        bytea NOT NULL UNIQUE,
    last_used_at    timestamptz,
    expires_at      timestamptz,
    revoked_at      timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_api_keys_user ON api_keys(user_id) WHERE revoked_at IS NULL;
CREATE INDEX idx_api_keys_tenant ON api_keys(tenant_id) WHERE revoked_at IS NULL;
CREATE INDEX idx_api_keys_expires ON api_keys(expires_at) WHERE revoked_at IS NULL;
