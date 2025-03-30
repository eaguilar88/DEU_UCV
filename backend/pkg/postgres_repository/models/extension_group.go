package models

import "database/sql"

type ExtensionGroup struct {
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
	Objective         sql.NullString
	Location          sql.NullString
	CreatedAt         string
	UpdatedAt         string
	DeletedAt         string
}
