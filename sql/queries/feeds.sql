-- name: CreateFeed :one
INSERT INTO feeds (name, url, user_id)
VALUES (
    @name,
    @url,
    @user_id
)
RETURNING *;
