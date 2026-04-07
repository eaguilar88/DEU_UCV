package models

import "database/sql"

type Course struct {
	ID                string
	Name              string
	Description       sql.NullString
	OwnerID           string
	Objectives        sql.NullString
	Rationale         sql.NullString
	Duration          sql.NullString
	Cost              sql.NullString
	InstructorProfile sql.NullString
	Profiles          sql.NullString
	Requirements      sql.NullString
	Content           sql.NullString
	Evaluation        sql.NullString
	Schedule          sql.NullString
	Type              sql.NullString
	Faculty           sql.NullString
	Location          sql.NullString
	IsActive          bool
	CreatedAt         string
	UpdatedAt         string
	DeletedAt         sql.NullString
}
