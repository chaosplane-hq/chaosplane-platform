CREATE TABLE subscriptions (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    plan            text NOT NULL DEFAULT 'free'
                    CHECK (plan IN ('free','team','business','enterprise','government')),
    status          text NOT NULL DEFAULT 'trialing'
                    CHECK (status IN ('trialing','active','past_due','cancelled','suspended')),
    gateway         text NOT NULL DEFAULT 'none'
                    CHECK (gateway IN ('none','stripe','toss','dodo')),
    gateway_subscription_id text,
    gateway_customer_id     text,
    current_period_start    timestamptz,
    current_period_end      timestamptz,
    trial_ends_at           timestamptz,
    cancelled_at            timestamptz,
    suspended_at            timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id)
);

CREATE INDEX idx_subscriptions_status ON subscriptions(status) WHERE status != 'cancelled';
CREATE INDEX idx_subscriptions_gateway ON subscriptions(gateway, gateway_subscription_id);

CREATE TRIGGER trg_subscriptions_updated_at
    BEFORE UPDATE ON subscriptions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE payment_methods (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    gateway         text NOT NULL CHECK (gateway IN ('stripe','toss','dodo')),
    gateway_payment_method_id text NOT NULL,
    type            text NOT NULL DEFAULT 'card',
    last4           text,
    brand           text,
    is_default      boolean NOT NULL DEFAULT false,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_payment_methods_tenant ON payment_methods(tenant_id);

CREATE TABLE invoices (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    subscription_id uuid REFERENCES subscriptions(id),
    gateway         text NOT NULL CHECK (gateway IN ('stripe','toss','dodo')),
    gateway_invoice_id text,
    amount_cents    bigint NOT NULL,
    currency        text NOT NULL DEFAULT 'usd',
    status          text NOT NULL DEFAULT 'draft'
                    CHECK (status IN ('draft','open','paid','void','uncollectible')),
    period_start    timestamptz,
    period_end      timestamptz,
    paid_at         timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_invoices_tenant ON invoices(tenant_id);
CREATE INDEX idx_invoices_subscription ON invoices(subscription_id);

CREATE TABLE usage_records (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    metric          text NOT NULL CHECK (metric IN ('experiments','agents','api_calls')),
    quantity        bigint NOT NULL DEFAULT 0,
    period_start    timestamptz NOT NULL,
    period_end      timestamptz NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, metric, period_start)
);

CREATE INDEX idx_usage_records_tenant_period ON usage_records(tenant_id, period_start);

CREATE TABLE billing_events (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    event_type      text NOT NULL,
    gateway         text,
    gateway_event_id text,
    payload         jsonb NOT NULL DEFAULT '{}',
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_billing_events_tenant ON billing_events(tenant_id);
CREATE INDEX idx_billing_events_type ON billing_events(event_type);
