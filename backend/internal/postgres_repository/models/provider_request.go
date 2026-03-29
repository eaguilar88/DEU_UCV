package models

import "database/sql"

type ProviderRequest struct {
	ID         int64
	ProviderID int64
	Status     string
	ReviewerID sql.NullInt64
	Comments   sql.NullString
	ReviewedAt sql.NullString
	CreatedAt  string
	UpdatedAt  string
}
