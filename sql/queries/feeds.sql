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
  feeds.name, feeds.url, users.name
FROM
  feeds
JOIN
  users ON users.id = feeds.user_id;
