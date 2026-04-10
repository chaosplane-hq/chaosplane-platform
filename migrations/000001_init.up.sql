-- ============================================================
-- Migration: 000001_init
-- Description: Phase 0 bootstrap — 17 core tables, RLS, triggers, indexes
-- PostgreSQL 16
-- ============================================================

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- Shared trigger function: auto-update updated_at
-- ============================================================
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- 1. tenants
-- ============================================================
CREATE TABLE tenants (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name            text NOT NULL,
    slug            text NOT NULL UNIQUE,
    plan            text NOT NULL DEFAULT 'free'
                    CHECK (plan IN ('free','team','business','enterprise','government')),
    status          text NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active','suspended','cancelled')),
    settings        jsonb NOT NULL DEFAULT '{}',
    max_environments int NOT NULL DEFAULT 1,
    max_users       int NOT NULL DEFAULT 5,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz
);

CREATE UNIQUE INDEX idx_tenants_slug ON tenants(slug) WHERE deleted_at IS NULL;

CREATE TRIGGER trg_tenants_updated_at
    BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- 2. users (cross-tenant — NO RLS)
-- ============================================================
CREATE TABLE users (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email               text NOT NULL UNIQUE,
    name                text NOT NULL,
    avatar_url          text,
    password_hash       text,
    email_verified      boolean NOT NULL DEFAULT false,
    email_verified_at   timestamptz,
    status              text NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active','invited','disabled','locked')),
    accepted_tos_at     timestamptz,
    accepted_tos_version text,
    failed_login_count  int NOT NULL DEFAULT 0,
    locked_until        timestamptz,
    last_login_at       timestamptz,
    last_login_ip       inet,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    deleted_at          timestamptz
);

CREATE UNIQUE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- 3. user_tenants
-- ============================================================
CREATE TABLE user_tenants (
    user_id         uuid NOT NULL REFERENCES users(id),
    tenant_id       uuid NOT NULL REFERENCES tenants(id),
    default_org_id  uuid,  -- FK added after organizations table
    joined_at       timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, tenant_id)
);

CREATE INDEX idx_user_tenants_tenant ON user_tenants(tenant_id);
CREATE INDEX idx_user_tenants_user ON user_tenants(user_id);

-- ============================================================
-- 4. organizations
-- ============================================================
CREATE TABLE organizations (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id),
    name            text NOT NULL,
    slug            text NOT NULL,
    description     text,
    settings        jsonb NOT NULL DEFAULT '{}',
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz,
    UNIQUE (tenant_id, slug)
);

CREATE INDEX idx_organizations_tenant ON organizations(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_organizations_tenant_slug ON organizations(tenant_id, slug);

CREATE TRIGGER trg_organizations_updated_at
    BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Add deferred FK from user_tenants.default_org_id → organizations
ALTER TABLE user_tenants
    ADD CONSTRAINT fk_user_tenants_default_org
    FOREIGN KEY (default_org_id) REFERENCES organizations(id);

-- ============================================================
-- 5. workspaces
-- ============================================================
CREATE TABLE workspaces (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id),
    organization_id uuid NOT NULL REFERENCES organizations(id),
    name            text NOT NULL,
    slug            text NOT NULL,
    description     text,
    settings        jsonb NOT NULL DEFAULT '{}',
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz,
    UNIQUE (organization_id, slug)
);

CREATE INDEX idx_workspaces_org ON workspaces(organization_id) WHERE deleted_at IS NULL;

CREATE TRIGGER trg_workspaces_updated_at
    BEFORE UPDATE ON workspaces
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- 6. teams
-- ============================================================
CREATE TABLE teams (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id),
    workspace_id    uuid NOT NULL REFERENCES workspaces(id),
    name            text NOT NULL,
    slug            text NOT NULL,
    description     text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz,
    UNIQUE (workspace_id, slug)
);

CREATE TRIGGER trg_teams_updated_at
    BEFORE UPDATE ON teams
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- 7. team_members
-- ============================================================
CREATE TABLE team_members (
    team_id         uuid NOT NULL REFERENCES teams(id),
    user_id         uuid NOT NULL REFERENCES users(id),
    role            text NOT NULL DEFAULT 'member'
                    CHECK (role IN ('lead','member','viewer')),
    joined_at       timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (team_id, user_id)
);

-- ============================================================
-- 8. projects
-- ============================================================
CREATE TABLE projects (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id),
    workspace_id    uuid NOT NULL REFERENCES workspaces(id),
    name            text NOT NULL,
    slug            text NOT NULL,
    description     text,
    settings        jsonb NOT NULL DEFAULT '{}',
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz,
    UNIQUE (workspace_id, slug)
);

CREATE INDEX idx_projects_workspace ON projects(workspace_id) WHERE deleted_at IS NULL;

CREATE TRIGGER trg_projects_updated_at
    BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- 9. environments
-- ============================================================
CREATE TABLE environments (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id),
    project_id      uuid NOT NULL REFERENCES projects(id),
    name            text NOT NULL,
    slug            text NOT NULL,
    type            text NOT NULL DEFAULT 'staging'
                    CHECK (type IN ('development','staging','production')),
    cluster_info    jsonb NOT NULL DEFAULT '{}',
    agent_status    text NOT NULL DEFAULT 'disconnected'
                    CHECK (agent_status IN ('connected','disconnected','degraded')),
    agent_version   text,
    last_heartbeat  timestamptz,
    settings        jsonb NOT NULL DEFAULT '{}',
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz,
    UNIQUE (project_id, slug)
);

CREATE INDEX idx_environments_project ON environments(project_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_environments_agent ON environments(agent_status);

CREATE TRIGGER trg_environments_updated_at
    BEFORE UPDATE ON environments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- 10. experiments
-- ============================================================
CREATE TABLE experiments (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           uuid NOT NULL REFERENCES tenants(id),
    environment_id      uuid NOT NULL REFERENCES environments(id),
    name                text NOT NULL,
    description         text,
    hypothesis          text,
    experiment_type     text NOT NULL DEFAULT 'single'
                        CHECK (experiment_type IN ('single','workflow','gameday')),
    target              jsonb NOT NULL,
    action              jsonb NOT NULL,
    steady_state        jsonb,
    rollback            jsonb,
    abort_conditions    jsonb DEFAULT '[]',
    duration_seconds    int NOT NULL DEFAULT 60,
    status              text NOT NULL DEFAULT 'draft'
                        CHECK (status IN (
                            'draft','pending_approval','approved','rejected',
                            'scheduled','running','paused','completed',
                            'failed','aborted','rolled_back'
                        )),
    approved_by         uuid REFERENCES users(id),
    approved_at         timestamptz,
    created_by          uuid NOT NULL REFERENCES users(id),
    tags                text[] DEFAULT '{}',
    metadata            jsonb NOT NULL DEFAULT '{}',
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    deleted_at          timestamptz
);

CREATE INDEX idx_experiments_env ON experiments(environment_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_experiments_status ON experiments(tenant_id, status);
CREATE INDEX idx_experiments_tenant_env ON experiments(tenant_id, environment_id);
CREATE INDEX idx_experiments_created_by ON experiments(created_by);
CREATE INDEX idx_experiments_tags ON experiments USING gin(tags);

CREATE TRIGGER trg_experiments_updated_at
    BEFORE UPDATE ON experiments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- 11. experiment_results
-- ============================================================
CREATE TABLE experiment_results (
    id                      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id               uuid NOT NULL REFERENCES tenants(id),
    experiment_id           uuid NOT NULL REFERENCES experiments(id),
    run_number              int NOT NULL DEFAULT 1,
    status                  text NOT NULL
                            CHECK (status IN ('running','completed','failed','aborted')),
    started_at              timestamptz NOT NULL DEFAULT now(),
    finished_at             timestamptz,
    duration_ms             bigint,
    steady_state_before     jsonb,
    steady_state_after      jsonb,
    steady_state_met        boolean,
    impact_summary          jsonb NOT NULL DEFAULT '{}',
    metrics                 jsonb NOT NULL DEFAULT '{}',
    error_message           text,
    created_at              timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_exp_results_experiment ON experiment_results(experiment_id);
CREATE INDEX idx_exp_results_status ON experiment_results(tenant_id, status);
CREATE UNIQUE INDEX idx_exp_results_run ON experiment_results(experiment_id, run_number);

-- ============================================================
-- 12. email_verification_tokens (NO RLS — service role)
-- ============================================================
CREATE TABLE email_verification_tokens (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      text NOT NULL UNIQUE,
    expires_at      timestamptz NOT NULL,
    used_at         timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_email_verify_user ON email_verification_tokens(user_id) WHERE used_at IS NULL;
CREATE INDEX idx_email_verify_expires ON email_verification_tokens(expires_at) WHERE used_at IS NULL;

-- ============================================================
-- 13. password_reset_tokens (NO RLS — service role)
-- ============================================================
CREATE TABLE password_reset_tokens (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      text NOT NULL UNIQUE,
    expires_at      timestamptz NOT NULL,
    used_at         timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_pw_reset_user ON password_reset_tokens(user_id) WHERE used_at IS NULL;
CREATE INDEX idx_pw_reset_expires ON password_reset_tokens(expires_at) WHERE used_at IS NULL;

-- ============================================================
-- 14. refresh_tokens (NO RLS — service role)
-- ============================================================
CREATE TABLE refresh_tokens (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id       uuid NOT NULL REFERENCES tenants(id),
    token_hash      text NOT NULL UNIQUE,
    family_id       uuid NOT NULL,
    parent_id       uuid REFERENCES refresh_tokens(id),
    ip_address      inet,
    user_agent      text,
    expires_at      timestamptz NOT NULL,
    revoked_at      timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_user ON refresh_tokens(user_id) WHERE revoked_at IS NULL;
CREATE INDEX idx_refresh_family ON refresh_tokens(family_id);
CREATE INDEX idx_refresh_expires ON refresh_tokens(expires_at) WHERE revoked_at IS NULL;

-- ============================================================
-- 15. jwt_blacklist (NO RLS — service role)
-- ============================================================
CREATE TABLE jwt_blacklist (
    jti             uuid PRIMARY KEY,
    user_id         uuid NOT NULL REFERENCES users(id),
    reason          text NOT NULL
                    CHECK (reason IN ('password_change','logout','forced_termination','permission_change','key_revocation')),
    expires_at      timestamptz NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_jwt_blacklist_expires ON jwt_blacklist(expires_at);

-- ============================================================
-- 16. invitations (NO RLS — service role)
-- ============================================================
CREATE TABLE invitations (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id),
    organization_id uuid NOT NULL REFERENCES organizations(id),
    team_id         uuid REFERENCES teams(id),
    email           text NOT NULL,
    role            text NOT NULL DEFAULT 'member'
                    CHECK (role IN ('admin','editor','viewer','operator','member')),
    invited_by      uuid NOT NULL REFERENCES users(id),
    token_hash      text NOT NULL UNIQUE,
    status          text NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','accepted','declined','expired','revoked')),
    expires_at      timestamptz NOT NULL,
    accepted_at     timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_invitations_tenant ON invitations(tenant_id) WHERE status = 'pending';
CREATE INDEX idx_invitations_email ON invitations(email) WHERE status = 'pending';
CREATE INDEX idx_invitations_tenant_email ON invitations(tenant_id, email);
CREATE INDEX idx_invitations_token ON invitations(token_hash);
CREATE INDEX idx_invitations_expires ON invitations(expires_at) WHERE status = 'pending';

-- ============================================================
-- 17. onboarding_progress (NO RLS — service role)
-- ============================================================
CREATE TABLE onboarding_progress (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id       uuid NOT NULL REFERENCES tenants(id),
    step_org_created        boolean NOT NULL DEFAULT false,
    step_workspace_created  boolean NOT NULL DEFAULT false,
    step_team_created       boolean NOT NULL DEFAULT false,
    step_member_invited     boolean NOT NULL DEFAULT false,
    step_cluster_connected  boolean NOT NULL DEFAULT false,
    step_first_experiment   boolean NOT NULL DEFAULT false,
    step_result_viewed      boolean NOT NULL DEFAULT false,
    completed_at    timestamptz,
    skipped_at      timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, tenant_id)
);

CREATE INDEX idx_onboarding_user ON onboarding_progress(user_id) WHERE completed_at IS NULL;

CREATE TRIGGER trg_onboarding_progress_updated_at
    BEFORE UPDATE ON onboarding_progress
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- Row-Level Security (8 tenant-scoped tables)
-- ============================================================

-- tenants: isolate by id
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenants FORCE ROW LEVEL SECURITY;
CREATE POLICY tenants_tenant_isolation ON tenants
    USING (id = current_setting('app.current_tenant_id')::uuid);

-- organizations
ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE organizations FORCE ROW LEVEL SECURITY;
CREATE POLICY organizations_tenant_isolation ON organizations
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- workspaces
ALTER TABLE workspaces ENABLE ROW LEVEL SECURITY;
ALTER TABLE workspaces FORCE ROW LEVEL SECURITY;
CREATE POLICY workspaces_tenant_isolation ON workspaces
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- teams
ALTER TABLE teams ENABLE ROW LEVEL SECURITY;
ALTER TABLE teams FORCE ROW LEVEL SECURITY;
CREATE POLICY teams_tenant_isolation ON teams
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- projects
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;
ALTER TABLE projects FORCE ROW LEVEL SECURITY;
CREATE POLICY projects_tenant_isolation ON projects
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- environments
ALTER TABLE environments ENABLE ROW LEVEL SECURITY;
ALTER TABLE environments FORCE ROW LEVEL SECURITY;
CREATE POLICY environments_tenant_isolation ON environments
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- experiments
ALTER TABLE experiments ENABLE ROW LEVEL SECURITY;
ALTER TABLE experiments FORCE ROW LEVEL SECURITY;
CREATE POLICY experiments_tenant_isolation ON experiments
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- experiment_results
ALTER TABLE experiment_results ENABLE ROW LEVEL SECURITY;
ALTER TABLE experiment_results FORCE ROW LEVEL SECURITY;
CREATE POLICY experiment_results_tenant_isolation ON experiment_results
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
