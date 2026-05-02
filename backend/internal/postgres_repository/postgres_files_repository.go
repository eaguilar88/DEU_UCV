package repository

import (
	"context"
	"encoding/json"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
)

func (r *PostgresRepository) SaveFilesToDB(ctx context.Context, files []*entities.File) error {
	models := make([]models.File, 0, len(files))
	for _, f := range files {
		m := newFileFromEntity(*f)
		models = append(models, m)
	}
	query, args, err := queries.InsertFile(models).ToSql()
	if err != nil {
		return err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepository) GetFilesByOwner(ctx context.Context, ownerID string, ownerType entities.OwnerType) (entities.GroupedFiles, error) {
	query, args, err := queries.GetFilesByOwner(ownerID, ownerType.String()).ToSql()
	if err != nil {
		return nil, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := make(entities.GroupedFiles)

	for rows.Next() {
		file, err := r.scanFile(rows)
		if err != nil {
			return nil, err
		}
		files[file.Purpose] = append(files[file.Purpose], newFileFromModel(file))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

func (r *PostgresRepository) scanFile(rows scannable) (models.File, error) {
	var file models.File
	var metadataBytes []byte

	err := rows.Scan(
		&file.ID,
		&file.OwnerID,
		&file.OwnerType,
		&file.FileKey,
		&file.Public,
		&file.Purpose,
		&metadataBytes,
		&file.UploadedBy,
		&file.CreatedAt,
		&file.DeletedAt,
		&file.Version,
	)
	if err != nil {
		return file, err
	}
	var md map[string]string
	if err := json.Unmarshal(metadataBytes, &md); err != nil {
		return file, err
	}
	file.Metadata = md
	return file, nil
}

func newFileFromEntity(file entities.File) models.File {
	model := models.File{
		ID:         file.ID,
		OwnerID:    file.OwnerID,
		OwnerType:  string(file.OwnerType),
		FileKey:    file.Key,
		Public:     file.Public,
		Purpose:    file.Purpose,
		Version:    file.Version,
		Metadata:   file.MetaData,
		UploadedBy: file.UploadedBy,
		CreatedAt:  file.CreatedAt,
	}
	if file.DeletedAt != "" {
		model.DeletedAt.Valid = true
		model.DeletedAt.String = file.DeletedAt
	}
	return model
}

func newFileFromModel(file models.File) *entities.File {
	model := &entities.File{
		ID:         file.ID,
		OwnerID:    file.OwnerID,
		OwnerType:  entities.OwnerType(file.OwnerType),
		Public:     file.Public,
		Purpose:    file.Purpose,
		Version:    file.Version,
		MetaData:   file.Metadata,
		Key:        file.FileKey,
		UploadedBy: file.UploadedBy,
		CreatedAt:  file.CreatedAt,
	}

	if file.DeletedAt.Valid {
		model.DeletedAt = file.DeletedAt.String
	}

	return model
}

func (r *PostgresRepository) GetFilesByOwnerAndPurpose(ctx context.Context, ownerID, ownerType, purpose string) ([]*entities.File, error) {
	query, args, err := queries.GetFilesByOwnerAndPurpose(ownerID, ownerType, purpose).ToSql()
	if err != nil {
		return nil, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*entities.File
	for rows.Next() {
		file, err := r.scanFile(rows)
		if err != nil {
			return nil, err
		}
		files = append(files, newFileFromModel(file))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if files == nil {
		files = []*entities.File{}
	}
	return files, nil
}
