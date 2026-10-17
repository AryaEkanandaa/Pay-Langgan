-- +goose Up
ALTER TABLE checkout_sessions DROP CONSTRAINT chk_checkout_sessions_status;
ALTER TABLE checkout_sessions ADD CONSTRAINT chk_checkout_sessions_status
  CHECK (status IN ('created','creating','active','completed','expired','failed','cancelled'));

ALTER TABLE trial_card_sessions DROP CONSTRAINT trial_card_sessions_status_check;
ALTER TABLE trial_card_sessions ADD CONSTRAINT trial_card_sessions_status_check
  CHECK (status IN ('created','creating','active','completed','expired','cancelled'));

-- +goose Down
UPDATE checkout_sessions SET status='expired' WHERE status='cancelled';
UPDATE trial_card_sessions SET status='expired' WHERE status='cancelled';
ALTER TABLE trial_card_sessions DROP CONSTRAINT trial_card_sessions_status_check;
ALTER TABLE trial_card_sessions ADD CONSTRAINT trial_card_sessions_status_check
  CHECK (status IN ('created','creating','active','completed','expired'));
ALTER TABLE checkout_sessions DROP CONSTRAINT chk_checkout_sessions_status;
ALTER TABLE checkout_sessions ADD CONSTRAINT chk_checkout_sessions_status
  CHECK (status IN ('created','creating','active','completed','expired','failed'));
