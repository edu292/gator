-- name: CreateUser :one
INSERT INTO users (id, name)
VALUES (
    @id,
    @name
)
RETURNING *;

-- name: GetUserByID :one
SELECT
  *
FROM
  users
WHERE
  id = @id;

-- name: GetUserByName :one
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
