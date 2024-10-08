// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Link struct {
	ID        int64       `json:"id"`
	UserID    int64       `json:"user_id"`
	Code      string      `json:"code"`
	Link      string      `json:"link"`
	CreatedAt time.Time   `json:"created_at"`
	Active    pgtype.Bool `json:"active"`
}

type Session struct {
	ID           uuid.UUID `json:"id"`
	UserID       int64     `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	ClientIp     string    `json:"client_ip"`
	UserAgent    string    `json:"user_agent"`
	IsBlocked    bool      `json:"is_blocked"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type User struct {
	ID                int64            `json:"id"`
	Username          string           `json:"username"`
	HashedPassword    string           `json:"hashed_password"`
	FirstName         string           `json:"first_name"`
	LastName          string           `json:"last_name"`
	Email             string           `json:"email"`
	PasswordChangedAt pgtype.Timestamp `json:"password_changed_at"`
	CreatedAt         time.Time        `json:"created_at"`
}
