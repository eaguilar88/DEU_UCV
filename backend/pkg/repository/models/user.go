package models

import "database/sql"

type User struct {
	ID             int
	IDType         string
	CI             int
	Username       string
	FirstName      string
	LastName       string
	DateOfBirth    string
	Gender         sql.NullString
	EducationLevel string
	Address        sql.NullString
	Password       string
	CreatedAt      string
	UpdatedAt      string
}
