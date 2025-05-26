-- +goose Up
-- +goose StatementBegin
ALTER TABLE refresh_tokens
ADD COLUMN IF NOT EXISTS device_id TEXT NOT NULL DEFAULT 'unknown';

ALTER TABLE refresh_tokens
ADD COLUMN IF NOT EXISTS ip TEXT;

ALTER TABLE refresh_tokens
ADD COLUMN IF NOT EXISTS user_agent TEXT;

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_device ON refresh_tokens (user_id, device_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE refresh_tokens
DROP COLUMN IF EXISTS device_id;

ALTER TABLE refresh_tokens
DROP COLUMN IF EXISTS ip;

ALTER TABLE refresh_tokens
DROP COLUMN IF EXISTS user_agent;

DROP INDEX IF EXISTS idx_refresh_tokens_user_device;
-- +goose StatementEnd
