-- name: StoreRefreshToken :exec
INSERT INTO refresh_tokens (user_id, token_hash, device_id, ip, user_agent, expires_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetRefreshToken :one
SELECT user_id, token_hash, device_id, expires_at, created_at
FROM refresh_tokens
WHERE user_id = $1 AND token_hash = $2
LIMIT 1;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens
WHERE user_id = $1 AND token_hash = $2;

-- name: DeleteRefreshTokenByDevice :exec
DELETE FROM refresh_tokens
WHERE user_id = $1 AND device_id = $2;

-- name: DeleteRefreshTokensByUser :exec
DELETE FROM refresh_tokens
WHERE user_id = $1;
