package queries

import (
	"encoding/json"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
)

var fileQuerySelectCommon = []string{
	"f.id",
	"f.owner_id",
	"f.owner_type",
	"f.file_key",
	"f.public",
	"f.purpose",
	"f.metadata",
	"f.uploaded_by",
	"f.created_at",
	"f.deleted_at",
	"f.version",
}

func GetFileByID(fileID string) sq.SelectBuilder {
	return psql.Select(fileQuerySelectCommon...).
		From(fmt.Sprintf("%s AS f", filesTableName)).
		Where(sq.Eq{"f.id": fileID})
}

func GetFilesByOwner(ownerID, ownerType string) sq.SelectBuilder {
	return psql.Select(fileQuerySelectCommon...).
		From(fmt.Sprintf("%s AS f", filesTableName)).
		Where(sq.Eq{"f.owner_id": ownerID, "f.owner_type": ownerType})
}

func InsertFile(files []models.File) sq.InsertBuilder {
	queryBuilder := psql.Insert(filesTableName).
		Columns(
			"owner_id",
			"owner_type",
			"file_key",
			"purpose",
			"version",
			"metadata",
			"uploaded_by",
		).
		Suffix("RETURNING id")
	for _, file := range files {
		queryBuilder = queryBuilder.Values(
			file.OwnerID,
			file.OwnerType,
			file.FileKey,
			file.Purpose,
			file.Version,
			marshalMetadata(file.Metadata),
			file.UploadedBy,
		)
	}
	return queryBuilder
}

func GetFilesByOwnerAndPurpose(ownerID, ownerType, purpose string) sq.SelectBuilder {
	return psql.Select(fileQuerySelectCommon...).
		From(fmt.Sprintf("%s AS f", filesTableName)).
		Where(sq.Eq{"f.owner_id": ownerID, "f.owner_type": ownerType, "f.purpose": purpose}).
		Where(sq.Eq{"f.deleted_at": nil}).
		OrderBy("f.version ASC")
}

func marshalMetadata(metadata map[string]string) string {
	if metadata == nil {
		return "{}"
	}
	bytes, err := json.Marshal(metadata)
	if err != nil {
		// Handle error as you see fit (panic/log/fallback)
		return "{}"
	}
	return string(bytes)
}
