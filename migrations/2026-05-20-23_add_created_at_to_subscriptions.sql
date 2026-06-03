-- +goose Up
ALTER TABLE subscriptions ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT NOW();

-- +goose Down
ALTER TABLE subscriptions DROP COLUMN created_at;
