package models

import "database/sql"

type Course struct {
	ID            int
	UserID        int
	EndorsementID int
	EndorsedBy    string
	Objectives    sql.NullString
	Cost          sql.NullFloat64
	Location      sql.NullString
	Content       string
	CreatedAt     string
	UpdatedAt     string
}
