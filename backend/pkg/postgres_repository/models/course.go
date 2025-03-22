package models

import "database/sql"

type Course struct {
	ID                int
	Name              string
	Description       sql.NullString
	OwnerID           int
	OwnerFirstName    string
	OwnerLastName     string
	EndorsementID     int
	EndorserID        int
	EndorserFirstName string
	EndorserLastName  string
	Content           string
	Objectives        sql.NullString
	Cost              sql.NullFloat64
	Location          sql.NullString
	CreatedAt         string
	UpdatedAt         string
	DeletedAt         string
}
