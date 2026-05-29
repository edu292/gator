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
JOIN
  feeds
  ON
    feeds.id = feed_follows.feed_id 
JOIN
  users
  ON
    users.id = feed_follows.user_id
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
