CREATE TABLE agent_tokens (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    environment_id  uuid NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    token_hash      text NOT NULL UNIQUE,
    name            text NOT NULL,
    revoked_at      timestamptz,
    last_used_at    timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE (environment_id, name)
);

CREATE INDEX idx_agent_tokens_env ON agent_tokens(environment_id) WHERE revoked_at IS NULL;
CREATE INDEX idx_agent_tokens_tenant ON agent_tokens(tenant_id) WHERE revoked_at IS NULL;
