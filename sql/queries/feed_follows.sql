-- name: CreateFeedFollow :one
INSERT INTO
feed_follows (
    user_id,
    feed_id
)
VALUES (
    @user_id,
    @feed_id
)
RETURNING *;

-- name: GetFeedFollowsForUser :many
SELECT
    feed_follows.*,
    feeds.name AS feed_name,
    feeds.url AS feed_url
FROM
    feed_follows
INNER JOIN
    feeds
    ON
    feed_follows.feed_id = feeds.id
INNER JOIN
    users
    ON
    feed_follows.user_id = users.id
WHERE
    feed_follows.user_id = @user_id;

-- name: UnfollowFeed :one
DELETE FROM
feed_follows
WHERE
    feed_follows.user_id = @user_id
    AND
    feed_follows.feed_id = @feed_id
RETURNING *;

-- name: ResetFeedFollows :exec
DELETE FROM feed_follows;

