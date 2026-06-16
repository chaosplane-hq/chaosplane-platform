ALTER TABLE experiments
    ADD COLUMN IF NOT EXISTS steps jsonb;

ALTER TABLE experiments
    ALTER COLUMN action DROP NOT NULL,
    ALTER COLUMN target DROP NOT NULL;

ALTER TABLE experiments
    ADD CONSTRAINT experiments_shape_by_type CHECK (
        (experiment_type = 'workflow' AND steps IS NOT NULL)
        OR (experiment_type <> 'workflow' AND action IS NOT NULL AND target IS NOT NULL)
    );
