-- name: CreateFeedFollow :one
WITH follow as (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES (
        $1,
        $2,
        $3,
        $4,
        $5
    )
    RETURNING *
)
SELECT f.*, u.name as user_name, fe.name as feed_name
FROM follow f
JOIN users u ON f.user_id = u.id
JOIN feeds fe ON f.feed_id = fe.id;

-- name: GetFeedFollowsForUser :many
SELECT f.name AS feed_name, u.name AS user_name
FROM feed_follows ff
JOIN feeds f ON ff.feed_id = f.id
JOIN users u ON ff.user_id = u.id
WHERE ff.user_id = $1;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
WHERE user_id = $1 AND feed_id = $2;
