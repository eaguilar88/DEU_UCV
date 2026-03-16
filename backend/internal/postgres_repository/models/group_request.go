package models

import (
	"database/sql"
	"time"
)

type GroupRequest struct {
	ID         int64          `db:"id"`
	GroupID    int64          `db:"group_id"`
	Status     string         `db:"status"`
	Faculty    string         `db:"faculty"`
	ReviewerID sql.NullInt64  `db:"reviewer_id"`
	ReviewedAt sql.NullString `db:"reviewed_at"`
	Comments   sql.NullString `db:"comments"`
	CreatedAt  time.Time      `db:"created_at"`
	UpdatedAt  time.Time      `db:"updated_at"`
}

func (m *GroupRequest) GetComments() string {
	if m.Comments.Valid {
		return m.Comments.String
	}
	return ""
}
