package models

import "database/sql"

type Course struct {
	ID                string
	Name              string
	Description       sql.NullString
	OwnerID           string
	OwnerFirstName    string
	OwnerLastName     string
	EndorsementID     string
	EndorserID        string
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
