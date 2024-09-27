-- name: CreateSession :one
INSERT INTO sessions (id, user_id, refresh_token, client_ip, user_agent, is_blocked, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetSessions :many
SELECT *
FROM sessions
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetActiveSessions :many
SELECT *
FROM sessions
WHERE user_id = $1 AND expires_at > NOW() AND is_blocked = false
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetSession :one
SELECT *
FROM sessions
WHERE id = $1
limit 1;

-- name: BlockSession :one
UPDATE sessions
SET is_blocked = true
WHERE id = $1
RETURNING *;