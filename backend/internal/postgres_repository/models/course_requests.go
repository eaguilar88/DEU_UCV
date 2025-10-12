package models

import "database/sql"

type CourseRequest struct {
	ID                string
	UserID            string
	UserFirstName     string
	UserLastName      string
	ReviewerID        string
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
