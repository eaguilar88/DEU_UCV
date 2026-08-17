package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"

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

func (r *PostgresRepository) GetActivityMetrics(ctx context.Context, groupID string) (entities.ActivityMetrics, error) {
	var metrics entities.ActivityMetrics
	query, args, err := queries.GetActivityMetrics(groupID).ToSql()
	if err != nil {
		return metrics, err
	}

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return metrics, err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, args...).Scan(
		&metrics.TotalCount,
		&metrics.UpcomingCount,
		&metrics.InCourseCount,
		&metrics.PendingReportCount,
	)
	if err != nil {
		return metrics, err
	}

	return metrics, nil
}

func (r *PostgresRepository) GetGroupDashboardSummary(ctx context.Context, groupID string) (entities.GroupDashboardSummary, error) {
	var summary entities.GroupDashboardSummary

	// Actividades planificadas
	qPlanned, argsPlanned, err := queries.CountPlannedActivitiesByGroupID(groupID).ToSql()
	if err != nil {
		return summary, err
	}
	if err := r.db.QueryRowContext(ctx, qPlanned, argsPlanned...).Scan(&summary.PlannedCount); err != nil {
		return summary, err
	}

	// Reportes pendientes
	qPending, argsPending, err := queries.CountPendingReportActivitiesByGroupID(groupID).ToSql()
	if err != nil {
		return summary, err
	}
	if err := r.db.QueryRowContext(ctx, qPending, argsPending...).Scan(&summary.PendingReportCount); err != nil {
		return summary, err
	}

	// Actividades actuales
	qCurrent, argsCurrent, err := queries.GetCurrentActivitiesByGroupID(groupID).ToSql()
	if err != nil {
		return summary, err
	}
	rows, err := r.db.QueryContext(ctx, qCurrent, argsCurrent...)
	if err != nil {
		return summary, err
	}
	defer rows.Close()

	summary.CurrentActivities = make([]entities.CurrentActivitySummary, 0)
	for rows.Next() {
		var curr entities.CurrentActivitySummary
		if err := rows.Scan(&curr.ID, &curr.Name); err != nil {
			return summary, err
		}
		summary.CurrentActivities = append(summary.CurrentActivities, curr)
	}

	return summary, nil
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

func (r *PostgresRepository) UpdateReportCheckStatus(ctx context.Context, id string, checked bool) error {
	query, args, err := queries.UpdateReportCheckStatus(id, checked).ToSql()
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

func (r *PostgresRepository) UpdateFeatureStatus(ctx context.Context, id string, featured bool) error {
	query, args, err := queries.UpdateFeatureStatus(id, featured).ToSql()
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

func (r *PostgresRepository) CountActivities(ctx context.Context, filter entities.ActivityFilter) (int, error) {
	countQuery, countArgs, err := queries.CountActivities(filter).ToSql()
	if err != nil {
		return 0, err
	}
	stmt, err := r.db.PrepareContext(ctx, countQuery)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var total int
	if err := stmt.QueryRowContext(ctx, countArgs...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
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
		&a.GroupName,
		&a.Name,
		&a.Description,
		&a.DateStart,
		&a.DateEnd,
		&a.Location,
		&a.KnowledgeArea,
		&a.Allies,
		&a.GroupParticipants,
		&a.EstimatedParticipants,
		&a.ActualParticipants,
		&a.Financing,
		&a.Comments,
		&a.GalleryURL,
		&a.ReportChecked,
		&a.IsFeatured,
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
	knowledgeArea := a.KnowledgeArea
	if knowledgeArea == nil {
		knowledgeArea = make([]string, 0)
	}

	return models.Activity{
		ID:                    a.ID,
		GroupID:               a.GroupID,
		Name:                  toNullString(a.Name),
		Description:           toNullString(a.Description),
		DateStart:             toNullString(a.DateStart),
		DateEnd:               toNullString(a.DateEnd),
		Location:              toNullString(a.Location),
		KnowledgeArea:         pq.StringArray(knowledgeArea),
		Allies:                toNullString(a.Allies),
		GroupParticipants:     sql.NullInt64{Int64: int64(a.GroupParticipants), Valid: a.GroupParticipants != 0},
		EstimatedParticipants: sql.NullInt64{Int64: int64(a.EstimatedParticipants), Valid: a.EstimatedParticipants != 0},
		ActualParticipants:    sql.NullInt64{Int64: int64(a.ActualParticipants), Valid: a.ActualParticipants != 0},
		Financing:             toNullString(a.Financing),
		Comments:              toNullString(a.Comments),
		GalleryURL:            toNullString(a.GalleryURL),
		ReportChecked:         sql.NullBool{Bool: a.ReportChecked, Valid: true},
		IsFeatured:            sql.NullBool{Bool: a.IsFeatured, Valid: true},
	}
}

func newActivityFromModel(a models.Activity) entities.Activity {
	knowledgeArea := []string(a.KnowledgeArea)
	if knowledgeArea == nil {
		knowledgeArea = make([]string, 0)
	}

	return entities.Activity{
		ID:                    a.ID,
		GroupID:               a.GroupID,
		GroupName:             a.GroupName.String,
		Name:                  a.Name.String,
		Description:           a.Description.String,
		DateStart:             a.DateStart.String,
		DateEnd:               a.DateEnd.String,
		Location:              a.Location.String,
		KnowledgeArea:         knowledgeArea,
		Allies:                a.Allies.String,
		GroupParticipants:     int(a.GroupParticipants.Int64),
		EstimatedParticipants: int(a.EstimatedParticipants.Int64),
		ActualParticipants:    int(a.ActualParticipants.Int64),
		Financing:             a.Financing.String,
		Comments:              a.Comments.String,
		GalleryURL:            a.GalleryURL.String,
		ReportChecked:         a.ReportChecked.Bool,
		IsFeatured:            a.IsFeatured.Bool,
		CreatedAt:             a.CreatedAt.String,
		UpdatedAt:             a.UpdatedAt.String,
		DeletedAt:             a.DeletedAt.String,
	}
}
