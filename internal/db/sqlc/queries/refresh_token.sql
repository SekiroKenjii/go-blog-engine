-- name: StoreRefreshToken :exec
INSERT INTO refresh_tokens (user_id, token_hash, device_id, ip, user_agent, expires_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE user_id = $1 AND token_hash = $2;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens
WHERE user_id = $1 AND token_hash = $2;

-- name: DeleteAllRefreshTokensForUser :exec
DELETE FROM refresh_tokens
WHERE user_id = $1;

-- name: DeleteRefreshTokenByDevice :exec
DELETE FROM refresh_tokens
WHERE user_id = $1 AND device_id = $2;
