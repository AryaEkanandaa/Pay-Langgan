-- +goose Up
ALTER TABLE checkout_sessions ADD COLUMN checkout_mode VARCHAR(20) NOT NULL DEFAULT 'SNAP';
ALTER TABLE checkout_sessions ADD COLUMN authentication_url TEXT;
ALTER TABLE checkout_sessions ADD CONSTRAINT chk_checkout_mode CHECK (checkout_mode IN ('SNAP','CORE'));

-- +goose Down
ALTER TABLE checkout_sessions DROP COLUMN authentication_url;
ALTER TABLE checkout_sessions DROP COLUMN checkout_mode;
