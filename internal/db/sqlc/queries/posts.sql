-- name: CreatePost :one
INSERT INTO posts (id, user_id, title, content, image_url, published)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListPublishedPost :many
SELECT id, user_id, title, content, image_url, published, created_at, published_at, updated_at FROM posts
WHERE published = TRUE
ORDER BY published_at DESC
LIMIT $1 OFFSET $2;

-- name: GetPostByID :one
SELECT id, user_id, title, content, image_url, published, created_at, published_at, updated_at FROM posts
WHERE id = $1
LIMIT 1;

-- name: PublishPost :exec
UPDATE posts
SET published = TRUE, published_at = CURRENT_TIMESTAMP
WHERE id = $1;
