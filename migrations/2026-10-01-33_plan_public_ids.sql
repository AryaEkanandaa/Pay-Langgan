-- +goose Up
ALTER TABLE plans ADD COLUMN public_id VARCHAR(32);

UPDATE plans
SET public_id = 'pl_plan_' || SUBSTRING(MD5(id::text || ':' || random()::text || ':' || clock_timestamp()::text) FROM 1 FOR 24);

ALTER TABLE plans
    ALTER COLUMN public_id SET DEFAULT ('pl_plan_' || SUBSTRING(MD5(random()::text || ':' || clock_timestamp()::text || ':' || pg_backend_pid()::text) FROM 1 FOR 24)),
    ALTER COLUMN public_id SET NOT NULL;

ALTER TABLE plans ADD CONSTRAINT plans_public_id_format CHECK (public_id ~ '^pl_plan_[0-9a-f]{24}$');
CREATE UNIQUE INDEX plans_public_id_unique ON plans(public_id);

-- +goose Down
DROP INDEX plans_public_id_unique;
ALTER TABLE plans DROP CONSTRAINT plans_public_id_format;
ALTER TABLE plans DROP COLUMN public_id;
