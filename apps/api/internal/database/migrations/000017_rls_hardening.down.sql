-- Reverse RLS hardening

DROP INDEX IF EXISTS idx_ai_chat_sessions_tenant_user;
DROP INDEX IF EXISTS idx_subscriptions_tenant;

DROP POLICY IF EXISTS active_sessions_tenant ON active_sessions;
ALTER TABLE active_sessions DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS gamedays_tenant ON gamedays;
ALTER TABLE gamedays DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS resilience_scores_tenant ON resilience_scores;
ALTER TABLE resilience_scores DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS vulnerability_findings_tenant ON vulnerability_findings;
ALTER TABLE vulnerability_findings DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS service_dependencies_tenant ON service_dependencies;
ALTER TABLE service_dependencies DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS topology_snapshots_tenant ON topology_snapshots;
ALTER TABLE topology_snapshots DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS notification_history_tenant ON notification_history;
ALTER TABLE notification_history DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS notification_rules_tenant ON notification_rules;
ALTER TABLE notification_rules DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS notification_channels_tenant ON notification_channels;
ALTER TABLE notification_channels DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS subscriptions_tenant ON subscriptions;
ALTER TABLE subscriptions DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS audit_logs_tenant ON audit_logs;
ALTER TABLE audit_logs DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS agent_tokens_tenant ON agent_tokens;
ALTER TABLE agent_tokens DISABLE ROW LEVEL SECURITY;
