package repository

import (
	"context"

	"github.com/eaguilar88/deu/internal/entities"
	errs "github.com/eaguilar88/deu/internal/errors"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

func (r *PostgresRepository) GetProvider(ctx context.Context, providerID string) (entities.Provider, error) {
	query := queries.GetProviderByID(providerID, true, false)
	sql, args, err := query.ToSql()
	if err != nil {
		return entities.Provider{}, err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return entities.Provider{}, err
	}
	defer stmt.Close()
	var provider models.Provider
	row := stmt.QueryRowContext(ctx, args...)
	provider, err = scanProvider(row)
	if err != nil {
		return entities.Provider{}, err
	}
	return newProviderFromModel(provider), nil
}

func (r *PostgresRepository) GetProviderByCode(ctx context.Context, code string) (entities.Provider, error) {
	query := queries.GetProviderByCode(code)
	sql, args, err := query.ToSql()
	if err != nil {
		return entities.Provider{}, err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return entities.Provider{}, err
	}
	defer stmt.Close()
	var provider models.Provider
	row := stmt.QueryRowContext(ctx, args...)
	provider, err = scanProvider(row)
	if err != nil {
		return entities.Provider{}, err
	}
	return newProviderFromModel(provider), nil
}

func (r *PostgresRepository) GetProviderByUserID(ctx context.Context, userID string) (entities.Provider, error) {
	query := queries.GetProviderByUserID(userID)
	sql, args, err := query.ToSql()
	if err != nil {
		return entities.Provider{}, err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return entities.Provider{}, err
	}
	defer stmt.Close()
	var provider models.Provider
	row := stmt.QueryRowContext(ctx, args...)
	provider, err = scanProvider(row)
	if err != nil {
		return entities.Provider{}, err
	}
	return newProviderFromModel(provider), nil
}

func (r *PostgresRepository) GetProviders(ctx context.Context, pageScope entities.PageScope) ([]entities.Provider, entities.PageScope, error) {
	sql, args, err := queries.GetProviders(pageScope.PerPage, pageScope.Offset()).ToSql()
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	defer rows.Close()
	var providers []entities.Provider
	for rows.Next() {
		provider, err := scanProvider(rows)
		if err != nil {
			return nil, entities.PageScope{}, err
		}
		providers = append(providers, newProviderFromModel(provider))
	}
	pageScope.Count = len(providers)
	return providers, pageScope, nil
}

func (r *PostgresRepository) CreateProvider(ctx context.Context, provider entities.Provider) (int64, error) {
	sql, args, err := queries.CreateProvider(newProviderModelFromEntities(provider)).ToSql()
	if err != nil {
		r.logger.Error("error creating query", zap.Error(err))
		return -1, err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		r.logger.Error("error preparing query", zap.Error(err))
		return -1, err
	}
	defer stmt.Close()
	var lastInsertedID int64
	err = stmt.QueryRowContext(ctx, args...).Scan(&lastInsertedID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == pgErrorCodeUniqueViolation {
			r.logger.Error("error inserting provider", zap.Error(err))
			return -1, errs.NewDuplicateEntryError(err)
		}
		r.logger.Error("error inserting provider", zap.Error(err))
		return -1, errs.NewInternalError(err)
	}
	return lastInsertedID, nil
}

func (r *PostgresRepository) UpdateProvider(ctx context.Context, providerID string, provider entities.Provider) error {
	sql, args, err := queries.UpdateProvider(newProviderModelFromEntities(provider)).ToSql()
	if err != nil {
		return errs.NewBadQueryError(err)
	}

	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return errs.NewBadQueryError(err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}

	if affected, err := result.RowsAffected(); err != nil || affected == 0 {
		return errs.NewNotFoundError(err)
	}

	return nil
}

func (r *PostgresRepository) DeleteProvider(ctx context.Context, providerID string) error {
	sql, args, err := queries.DeleteProvider(providerID).ToSql()
	if err != nil {
		return errs.NewBadQueryError(err)
	}

	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return errs.NewBadQueryError(err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}

	if affected, err := result.RowsAffected(); err != nil || affected == 0 {
		return errs.NewNotFoundError(err)
	}

	return nil
}

func scanProvider(row scannable) (models.Provider, error) {
	var provider models.Provider
	err := row.Scan(
		&provider.ID,
		&provider.UserID,
		&provider.Code,
		&provider.Active,
		&provider.CreatedAt,
		&provider.UpdatedAt,
		&provider.DeletedAt,
	)

	return provider, err
}

func newProviderFromModel(provider models.Provider) entities.Provider {
	p := entities.Provider{
		ID: provider.ID,
		User: entities.User{
			ID:        provider.UserID,
			FirstName: provider.FirstName,
			LastName:  provider.LastName,
		},
		Code: provider.Code,
	}
	if provider.CreatedAt.Valid {
		p.CreatedAt = provider.CreatedAt.String
	}

	if provider.UpdatedAt.Valid {
		p.UpdatedAt = provider.UpdatedAt.String
	}

	if provider.DeletedAt.Valid {
		p.DeletedAt = provider.DeletedAt.String
	}
	return p
}

func newProviderModelFromEntities(provider entities.Provider) models.Provider {
	return models.Provider{
		Code:   provider.Code,
		UserID: provider.User.ID,
	}
}
