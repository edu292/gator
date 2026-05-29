-- name: CreateUser :one
INSERT INTO users (id, name)
VALUES (
    @id,
    @name
)
RETURNING *;

-- name: GetUser :one
SELECT
  *
FROM
  users
WHERE
  name = @name;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: GetUsers :many
SELECT
  *
FROM
  users;
