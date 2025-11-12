package repository

import (
	"context"

	"github.com/eaguilar88/deu/internal/entities"
	errs "github.com/eaguilar88/deu/internal/errors"
	"github.com/eaguilar88/deu/internal/postgres_repository/mappers"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

func (r *PostgresRepository) GetCoursePeriodByID(ctx context.Context, periodID string) (entities.CoursePeriod, error) {
	query := queries.GetCoursePeriodByID(periodID)
	sql, args, err := query.ToSql()
	if err != nil {
		return entities.CoursePeriod{}, err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return entities.CoursePeriod{}, err
	}
	//nolint:errcheck
	defer stmt.Close()
	var coursePeriod models.CoursePeriod
	row := stmt.QueryRowContext(ctx, args...)
	coursePeriod, err = scanCoursePeriod(row)
	if err != nil {
		return entities.CoursePeriod{}, err
	}
	return newCoursePeriodFromModel(coursePeriod), nil
}

func (r *PostgresRepository) GetCoursePeriods(ctx context.Context, courseID string, pageScope entities.PageScope) ([]entities.CoursePeriod, entities.PageScope, error) {
	sql, args, err := queries.GetCoursePeriods(courseID, pageScope).ToSql()
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
	var coursePeriods []entities.CoursePeriod
	for rows.Next() {
		coursePeriod, err := scanCoursePeriod(rows)
		if err != nil {
			return nil, entities.PageScope{}, err
		}
		coursePeriods = append(coursePeriods, newCoursePeriodFromModel(coursePeriod))
	}
	return coursePeriods, pageScope, nil
}

func (r *PostgresRepository) GetLatestCoursePeriod(ctx context.Context, courseID string) (entities.CoursePeriod, error) {
	sql, args, err := queries.GetLatestCoursePeriod(courseID).ToSql()
	if err != nil {
		return entities.CoursePeriod{}, err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return entities.CoursePeriod{}, err
	}
	defer stmt.Close()

	var period entities.CoursePeriod
	row := stmt.QueryRowContext(ctx, args...)
	err = row.Scan(&period.ID, &period.StartDate, &period.EndDate, &period.InscriptionDate)
	if err != nil {
		// If no period found, return empty period without error
		if err.Error() == "sql: no rows in result set" {
			return entities.CoursePeriod{}, nil
		}
		return entities.CoursePeriod{}, err
	}

	return period, nil
}

func (r *PostgresRepository) CreateCoursePeriod(ctx context.Context, coursePeriod entities.CoursePeriod) (int64, error) {
	cpModel := models.CoursePeriod{
		ID:              coursePeriod.ID,
		CourseID:        coursePeriod.Course.ID,
		StartDate:       coursePeriod.StartDate,
		EndDate:         coursePeriod.EndDate,
		InscriptionDate: coursePeriod.InscriptionDate,
	}

	sql, args, err := queries.InsertCoursePeriod(cpModel).ToSql()
	if err != nil {
		return -1, err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()
	var lastInsertedID int64
	err = stmt.QueryRowContext(ctx, args...).Scan(&lastInsertedID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == pgErrorCodeUniqueViolation {
			r.logger.Error("duplicated course period", zap.Error(err))
			return -1, errs.NewDuplicateEntryError(err)
		}
		r.logger.Error("error inserting course period", zap.Error(err))
		return -1, err
	}
	return lastInsertedID, nil
}

func (r *PostgresRepository) UpdateCoursePeriod(ctx context.Context, periodID string, coursePeriod entities.CoursePeriod) error {
	sql, args, err := queries.UpdateCoursePeriod(periodID, newCoursePeriodModelFromEntities(coursePeriod)).
		ToSql()
	if err != nil {
		return err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return err
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

func (r *PostgresRepository) DeleteCoursePeriod(ctx context.Context, periodID string) error {
	sql, args, err := queries.DeleteCoursePeriod(periodID).ToSql()
	if err != nil {
		return errs.NewBadQueryError(err)
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return err
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

func (r *PostgresRepository) GetAnnouncementsByCoursePeriodID(ctx context.Context, periodID string) ([]entities.Announcement, error) {
	sql, args, err := queries.GetAnnouncementsByCoursePeriodID(periodID).ToSql()
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

	var announcements []models.Announcement
	for rows.Next() {
		var announcement models.Announcement
		err := rows.Scan(
			&announcement.ID,
			&announcement.CourseCycleID,
			&announcement.Title,
			&announcement.Content,
			&announcement.CreatedAt,
			&announcement.UpdatedAt,
			&announcement.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		announcements = append(announcements, announcement)
	}

	return mappers.AnnouncementModelsToEntities(announcements), nil
}

func (r *PostgresRepository) GetAnnouncementByID(ctx context.Context, announcementID string) (entities.Announcement, error) {
	sql, args, err := queries.GetAnnouncementByID(announcementID).ToSql()
	if err != nil {
		return entities.Announcement{}, err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return entities.Announcement{}, err
	}
	defer stmt.Close()

	var announcement models.Announcement
	row := stmt.QueryRowContext(ctx, args...)
	err = row.Scan(
		&announcement.ID,
		&announcement.CourseCycleID,
		&announcement.Title,
		&announcement.Content,
		&announcement.CreatedAt,
		&announcement.UpdatedAt,
		&announcement.DeletedAt,
	)
	if err != nil {
		return entities.Announcement{}, err
	}

	return mappers.AnnouncementModelToEntity(announcement), nil
}

func (r *PostgresRepository) CreateAnnouncement(ctx context.Context, periodID string, announcement entities.Announcement) (int64, error) {
	announcementModel := models.Announcement{
		CourseCycleID: periodID,
		Title:         announcement.Title,
		Content:       announcement.Content,
	}

	sql, args, err := queries.InsertAnnouncement(announcementModel).ToSql()
	if err != nil {
		return -1, err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	var lastInsertedID int64
	err = stmt.QueryRowContext(ctx, args...).Scan(&lastInsertedID)
	if err != nil {
		r.logger.Error("error inserting announcement", zap.Error(err))
		return -1, err
	}

	return lastInsertedID, nil
}

func (r *PostgresRepository) UpdateAnnouncement(ctx context.Context, announcementID string, announcement entities.Announcement) error {
	announcementModel := models.Announcement{
		Title:   announcement.Title,
		Content: announcement.Content,
	}

	sql, args, err := queries.UpdateAnnouncement(announcementID, announcementModel).ToSql()
	if err != nil {
		return err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return err
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

func (r *PostgresRepository) DeleteAnnouncement(ctx context.Context, announcementID string) error {
	sql, args, err := queries.DeleteAnnouncement(announcementID).ToSql()
	if err != nil {
		return errs.NewBadQueryError(err)
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return err
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

func scanCoursePeriod(row scannable) (models.CoursePeriod, error) {
	var coursePeriod models.CoursePeriod
	err := row.Scan(
		&coursePeriod.ID,
		&coursePeriod.CourseID,
		&coursePeriod.StartDate,
		&coursePeriod.EndDate,
		&coursePeriod.IsActive,
		&coursePeriod.InscriptionDate,
		&coursePeriod.CreatedAt,
		&coursePeriod.UpdatedAt,
		&coursePeriod.DeletedAt,
	)
	return coursePeriod, err
}

func newCoursePeriodFromModel(coursePeriod models.CoursePeriod) entities.CoursePeriod {
	var deletedAt string

	if coursePeriod.DeletedAt.Valid {
		deletedAt = coursePeriod.DeletedAt.String
	}

	return entities.CoursePeriod{
		ID: coursePeriod.ID,
		Course: entities.Course{
			ID: coursePeriod.CourseID,
		},
		StartDate:       coursePeriod.StartDate,
		EndDate:         coursePeriod.EndDate,
		InscriptionDate: coursePeriod.InscriptionDate,
		IsActive:        coursePeriod.IsActive,
		CreatedAt:       coursePeriod.CreatedAt,
		UpdatedAt:       coursePeriod.UpdatedAt,
		DeletedAt:       deletedAt,
	}
}

func newCoursePeriodModelFromEntities(cp entities.CoursePeriod) models.CoursePeriod {
	return models.CoursePeriod{
		ID:              cp.ID,
		CourseID:        cp.Course.ID,
		StartDate:       cp.StartDate,
		EndDate:         cp.EndDate,
		IsActive:        cp.IsActive,
		InscriptionDate: cp.InscriptionDate,
	}
}
