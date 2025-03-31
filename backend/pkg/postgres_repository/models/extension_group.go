package models

import "database/sql"

type ExtensionGroup struct {
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
	Objective         sql.NullString
	Location          sql.NullString
	IsActive          bool
	CreatedAt         string
	UpdatedAt         string
	DeletedAt         string
}
