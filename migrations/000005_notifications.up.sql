CREATE TABLE notification_channels (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    type            text NOT NULL CHECK (type IN ('slack','email','webhook')),
    name            text NOT NULL,
    config          jsonb NOT NULL DEFAULT '{}',
    enabled         boolean NOT NULL DEFAULT true,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_notification_channels_tenant ON notification_channels(tenant_id) WHERE enabled = true;

CREATE TRIGGER trg_notification_channels_updated_at
    BEFORE UPDATE ON notification_channels
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE notification_rules (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    channel_id      uuid NOT NULL REFERENCES notification_channels(id) ON DELETE CASCADE,
    event_type      text NOT NULL,
    filters         jsonb NOT NULL DEFAULT '{}',
    enabled         boolean NOT NULL DEFAULT true,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_notification_rules_tenant ON notification_rules(tenant_id) WHERE enabled = true;
CREATE INDEX idx_notification_rules_event ON notification_rules(event_type) WHERE enabled = true;

CREATE TRIGGER trg_notification_rules_updated_at
    BEFORE UPDATE ON notification_rules
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE notification_history (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    channel_id      uuid NOT NULL REFERENCES notification_channels(id),
    rule_id         uuid REFERENCES notification_rules(id),
    event_type      text NOT NULL,
    status          text NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','sent','failed')),
    payload         jsonb NOT NULL DEFAULT '{}',
    error_message   text,
    sent_at         timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_notification_history_tenant ON notification_history(tenant_id);
CREATE INDEX idx_notification_history_channel ON notification_history(channel_id);
