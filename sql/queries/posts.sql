-- name: CreatePost :one
INSERT INTO
posts (
    title,
    url,
    description,
    published_at,
    feed_id
)
VALUES (
    @title,
    @url,
    @description,
    @published_at,
    @feed_id
)
RETURNING
    *;

-- name: GetPostsForUser :many
SELECT *
FROM
    posts
WHERE
    EXISTS (
        SELECT 1
        FROM
            feed_follows
        WHERE
            feed_follows.user_id = @user_id
            AND
            feed_follows.feed_id = posts.feed_id
    )
ORDER BY
    published_at DESC
LIMIT $1;

-- name: ResetPosts :exec
DELETE FROM posts;
