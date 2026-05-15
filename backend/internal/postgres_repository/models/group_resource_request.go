package models

import (
	"database/sql"
	"time"
)

type GroupResourceRequest struct {
	ID        int64
	GroupID   int64
	Type      string
	Content   sql.NullString
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}
