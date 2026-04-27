package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
	"github.com/eaguilar88/deu/internal/providers"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

func (r *PostgresRepository) GetProvider(ctx context.Context, providerID string) (entities.Provider, error) {
	query, args, err := queries.GetProviderByID(providerID, true, false).ToSql()
	if err != nil {
		return entities.Provider{}, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return entities.Provider{}, err
	}
	defer stmt.Close()
	var provider models.Provider
	row := stmt.QueryRowContext(ctx, args...)
	provider, err = scanProvider(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entities.Provider{}, fmt.Errorf("%w: %w", providers.ErrProviderNotFound, err)
		}
		return entities.Provider{}, err
	}
	return newProviderFromModel(provider), nil
}

func (r *PostgresRepository) GetProviderByCode(ctx context.Context, code string) (entities.Provider, error) {
	query, args, err := queries.GetProviderByCode(code).ToSql()
	if err != nil {
		return entities.Provider{}, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return entities.Provider{}, err
	}
	defer stmt.Close()
	var provider models.Provider
	row := stmt.QueryRowContext(ctx, args...)
	provider, err = scanProvider(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entities.Provider{}, fmt.Errorf("%w: %w", providers.ErrProviderNotFound, err)
		}
		return entities.Provider{}, err
	}
	return newProviderFromModel(provider), nil
}

func (r *PostgresRepository) GetProviderByUserID(ctx context.Context, userID string) (entities.Provider, error) {
	query, args, err := queries.GetProviderByUserID(userID).ToSql()
	if err != nil {
		return entities.Provider{}, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return entities.Provider{}, err
	}
	defer stmt.Close()
	var provider models.Provider
	row := stmt.QueryRowContext(ctx, args...)
	provider, err = scanProvider(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return entities.Provider{}, fmt.Errorf("%w: %w", providers.ErrProviderNotFound, err)
		}
		return entities.Provider{}, err
	}
	return newProviderFromModel(provider), nil
}

func (r *PostgresRepository) GetProviders(ctx context.Context, pageScope entities.PageScope, filters entities.ProviderFilters) ([]entities.Provider, entities.PageScope, error) {
	query, args, err := queries.GetProviders(filters, pageScope.PerPage, pageScope.Offset()).ToSql()
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
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
	query, args, err := queries.CreateProvider(newProviderModelFromEntities(provider)).ToSql()
	if err != nil {
		r.logger.Error("error creating query", zap.Error(err))
		return -1, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
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
			return -1, httperrors.NewDuplicateEntryError(err)
		}
		r.logger.Error("error inserting provider", zap.Error(err))
		return -1, httperrors.NewInternalError(err)
	}
	return lastInsertedID, nil
}

func (r *PostgresRepository) UpdateProvider(ctx context.Context, providerID string, provider entities.Provider) error {
	query, args, err := queries.UpdateProvider(newProviderModelFromEntities(provider)).ToSql()
	if err != nil {
		return httperrors.NewBadQueryError(err)
	}

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return httperrors.NewBadQueryError(err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}

	if affected, err := result.RowsAffected(); err != nil || affected == 0 {
		return fmt.Errorf("%w: %w", providers.ErrProviderNotFound, err)
	}

	return nil
}

func (r *PostgresRepository) DeleteProvider(ctx context.Context, providerID string) error {
	query, args, err := queries.DeleteProvider(providerID).ToSql()
	if err != nil {
		return httperrors.NewBadQueryError(err)
	}

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return httperrors.NewBadQueryError(err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}

	if affected, err := result.RowsAffected(); err != nil || affected == 0 {
		return fmt.Errorf("%w: %w", providers.ErrProviderNotFound, err)
	}

	return nil
}

func scanProvider(row scannable) (models.Provider, error) {
	var provider models.Provider
	err := row.Scan(
		&provider.ID,
		&provider.UserID,
		&provider.Name,
		&provider.PartyType,
		&provider.ProfitType,
		&provider.IsInternal,
		&provider.Bio,
		&provider.Code,
		&provider.IsActive,
		&provider.CreatedAt,
		&provider.UpdatedAt,
		&provider.DeletedAt,
		&provider.UserEmail,
		&provider.UserFirstName,
		&provider.UserLastName,
	)

	return provider, err
}

func newProviderFromModel(provider models.Provider) entities.Provider {
	p := entities.Provider{
		ID: provider.ID,
		User: entities.User{
			ID:        provider.UserID,
			Email:     provider.UserEmail,
			FirstName: provider.UserFirstName,
			LastName:  provider.UserLastName,
		},
		IsActive: provider.IsActive,
	}

	if provider.Name.Valid {
		p.Name = provider.Name.String
	}

	if provider.PartyType.Valid {
		p.PartyType = entities.ProviderPartyType(provider.PartyType.String)
	}

	p.ProfitType = entities.ProviderProfitType(provider.ProfitType)

	if provider.IsInternal.Valid {
		p.IsInternal = provider.IsInternal.Bool
	}

	if provider.Bio.Valid {
		p.Bio = provider.Bio.String
	}

	if provider.Code.Valid {
		p.Code = provider.Code.String
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
		ID:         provider.ID,
		UserID:     provider.User.ID,
		Name:       toNullString(provider.Name),
		PartyType:  toNullString(string(provider.PartyType)),
		ProfitType: string(provider.ProfitType),
		IsInternal: toNullBool(provider.IsInternal),
		Bio:        toNullString(provider.Bio),
		Code:       toNullString(provider.Code),
		IsActive:   provider.IsActive,
	}
}
