package models

import "database/sql"

type CoursePeriod struct {
	ID              string
	CourseID        string
	StartDate       string
	EndDate         string
	InscriptionDate string
	Capacity        int
	IsActive        bool
	ClosedAt        sql.NullString
	CreatedAt       string
	UpdatedAt       string
	DeletedAt       sql.NullString
}
