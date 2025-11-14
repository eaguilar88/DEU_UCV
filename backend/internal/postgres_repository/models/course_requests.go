package models

import "database/sql"

type CourseRequest struct {
	ID         string
	CourseID   string
	Status     string
	ReviewerID sql.NullString
	Comments   sql.NullString
	ReviewedAt sql.NullString
	CreatedAt  string
	UpdatedAt  string
	DeletedAt  sql.NullString
	Course     *Course // Associated course information
}
