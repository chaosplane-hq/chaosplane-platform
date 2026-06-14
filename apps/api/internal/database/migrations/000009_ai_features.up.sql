CREATE TABLE experiment_suggestions (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    environment_id  uuid NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    finding_id      uuid REFERENCES vulnerability_findings(id),
    source          text NOT NULL CHECK (source IN ('vulnerability','best_practice','ai','manual')),
    title           text NOT NULL,
    description     text NOT NULL,
    action_type     text NOT NULL,
    target_namespace text NOT NULL,
    target_name     text NOT NULL,
    duration        text NOT NULL DEFAULT '30s',
    parameters      jsonb NOT NULL DEFAULT '{}',
    confidence      double precision NOT NULL DEFAULT 0.5,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_experiment_suggestions_env ON experiment_suggestions(environment_id);
CREATE INDEX idx_experiment_suggestions_tenant ON experiment_suggestions(tenant_id);
CREATE INDEX idx_experiment_suggestions_finding ON experiment_suggestions(finding_id);

CREATE TABLE experiment_results_analysis (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    experiment_name text NOT NULL,
    environment_id  uuid REFERENCES environments(id),
    summary         text NOT NULL,
    impact_analysis text,
    recommendations text,
    severity_assessment text CHECK (severity_assessment IN ('none','low','medium','high','critical')),
    affected_services jsonb NOT NULL DEFAULT '[]',
    metrics_impact  jsonb NOT NULL DEFAULT '{}',
    analyzed_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_results_analysis_tenant ON experiment_results_analysis(tenant_id);
CREATE INDEX idx_results_analysis_env ON experiment_results_analysis(environment_id);

CREATE TABLE ai_chat_sessions (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id         uuid NOT NULL REFERENCES users(id),
    title           text NOT NULL DEFAULT 'New Chat',
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE ai_chat_messages (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id      uuid NOT NULL REFERENCES ai_chat_sessions(id) ON DELETE CASCADE,
    role            text NOT NULL CHECK (role IN ('user','assistant','system')),
    content         text NOT NULL,
    metadata        jsonb NOT NULL DEFAULT '{}',
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_ai_chat_sessions_user ON ai_chat_sessions(user_id);
CREATE INDEX idx_ai_chat_messages_session ON ai_chat_messages(session_id);

CREATE TRIGGER trg_ai_chat_sessions_updated_at
    BEFORE UPDATE ON ai_chat_sessions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
