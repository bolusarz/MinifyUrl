-- name: CreateUser :one
INSERT INTO users (first_name, last_name, email, username, hashed_password)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserById :one
SELECT *
FROM users
WHERE id = $1
LIMIT 1;

-- name: GetUser :one
SELECT *
FROM users
WHERE username = sqlc.arg(userNameOrEmail) OR email = sqlc.arg(userNameOrEmail)
LIMIT 1;

-- name: UpdateUser :one
update users
set first_name = $1,
    last_name  = $2,
    email      = $3,
    username   = $4
where id = $5
returning *;