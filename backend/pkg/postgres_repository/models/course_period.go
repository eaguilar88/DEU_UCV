package models

import "database/sql"

type CoursePeriod struct {
	ID              int
	CourseID        int
	StartDate       string
	EndDate         string
	InscriptionDate string
	IsActive        bool
	CreatedAt       string
	UpdatedAt       string
	DeletedAt       sql.NullString
}
