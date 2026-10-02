package models

import "database/sql"

type CoursePeriod struct {
	ID       string
	CourseID string
	Name     string
	// The dates are optional: an empty one is stored as NULL.
	StartDate       sql.NullString
	EndDate         sql.NullString
	InscriptionDate sql.NullString
	Capacity        int
	IsActive        bool
	ClosedAt        sql.NullString
	CreatedAt       string
	UpdatedAt       string
	DeletedAt       sql.NullString
}
