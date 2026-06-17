-- Add RLS to high-value tenant-scoped tables that were missing it

ALTER TABLE agent_tokens ENABLE ROW LEVEL SECURITY;
CREATE POLICY agent_tokens_tenant ON agent_tokens
  USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;
CREATE POLICY audit_logs_tenant ON audit_logs
  USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

ALTER TABLE subscriptions ENABLE ROW LEVEL SECURITY;
CREATE POLICY subscriptions_tenant ON subscriptions
  USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

ALTER TABLE notification_channels ENABLE ROW LEVEL SECURITY;
CREATE POLICY notification_channels_tenant ON notification_channels
  USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

ALTER TABLE notification_rules ENABLE ROW LEVEL SECURITY;
CREATE POLICY notification_rules_tenant ON notification_rules
  USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

ALTER TABLE notification_history ENABLE ROW LEVEL SECURITY;
CREATE POLICY notification_history_tenant ON notification_history
  USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

ALTER TABLE topology_snapshots ENABLE ROW LEVEL SECURITY;
CREATE POLICY topology_snapshots_tenant ON topology_snapshots
  USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

ALTER TABLE service_dependencies ENABLE ROW LEVEL SECURITY;
CREATE POLICY service_dependencies_tenant ON service_dependencies
  USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

ALTER TABLE vulnerability_findings ENABLE ROW LEVEL SECURITY;
CREATE POLICY vulnerability_findings_tenant ON vulnerability_findings
  USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

ALTER TABLE resilience_scores ENABLE ROW LEVEL SECURITY;
CREATE POLICY resilience_scores_tenant ON resilience_scores
  USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

ALTER TABLE gamedays ENABLE ROW LEVEL SECURITY;
CREATE POLICY gamedays_tenant ON gamedays
  USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

ALTER TABLE active_sessions ENABLE ROW LEVEL SECURITY;
CREATE POLICY active_sessions_tenant ON active_sessions
  USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- Missing index for billing queries
CREATE INDEX IF NOT EXISTS idx_subscriptions_tenant ON subscriptions(tenant_id);

-- Missing composite index for AI chat session listing
CREATE INDEX IF NOT EXISTS idx_ai_chat_sessions_tenant_user ON ai_chat_sessions(tenant_id, user_id);
