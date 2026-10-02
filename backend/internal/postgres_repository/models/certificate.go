package models

import (
	"database/sql"
	"time"
)

type Certificate struct {
	ID               int64
	CloseRequestID   int64
	CourseCycleID    int64
	FirstName        string
	LastName         string
	Document         string
	Email            sql.NullString
	VerificationCode string
	StorageKey       sql.NullString
	IssuedAt         sql.NullTime
	RevokedAt        sql.NullTime
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
