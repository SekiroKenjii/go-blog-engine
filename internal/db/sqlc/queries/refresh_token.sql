-- name: StoreRefreshToken :exec
INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3);

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE user_id = $1 AND token_hash = $2;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens
WHERE user_id = $1 AND token_hash = $2;

-- name: DeleteAllRefreshTokensForUser :exec
DELETE FROM refresh_tokens
WHERE user_id = $1;
