ALTER TABLE experiments
    ADD COLUMN IF NOT EXISTS desired_state        text NOT NULL DEFAULT 'run'
        CHECK (desired_state IN ('run','abort','pause')),
    ADD COLUMN IF NOT EXISTS generation           bigint NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS observed_generation  bigint,
    ADD COLUMN IF NOT EXISTS claimed_by           text,
    ADD COLUMN IF NOT EXISTS claim_expires_at     timestamptz,
    ADD COLUMN IF NOT EXISTS run_started_at       timestamptz,
    ADD COLUMN IF NOT EXISTS run_ended_at         timestamptz,
    ADD COLUMN IF NOT EXISTS last_agent_report_at timestamptz;

CREATE INDEX IF NOT EXISTS idx_experiments_claimable
    ON experiments (environment_id, created_at)
    WHERE deleted_at IS NULL AND status IN ('scheduled','running');
