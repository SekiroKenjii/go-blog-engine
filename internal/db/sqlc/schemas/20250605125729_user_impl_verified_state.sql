-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
ADD COLUMN IF NOT EXISTS is_verified BOOLEAN DEFAULT FALSE;

ALTER TABLE users ADD COLUMN IF NOT EXISTS locked_until TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
DROP COLUMN IF EXISTS is_verified;

ALTER TABLE users DROP COLUMN IF EXISTS locked_until;
-- +goose StatementEnd
