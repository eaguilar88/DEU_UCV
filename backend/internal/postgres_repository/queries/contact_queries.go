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

// UpsertContact inserts a contact, or updates its value if the (type, value, owner) combination already exists.
func UpsertContact(contactType, value, ownerType string, ownerID int64) sq.InsertBuilder {
	return psql.Insert(contactsTableName).
		Columns("contact_type", "contact_value", "owner_type", "owner_id").
		Values(contactType, value, ownerType, ownerID).
		Suffix("ON CONFLICT (contact_type, contact_value, owner_type, owner_id) DO UPDATE SET contact_value = EXCLUDED.contact_value, updated_at = NOW()")
}

func SelectContactsByOwner(ownerID, ownerType string) sq.SelectBuilder {
	return psql.Select("id", "contact_type", "contact_value", "owner_type", "owner_id").
		From(contactsTableName).
		Where(sq.Eq{"owner_id": ownerID, "owner_type": ownerType, "deleted_at": nil})
}
