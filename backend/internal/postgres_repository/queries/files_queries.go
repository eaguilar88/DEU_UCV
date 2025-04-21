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
	"f.metadata",
	"f.uploaded_by",
	"f.created_at",
	"f.updated_at",
	"f.deleted_at",
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
			"metadata",
			"uploaded_by",
		).
		Suffix("RETURNING id")
	for _, file := range files {
		queryBuilder = queryBuilder.Values(
			file.OwnerID,
			file.OwnerType,
			file.FileKey,
			marshalMetadata(file.Metadata),
			file.UploadedBy,
		)
	}
	return queryBuilder
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
