package repository

import (
	"context"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
)

func (r *PostgresRepository) SaveFilesToDB(ctx context.Context, files []entities.File) error {
	models := make([]models.File, 0, len(files))
	for _, f := range files {
		m := newFileFromEntity(f)
		models = append(models, m)
	}
	sql, args, err := queries.InsertFile(models).ToSql()
	if err != nil {
		return err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
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

func (r *PostgresRepository) GetFilesByOwner(ctx context.Context, ownerID string) ([]entities.File, error) {
	sql, args, err := queries.GetFilesByOwner(ownerID, string(entities.OwnerTypeProvider)).ToSql()
	if err != nil {
		return nil, err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := make([]entities.File, 0)

	for rows.Next() {
		file, err := r.scanFile(rows)
		if err != nil {
			return nil, err
		}
		files = append(files, newFileFromModel(file))
	}

	return files, nil
}

func (r *PostgresRepository) scanFile(rows scannable) (models.File, error) {
	var file models.File

	err := rows.Scan(
		&file.ID,
		&file.OwnerID,
		&file.OwnerType,
		&file.Public,
		&file.Metadata,
		&file.UploadedBy,
		&file.FileKey,
		&file.DeletedAt,
		&file.CreatedAt,
		&file.UpdatedAt,
	)
	return file, err
}

func newFileFromEntity(file entities.File) models.File {
	return models.File{
		ID:         file.ID,
		OwnerID:    file.OwnerID,
		OwnerType:  string(file.OwnerType),
		FileKey:    file.Key,
		Public:     file.Public,
		Metadata:   file.MetaData,
		UploadedBy: file.UploadedBy,
		DeletedAt:  file.DeletedAt,
		CreatedAt:  file.CreatedAt,
	}
}

func newFileFromModel(file models.File) entities.File {
	return entities.File{
		ID:         file.ID,
		OwnerID:    file.OwnerID,
		OwnerType:  entities.OwnerType(file.OwnerType),
		Public:     file.Public,
		MetaData:   file.Metadata,
		Key:        file.FileKey,
		UploadedBy: file.UploadedBy,
		CreatedAt:  file.CreatedAt,
		DeletedAt:  file.DeletedAt,
		UpdatedAt:  file.UpdatedAt,
	}
}
