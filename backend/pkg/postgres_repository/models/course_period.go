package models

import "database/sql"

type CoursePeriod struct {
	ID              int
	Course          Course
	StartDate       string
	EndDate         string
	InscriptionDate string
	IsActive        bool
	CreatedAt       string
	UpdatedAt       string
	DeletedAt       sql.NullString
}
