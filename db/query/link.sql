-- name: CreateLink :one
INSERT INTO links (code, link, user_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetLinks :many
SELECT *
FROM links
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: GetLinksByUser :many
select *
from links
where user_id = $1
order by id desc
limit $2 offset $3;

-- name: GetLinkById :one
select *
from links
where id = $1
limit 1;

-- name: GetLinkByCode :one
select *
from links
where code = $1
limit 1;

-- name: UpdateCode :one
update links
set code = $1
where id = $2
returning *;

-- name: ToggleStatus :one
update links
set active = $1
where id = $2
returning *;