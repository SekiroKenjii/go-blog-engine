-- name: CreateUser :one
INSERT INTO users (id, email, first_name, last_name, password_hash)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByEmail :one
SELECT id, email, first_name, last_name, password_hash FROM users
WHERE email = $1
LIMIT 1;

-- name: GetUserByID :one
SELECT id, email, first_name, last_name, password_hash FROM users
WHERE id = $1
LIMIT 1;

-- name: MarkUserVerified :exec
UPDATE users SET is_verified = TRUE WHERE id = $1;
