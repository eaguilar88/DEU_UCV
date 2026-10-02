package models

import (
	"database/sql"
	"time"
)

type Job struct {
	ID          int64
	Kind        string
	Payload     []byte
	Status      string
	Attempts    int
	MaxAttempts int
	RunAt       time.Time
	LastError   sql.NullString
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
