DROP INDEX IF EXISTS idx_experiments_claimable;

ALTER TABLE experiments
    DROP COLUMN IF EXISTS desired_state,
    DROP COLUMN IF EXISTS generation,
    DROP COLUMN IF EXISTS observed_generation,
    DROP COLUMN IF EXISTS claimed_by,
    DROP COLUMN IF EXISTS claim_expires_at,
    DROP COLUMN IF EXISTS run_started_at,
    DROP COLUMN IF EXISTS run_ended_at,
    DROP COLUMN IF EXISTS last_agent_report_at;
