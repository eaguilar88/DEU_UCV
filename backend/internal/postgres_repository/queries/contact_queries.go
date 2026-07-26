package queries

import sq "github.com/Masterminds/squirrel"

func InsertContact(contactType, value, ownerType string, ownerID int64) sq.InsertBuilder {
	return psql.Insert(contactsTableName).
		Columns(
			"contact_type",
			"contact_value",
			"owner_type",
			"owner_id",
			"created_at",
			"updated_at",
		).Values(
		contactType,
		value,
		ownerType,
		ownerID,
		sq.Expr("NOW()"),
		sq.Expr("NOW()"),
	).Suffix("RETURNING id")
}
