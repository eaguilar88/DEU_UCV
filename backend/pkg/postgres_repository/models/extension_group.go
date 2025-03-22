package models

import "database/sql"

type ExtensionGroup struct {
	ID            int
	UserID        int
	EndorsementID int
	Name          string
	Description   sql.NullString
	Objective     sql.NullString
	Action        sql.NullString
	Reach         string
	Path          string
	CreatedAt     string
	UpdatedAt     string
}
