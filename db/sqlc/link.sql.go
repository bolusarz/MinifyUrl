// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: link.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createLink = `-- name: CreateLink :one
INSERT INTO links (code, link, user_id)
VALUES ($1, $2, $3)
RETURNING id, user_id, code, link, created_at, active
`

type CreateLinkParams struct {
	Code   string `json:"code"`
	Link   string `json:"link"`
	UserID int64  `json:"user_id"`
}

func (q *Queries) CreateLink(ctx context.Context, arg CreateLinkParams) (Link, error) {
	row := q.db.QueryRow(ctx, createLink, arg.Code, arg.Link, arg.UserID)
	var i Link
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Code,
		&i.Link,
		&i.CreatedAt,
		&i.Active,
	)
	return i, err
}

const getLinkByCode = `-- name: GetLinkByCode :one
select id, user_id, code, link, created_at, active
from links
where code = $1
limit 1
`

func (q *Queries) GetLinkByCode(ctx context.Context, code string) (Link, error) {
	row := q.db.QueryRow(ctx, getLinkByCode, code)
	var i Link
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Code,
		&i.Link,
		&i.CreatedAt,
		&i.Active,
	)
	return i, err
}

const getLinkById = `-- name: GetLinkById :one
select id, user_id, code, link, created_at, active
from links
where id = $1
limit 1
`

func (q *Queries) GetLinkById(ctx context.Context, id int64) (Link, error) {
	row := q.db.QueryRow(ctx, getLinkById, id)
	var i Link
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Code,
		&i.Link,
		&i.CreatedAt,
		&i.Active,
	)
	return i, err
}

const getLinks = `-- name: GetLinks :many
SELECT id, user_id, code, link, created_at, active
FROM links
ORDER BY id
LIMIT $1 OFFSET $2
`

type GetLinksParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) GetLinks(ctx context.Context, arg GetLinksParams) ([]Link, error) {
	rows, err := q.db.Query(ctx, getLinks, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Link{}
	for rows.Next() {
		var i Link
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Code,
			&i.Link,
			&i.CreatedAt,
			&i.Active,
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

const getLinksByUser = `-- name: GetLinksByUser :many
select id, user_id, code, link, created_at, active
from links
where user_id = $1
order by id desc
limit $2 offset $3
`

type GetLinksByUserParams struct {
	UserID int64 `json:"user_id"`
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) GetLinksByUser(ctx context.Context, arg GetLinksByUserParams) ([]Link, error) {
	rows, err := q.db.Query(ctx, getLinksByUser, arg.UserID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Link{}
	for rows.Next() {
		var i Link
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Code,
			&i.Link,
			&i.CreatedAt,
			&i.Active,
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

const toggleStatus = `-- name: ToggleStatus :one
update links
set active = $1
where id = $2
returning id, user_id, code, link, created_at, active
`

type ToggleStatusParams struct {
	Active pgtype.Bool `json:"active"`
	ID     int64       `json:"id"`
}

func (q *Queries) ToggleStatus(ctx context.Context, arg ToggleStatusParams) (Link, error) {
	row := q.db.QueryRow(ctx, toggleStatus, arg.Active, arg.ID)
	var i Link
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Code,
		&i.Link,
		&i.CreatedAt,
		&i.Active,
	)
	return i, err
}

const updateCode = `-- name: UpdateCode :one
update links
set code = $1
where id = $2
returning id, user_id, code, link, created_at, active
`

type UpdateCodeParams struct {
	Code string `json:"code"`
	ID   int64  `json:"id"`
}

func (q *Queries) UpdateCode(ctx context.Context, arg UpdateCodeParams) (Link, error) {
	row := q.db.QueryRow(ctx, updateCode, arg.Code, arg.ID)
	var i Link
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Code,
		&i.Link,
		&i.CreatedAt,
		&i.Active,
	)
	return i, err
}
