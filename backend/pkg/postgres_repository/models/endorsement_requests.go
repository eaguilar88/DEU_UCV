package models

import "database/sql"

type EndorsementRequest struct {
	ID                int
	UserID            int
	UserFirstName     string
	UserLastName      string
	ReviewerID        int
	ReviewerFirstName string
	ReviewerLastName  string
	Type              string
	Name              sql.NullString
	Description       sql.NullString
	Status            string
	Comments          sql.NullString
	ReviewedAt        string
	CreatedAt         string
	UpdatedAt         string
}
