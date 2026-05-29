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
