package models

import "database/sql"

type User struct {
	ID             string
	CI             int
	Email          string
	FirstName      string
	LastName       string
	DateOfBirth    sql.NullString
	Gender         sql.NullString
	EducationLevel string
	Address        sql.NullString
	ProviderCode   sql.NullString
	Password       string
	CreatedAt      string
	UpdatedAt      string
}

type Role struct {
	ID   int
	Name string
}
