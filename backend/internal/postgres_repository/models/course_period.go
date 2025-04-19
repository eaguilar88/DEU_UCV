package models

import "database/sql"

type CoursePeriod struct {
	ID              string
	CourseID        string
	StartDate       string
	EndDate         string
	InscriptionDate string
	IsActive        bool
	CreatedAt       string
	UpdatedAt       string
	DeletedAt       sql.NullString
}
