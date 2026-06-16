ALTER TABLE experiments
    DROP CONSTRAINT IF EXISTS experiments_shape_by_type;

UPDATE experiments SET action = '{}'::jsonb WHERE action IS NULL;
UPDATE experiments SET target = '{}'::jsonb WHERE target IS NULL;

ALTER TABLE experiments
    ALTER COLUMN action SET NOT NULL,
    ALTER COLUMN target SET NOT NULL;

ALTER TABLE experiments
    DROP COLUMN IF EXISTS steps;
