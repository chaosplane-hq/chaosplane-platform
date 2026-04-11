CREATE TABLE marketplace_plugins (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name            text NOT NULL UNIQUE,
    display_name    text NOT NULL,
    description     text,
    author          text NOT NULL,
    version         text NOT NULL,
    category        text NOT NULL CHECK (category IN ('chaos_action','workflow_template','integration','monitoring')),
    oci_reference   text,
    downloads       bigint NOT NULL DEFAULT 0,
    rating          double precision NOT NULL DEFAULT 0,
    verified        boolean NOT NULL DEFAULT false,
    published_at    timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_marketplace_category ON marketplace_plugins(category);
CREATE INDEX idx_marketplace_downloads ON marketplace_plugins(downloads DESC);

CREATE TRIGGER trg_marketplace_updated_at
    BEFORE UPDATE ON marketplace_plugins FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE marketplace_installs (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    plugin_id       uuid NOT NULL REFERENCES marketplace_plugins(id),
    installed_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, plugin_id)
);

CREATE INDEX idx_marketplace_installs_tenant ON marketplace_installs(tenant_id);

CREATE TABLE federation_clusters (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name            text NOT NULL,
    region          text NOT NULL,
    provider        text NOT NULL CHECK (provider IN ('aws','azure','gcp','on-prem')),
    api_endpoint    text NOT NULL,
    status          text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','connected','disconnected','error')),
    agent_id        uuid REFERENCES environments(id),
    metadata        jsonb NOT NULL DEFAULT '{}',
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, name)
);

CREATE INDEX idx_federation_clusters_tenant ON federation_clusters(tenant_id);

CREATE TRIGGER trg_federation_clusters_updated_at
    BEFORE UPDATE ON federation_clusters FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE cicd_integrations (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    provider        text NOT NULL CHECK (provider IN ('github_actions','gitlab_ci','jenkins','circleci','custom')),
    name            text NOT NULL,
    config          jsonb NOT NULL DEFAULT '{}',
    webhook_secret  text,
    enabled         boolean NOT NULL DEFAULT true,
    last_triggered  timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_cicd_integrations_tenant ON cicd_integrations(tenant_id);

CREATE TRIGGER trg_cicd_integrations_updated_at
    BEFORE UPDATE ON cicd_integrations FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE predictive_analyses (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    environment_id  uuid NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    prediction_type text NOT NULL CHECK (prediction_type IN ('failure_risk','capacity','performance_degradation')),
    severity        text NOT NULL CHECK (severity IN ('critical','high','medium','low','info')),
    title           text NOT NULL,
    description     text NOT NULL,
    confidence      double precision NOT NULL,
    recommended_action text,
    auto_remediation jsonb,
    status          text NOT NULL DEFAULT 'open' CHECK (status IN ('open','acknowledged','resolved','dismissed')),
    predicted_at    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_predictive_analyses_env ON predictive_analyses(environment_id);
CREATE INDEX idx_predictive_analyses_tenant ON predictive_analyses(tenant_id);
