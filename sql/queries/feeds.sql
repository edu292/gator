-- name: CreateFeed :one
INSERT INTO feeds (name, url, user_id)
VALUES (
    @name,
    @url,
    @user_id
)
RETURNING *;

-- name: GetFeeds :many
SELECT
  feeds.name AS feed_name,
  feeds.url,
  users.name AS user_name
FROM
  feeds
JOIN
  users ON users.id = feeds.user_id;

-- name: GetFeedByUrl :one
SELECT
  *
FROM
  feeds
WHERE
  url = @url;

-- name: MarkFeedFetched :exec
UPDATE
  feeds
SET
  updated_at=NOW(),
  last_fetched_at=NOW()
WHERE
  id = @id;

-- name: GetNextFeedToFetch :one
SELECT
  *
FROM
  feeds
ORDER BY
  last_fetched_at NULLS FIRST
LIMIT
  1;

-- name: ResetFeeds :exec
DELETE FROM feeds;
