package models

import "database/sql"

type CourseCycleCloseRequest struct {
	ID            int64
	CourseCycleID int64
	SubmittedBy   int64
	Status        string
	Comments      sql.NullString
	ReviewerID    sql.NullInt64
	ReviewedAt    sql.NullString
	CreatedAt     string
	UpdatedAt     string
}
