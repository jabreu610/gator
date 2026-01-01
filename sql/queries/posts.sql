-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: GetPostsForUser :many
SELECT p.*
FROM posts p
WHERE p.feed_id IN (
    SELECT f.id
    FROM feed_follows ff
    JOIN feeds f ON ff.feed_id = f.id
    WHERE ff.user_id = $1
)
ORDER BY p.published_at DESC
LIMIT $2;
