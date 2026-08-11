package models

import "database/sql"

type Activity struct {
	ID                    string
	GroupID               string
	GroupName             sql.NullString
	Name                  sql.NullString
	Description           sql.NullString
	DateStart             sql.NullString
	DateEnd               sql.NullString
	Location              sql.NullString
	KnowledgeArea         sql.NullString
	Allies                sql.NullString
	GroupParticipants     sql.NullInt64
	EstimatedParticipants sql.NullInt64
	ActualParticipants    sql.NullInt64
	Financing             sql.NullString
	Comments              sql.NullString
	GalleryURL            sql.NullString
	ReportChecked         sql.NullBool
	IsFeatured            sql.NullBool
	CreatedAt             sql.NullString
	UpdatedAt             sql.NullString
	DeletedAt             sql.NullString
}
