package repository

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
)

// CountStaleCourseRequestsByFaculty counts, per reviewing faculty, the course requests under
// review created before createdBefore.
func (r *PostgresRepository) CountStaleCourseRequestsByFaculty(ctx context.Context, createdBefore time.Time) ([]entities.FacultyPendingCount, error) {
	return r.countPendingByFaculty(ctx, queries.CountStaleCourseRequestsByFaculty(createdBefore))
}

// CountStaleGroupRequestsByFaculty counts, per reviewing faculty, the group requests under review
// created before createdBefore.
func (r *PostgresRepository) CountStaleGroupRequestsByFaculty(ctx context.Context, createdBefore time.Time) ([]entities.FacultyPendingCount, error) {
	return r.countPendingByFaculty(ctx, queries.CountStaleGroupRequestsByFaculty(createdBefore))
}

func (r *PostgresRepository) countPendingByFaculty(ctx context.Context, builder sq.SelectBuilder) ([]entities.FacultyPendingCount, error) {
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []entities.FacultyPendingCount
	for rows.Next() {
		var f string
		var count int
		if err := rows.Scan(&f, &count); err != nil {
			return nil, err
		}
		fac, err := entities.FromString(f)
		if err != nil {
			return nil, err
		}
		result = append(result, entities.FacultyPendingCount{Faculty: fac, Count: count})
	}
	return result, rows.Err()
}

// GetGroupsDueForRenewal lists the approved groups whose renewal is due before dueBefore.
func (r *PostgresRepository) GetGroupsDueForRenewal(ctx context.Context, dueBefore time.Time) ([]entities.GroupRenewalDue, error) {
	return r.queryGroupRenewalsDue(ctx, queries.GetGroupsDueForRenewal(dueBefore))
}

// DeactivateExpiredGroups deactivates the active groups whose renewal was due before now and
// returns them.
func (r *PostgresRepository) DeactivateExpiredGroups(ctx context.Context, now time.Time) ([]entities.GroupRenewalDue, error) {
	return r.queryGroupRenewalsDue(ctx, queries.DeactivateExpiredGroups(now))
}

func (r *PostgresRepository) queryGroupRenewalsDue(ctx context.Context, builder sqlizer) ([]entities.GroupRenewalDue, error) {
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []entities.GroupRenewalDue
	for rows.Next() {
		var g entities.GroupRenewalDue
		if err := rows.Scan(&g.GroupID, &g.GroupName, &g.RenewalDueAt); err != nil {
			return nil, err
		}
		result = append(result, g)
	}
	return result, rows.Err()
}

// WasNotifiedSince reports whether a reminder of kind about target was sent to recipient after since.
func (r *PostgresRepository) WasNotifiedSince(ctx context.Context, kind, target, recipient string, since time.Time) (bool, error) {
	query, args, err := queries.CountNotificationsSince(kind, target, recipient, since).ToSql()
	if err != nil {
		return false, err
	}

	var count int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

// LogNotification records that a reminder was sent.
func (r *PostgresRepository) LogNotification(ctx context.Context, n entities.NotificationLog) error {
	query, args, err := queries.InsertNotificationLog(n).ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}
