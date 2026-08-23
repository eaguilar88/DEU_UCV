package repository

import (
	"context"
	"fmt"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
	"go.uber.org/zap"
)

func (r *PostgresRepository) HasInitialContract(ctx context.Context, providerID string) (bool, error) {
	query, args, err := queries.CountInitialContracts(providerID).ToSql()
	if err != nil {
		return false, err
	}
	var count int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PostgresRepository) GetUncoveredCourseIDs(ctx context.Context, providerID string) ([]string, error) {
	query, args, err := queries.GetUncoveredCourseIDs(providerID).ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}

// CreateProviderContract inserts a provider_contracts row, links every covered
// course via provider_contract_courses, and flips has_documentation=true on
// those courses, all within a single transaction.
func (r *PostgresRepository) CreateProviderContract(ctx context.Context, contract entities.ProviderContract, coveredCourseIDs []string) (entities.ProviderContract, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("failed to begin transaction", zap.Error(err))
		return entities.ProviderContract{}, err
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				r.logger.Error("failed to rollback transaction", zap.Error(rbErr))
			}
		}
	}()

	insertSQL, insertArgs, err := queries.InsertProviderContract(contract.ProviderID, string(contract.Type)).ToSql()
	if err != nil {
		return entities.ProviderContract{}, err
	}
	var contractID int64
	if err = tx.QueryRowContext(ctx, insertSQL, insertArgs...).Scan(&contractID); err != nil {
		r.logger.Error("failed to insert provider contract", zap.Error(err))
		return entities.ProviderContract{}, err
	}
	contractIDStr := fmt.Sprintf("%d", contractID)

	if len(coveredCourseIDs) > 0 {
		coursesSQL, coursesArgs, err := queries.InsertProviderContractCourses(contractIDStr, coveredCourseIDs).ToSql()
		if err != nil {
			return entities.ProviderContract{}, err
		}
		if _, err = tx.ExecContext(ctx, coursesSQL, coursesArgs...); err != nil {
			r.logger.Error("failed to insert provider contract courses", zap.Error(err))
			return entities.ProviderContract{}, err
		}

		markSQL, markArgs, err := queries.SetCoursesDocumented(coveredCourseIDs).ToSql()
		if err != nil {
			return entities.ProviderContract{}, err
		}
		if _, err = tx.ExecContext(ctx, markSQL, markArgs...); err != nil {
			r.logger.Error("failed to mark courses as documented", zap.Error(err))
			return entities.ProviderContract{}, err
		}
	}

	if err = tx.Commit(); err != nil {
		r.logger.Error("failed to commit transaction", zap.Error(err))
		return entities.ProviderContract{}, err
	}

	contract.ID = contractIDStr
	contract.CoveredCourses = coveredCourseIDs
	return contract, nil
}

func (r *PostgresRepository) GetProviderContracts(ctx context.Context, providerID string) ([]entities.ProviderContract, error) {
	query, args, err := queries.GetProviderContracts(providerID).ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contracts []models.ProviderContract
	for rows.Next() {
		var c models.ProviderContract
		if err := rows.Scan(&c.ID, &c.ProviderID, &c.Type, &c.CreatedAt); err != nil {
			return nil, err
		}
		contracts = append(contracts, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]entities.ProviderContract, 0, len(contracts))
	for _, c := range contracts {
		courseIDs, err := r.getContractCourseIDs(ctx, c.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, entities.ProviderContract{
			ID:             c.ID,
			ProviderID:     c.ProviderID,
			Type:           entities.ContractType(c.Type),
			CreatedAt:      c.CreatedAt,
			CoveredCourses: courseIDs,
		})
	}
	return result, nil
}

func (r *PostgresRepository) getContractCourseIDs(ctx context.Context, contractID string) ([]string, error) {
	query, args, err := queries.GetContractCourseIDs(contractID).ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
