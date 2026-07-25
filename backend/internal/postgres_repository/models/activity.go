package models

import "database/sql"

type Activity struct {
	ID                    string
	GroupID               string
	Name                  sql.NullString
	Description           sql.NullString
	Date                  sql.NullString
	KnowledgeArea         sql.NullString
	Allies                sql.NullString
	EstimatedParticipants sql.NullInt64
	ActualParticipants    sql.NullInt64
	Financing             sql.NullString
	Comments              sql.NullString
	GalleryURL            sql.NullString
	CreatedAt             sql.NullString
	UpdatedAt             sql.NullString
	DeletedAt             sql.NullString
}
