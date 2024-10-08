// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: session.sql

package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const blockSession = `-- name: BlockSession :one
UPDATE sessions
SET is_blocked = true
WHERE id = $1
RETURNING id, user_id, refresh_token, client_ip, user_agent, is_blocked, expires_at, created_at
`

func (q *Queries) BlockSession(ctx context.Context, id uuid.UUID) (Session, error) {
	row := q.db.QueryRow(ctx, blockSession, id)
	var i Session
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.RefreshToken,
		&i.ClientIp,
		&i.UserAgent,
		&i.IsBlocked,
		&i.ExpiresAt,
		&i.CreatedAt,
	)
	return i, err
}

const createSession = `-- name: CreateSession :one
INSERT INTO sessions (id, user_id, refresh_token, client_ip, user_agent, is_blocked, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, user_id, refresh_token, client_ip, user_agent, is_blocked, expires_at, created_at
`

type CreateSessionParams struct {
	ID           uuid.UUID `json:"id"`
	UserID       int64     `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	ClientIp     string    `json:"client_ip"`
	UserAgent    string    `json:"user_agent"`
	IsBlocked    bool      `json:"is_blocked"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func (q *Queries) CreateSession(ctx context.Context, arg CreateSessionParams) (Session, error) {
	row := q.db.QueryRow(ctx, createSession,
		arg.ID,
		arg.UserID,
		arg.RefreshToken,
		arg.ClientIp,
		arg.UserAgent,
		arg.IsBlocked,
		arg.ExpiresAt,
	)
	var i Session
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.RefreshToken,
		&i.ClientIp,
		&i.UserAgent,
		&i.IsBlocked,
		&i.ExpiresAt,
		&i.CreatedAt,
	)
	return i, err
}

const getActiveSessions = `-- name: GetActiveSessions :many
SELECT id, user_id, refresh_token, client_ip, user_agent, is_blocked, expires_at, created_at
FROM sessions
WHERE user_id = $1 AND expires_at > NOW() AND is_blocked = false
ORDER BY created_at DESC
LIMIT $2 OFFSET $3
`

type GetActiveSessionsParams struct {
	UserID int64 `json:"user_id"`
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) GetActiveSessions(ctx context.Context, arg GetActiveSessionsParams) ([]Session, error) {
	rows, err := q.db.Query(ctx, getActiveSessions, arg.UserID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Session{}
	for rows.Next() {
		var i Session
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.RefreshToken,
			&i.ClientIp,
			&i.UserAgent,
			&i.IsBlocked,
			&i.ExpiresAt,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getSession = `-- name: GetSession :one
SELECT id, user_id, refresh_token, client_ip, user_agent, is_blocked, expires_at, created_at
FROM sessions
WHERE id = $1
limit 1
`

func (q *Queries) GetSession(ctx context.Context, id uuid.UUID) (Session, error) {
	row := q.db.QueryRow(ctx, getSession, id)
	var i Session
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.RefreshToken,
		&i.ClientIp,
		&i.UserAgent,
		&i.IsBlocked,
		&i.ExpiresAt,
		&i.CreatedAt,
	)
	return i, err
}

const getSessions = `-- name: GetSessions :many
SELECT id, user_id, refresh_token, client_ip, user_agent, is_blocked, expires_at, created_at
FROM sessions
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3
`

type GetSessionsParams struct {
	UserID int64 `json:"user_id"`
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) GetSessions(ctx context.Context, arg GetSessionsParams) ([]Session, error) {
	rows, err := q.db.Query(ctx, getSessions, arg.UserID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Session{}
	for rows.Next() {
		var i Session
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.RefreshToken,
			&i.ClientIp,
			&i.UserAgent,
			&i.IsBlocked,
			&i.ExpiresAt,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
