CREATE TABLE gamedays (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    environment_id  uuid NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    title           text NOT NULL,
    description     text,
    status          text NOT NULL DEFAULT 'planning'
                    CHECK (status IN ('planning','scheduled','running','reviewing','completed','cancelled')),
    scheduled_at    timestamptz,
    started_at      timestamptz,
    ended_at        timestamptz,
    created_by      uuid NOT NULL REFERENCES users(id),
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_gamedays_tenant ON gamedays(tenant_id);
CREATE INDEX idx_gamedays_env ON gamedays(environment_id);

CREATE TRIGGER trg_gamedays_updated_at
    BEFORE UPDATE ON gamedays FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE gameday_participants (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    gameday_id      uuid NOT NULL REFERENCES gamedays(id) ON DELETE CASCADE,
    user_id         uuid NOT NULL REFERENCES users(id),
    role            text NOT NULL DEFAULT 'observer'
                    CHECK (role IN ('facilitator','operator','observer')),
    joined_at       timestamptz NOT NULL DEFAULT now(),
    UNIQUE (gameday_id, user_id)
);

CREATE TABLE gameday_events (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    gameday_id      uuid NOT NULL REFERENCES gamedays(id) ON DELETE CASCADE,
    event_type      text NOT NULL,
    title           text NOT NULL,
    description     text,
    user_id         uuid REFERENCES users(id),
    metadata        jsonb NOT NULL DEFAULT '{}',
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_gameday_events_gameday ON gameday_events(gameday_id);

CREATE TABLE gameday_postmortems (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    gameday_id      uuid NOT NULL REFERENCES gamedays(id) ON DELETE CASCADE,
    summary         text NOT NULL,
    what_went_well  text,
    what_went_wrong text,
    action_items    jsonb NOT NULL DEFAULT '[]',
    created_by      uuid NOT NULL REFERENCES users(id),
    created_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE (gameday_id)
);

CREATE TABLE resilience_scores (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    environment_id  uuid NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    overall_grade   text NOT NULL CHECK (overall_grade IN ('A','B','C','D','F')),
    overall_score   double precision NOT NULL,
    availability    double precision NOT NULL DEFAULT 0,
    fault_tolerance double precision NOT NULL DEFAULT 0,
    recoverability  double precision NOT NULL DEFAULT 0,
    details         jsonb NOT NULL DEFAULT '{}',
    calculated_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_resilience_scores_env ON resilience_scores(environment_id);
CREATE INDEX idx_resilience_scores_tenant ON resilience_scores(tenant_id);
CREATE INDEX idx_resilience_scores_calculated ON resilience_scores(calculated_at DESC);

CREATE TABLE workflow_templates (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid REFERENCES tenants(id) ON DELETE CASCADE,
    name            text NOT NULL,
    description     text,
    category        text NOT NULL DEFAULT 'custom',
    is_public       boolean NOT NULL DEFAULT false,
    spec            jsonb NOT NULL DEFAULT '{}',
    created_by      uuid REFERENCES users(id),
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_workflow_templates_tenant ON workflow_templates(tenant_id);
CREATE INDEX idx_workflow_templates_public ON workflow_templates(is_public) WHERE is_public = true;

CREATE TRIGGER trg_workflow_templates_updated_at
    BEFORE UPDATE ON workflow_templates FOR EACH ROW EXECUTE FUNCTION set_updated_at();
