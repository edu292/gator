-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO
      feed_follows (
        user_id,
        feed_id
      )
    VALUES (
        @user_id,
        @feed_id
    )
    RETURNING *
)

SELECT
  inserted_feed_follow.*,
  feeds.name AS feed_name
FROM
  inserted_feed_follow
JOIN
  feeds
  ON
    feeds.id = inserted_feed_follow.feed_id;

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
