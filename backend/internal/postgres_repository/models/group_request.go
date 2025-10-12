package models

import (
	"database/sql"
	"time"
)

type GroupAuthRequest struct {
	ID         int64          `db:"id"`
	GroupID    int64          `db:"group_id"`
	Status     string         `db:"status"`
	Faculty    string         `db:"faculty"`
	ReviewerID sql.NullInt64  `db:"reviewer_id"`
	ReviewedAt sql.NullString `db:"reviewed_at"`
	Comments   string         `db:"comments"`
	CreatedAt  time.Time      `db:"created_at"`
	UpdatedAt  time.Time      `db:"updated_at"`
}
