package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/eaguilar88/deu/internal/activities"
	"github.com/eaguilar88/deu/internal/entities"

	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
	"go.uber.org/zap"
)

func (r *PostgresRepository) GetActivityByID(ctx context.Context, id string) (entities.Activity, error) {
	query, args, err := queries.GetActivityByID(id).ToSql()
	if err != nil {
		return entities.Activity{}, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return entities.Activity{}, err
	}
	defer stmt.Close()
	row := stmt.QueryRowContext(ctx, args...)
	a, err := scanActivity(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return entities.Activity{}, fmt.Errorf("%w: %w", activities.ErrActivityNotFound, err)
		}
		return entities.Activity{}, err
	}
	return newActivityFromModel(a), nil
}

func (r *PostgresRepository) GetActivities(ctx context.Context, filter entities.ActivityFilter, pageScope entities.PageScope) ([]entities.Activity, entities.PageScope, error) {
	countQuery, countArgs, err := queries.CountActivities(filter).ToSql()
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	countStmt, err := r.db.PrepareContext(ctx, countQuery)
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	defer countStmt.Close()
	if err := countStmt.QueryRowContext(ctx, countArgs...).Scan(&pageScope.Count); err != nil {
		return nil, entities.PageScope{}, err
	}

	query, args, err := queries.GetActivities(filter, pageScope.PerPage, pageScope.Offset()).ToSql()
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

	var result []entities.Activity
	for rows.Next() {
		a, err := scanActivity(rows)
		if err != nil {
			return nil, entities.PageScope{}, err
		}
		result = append(result, newActivityFromModel(a))
	}
	if err := rows.Err(); err != nil {
		return nil, entities.PageScope{}, err
	}
	return result, pageScope, nil
}

func (r *PostgresRepository) CreateActivity(ctx context.Context, a entities.Activity) (int64, error) {
	query, args, err := queries.InsertActivity(newActivityToModel(a)).ToSql()
	if err != nil {
		return 0, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	var lastInsertedID int64
	if err := stmt.QueryRowContext(ctx, args...).Scan(&lastInsertedID); err != nil {
		r.logger.Error("error inserting activity", zap.Error(err))
		return -1, err
	}
	return lastInsertedID, nil
}

func (r *PostgresRepository) UpdateActivity(ctx context.Context, a entities.Activity) error {
	query, args, err := queries.UpdateActivity(newActivityToModel(a)).ToSql()
	if err != nil {
		return err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err != nil || affected == 0 {
		return fmt.Errorf("%w", activities.ErrActivityNotFound)
	}
	return nil
}

func (r *PostgresRepository) DeleteActivity(ctx context.Context, id string) error {
	query, args, err := queries.SoftDeleteActivity(id).ToSql()
	if err != nil {
		return err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err != nil || affected == 0 {
		return fmt.Errorf("%w", activities.ErrActivityNotFound)
	}
	return nil
}

func scanActivity(row scannable) (models.Activity, error) {
	var a models.Activity
	err := row.Scan(
		&a.ID,
		&a.GroupID,
		&a.Name,
		&a.Description,
		&a.Date,
		&a.KnowledgeArea,
		&a.Allies,
		&a.EstimatedParticipants,
		&a.ActualParticipants,
		&a.Financing,
		&a.Comments,
		&a.CreatedAt,
		&a.UpdatedAt,
		&a.DeletedAt,
	)
	if err != nil {
		return models.Activity{}, err
	}
	return a, nil
}

func newActivityToModel(a entities.Activity) models.Activity {
	return models.Activity{
		ID:                    a.ID,
		GroupID:               a.GroupID,
		Name:                  toNullString(a.Name),
		Description:           toNullString(a.Description),
		Date:                  toNullString(a.Date),
		KnowledgeArea:         toNullString(a.KnowledgeArea),
		Allies:                toNullString(a.Allies),
		EstimatedParticipants: sql.NullInt64{Int64: int64(a.EstimatedParticipants), Valid: a.EstimatedParticipants != 0},
		ActualParticipants:    sql.NullInt64{Int64: int64(a.ActualParticipants), Valid: a.ActualParticipants != 0},
		Financing:             toNullString(a.Financing),
		Comments:              toNullString(a.Comments),
	}
}

func newActivityFromModel(a models.Activity) entities.Activity {
	return entities.Activity{
		ID:                    a.ID,
		GroupID:               a.GroupID,
		Name:                  a.Name.String,
		Description:           a.Description.String,
		Date:                  a.Date.String,
		KnowledgeArea:         a.KnowledgeArea.String,
		Allies:                a.Allies.String,
		EstimatedParticipants: int(a.EstimatedParticipants.Int64),
		ActualParticipants:    int(a.ActualParticipants.Int64),
		Financing:             a.Financing.String,
		Comments:              a.Comments.String,
		CreatedAt:             a.CreatedAt.String,
		UpdatedAt:             a.UpdatedAt.String,
		DeletedAt:             a.DeletedAt.String,
	}
}
