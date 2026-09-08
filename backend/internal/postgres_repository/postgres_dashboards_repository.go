package repository

import (
	"context"

	"github.com/eaguilar88/deu/internal/group_dashboards"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
)

func (r *PostgresRepository) GetGroupDashboardMetrics(ctx context.Context, groupID string) (int, int, error) {
	sqlUpcoming, argsUpcoming, err := queries.CountPlannedActivitiesByGroupID(groupID).ToSql()
	if err != nil {
		return 0, 0, err
	}
	var upcoming int
	if err := r.db.QueryRowContext(ctx, sqlUpcoming, argsUpcoming...).Scan(&upcoming); err != nil {
		return 0, 0, err
	}

	sqlPending, argsPending, err := queries.CountPendingReportActivitiesByGroupID(groupID).ToSql()
	if err != nil {
		return 0, 0, err
	}
	var pendingReports int
	if err := r.db.QueryRowContext(ctx, sqlPending, argsPending...).Scan(&pendingReports); err != nil {
		return 0, 0, err
	}

	return upcoming, pendingReports, nil
}

func (r *PostgresRepository) GetFacultyDashboardMetrics(ctx context.Context, faculty string) (int, int, error) {
	sqlRequests, argsRequests, err := queries.CountPendingGroupRequestsByFaculty(faculty).ToSql()
	if err != nil {
		return 0, 0, err
	}

	var pendingRequests int
	if err := r.db.QueryRowContext(ctx, sqlRequests, argsRequests...).Scan(&pendingRequests); err != nil {
		return 0, 0, err
	}

	sqlGroups, argsGroups, err := queries.CountGroupsByFaculty(faculty).ToSql()
	if err != nil {
		return 0, 0, err
	}

	var totalGroups int
	if err := r.db.QueryRowContext(ctx, sqlGroups, argsGroups...).Scan(&totalGroups); err != nil {
		return 0, 0, err
	}

	return pendingRequests, totalGroups, nil
}

func (r *PostgresRepository) GetDeuDashboardMetrics(ctx context.Context) (int, []group_dashboards.ResourceRequestsByFaculty, int, int, error) {
	sqlPendingDeu, argsPendingDeu, err := queries.GetDeuPendingRequestsCount().ToSql()
	if err != nil {
		return 0, nil, 0, 0, err
	}
	var pendingDeu int
	if err := r.db.QueryRowContext(ctx, sqlPendingDeu, argsPendingDeu...).Scan(&pendingDeu); err != nil {
		return 0, nil, 0, 0, err
	}

	sqlReqs, argsReqs, err := queries.CountPendingGroupResourceRequestsByFaculty("").ToSql()
	if err != nil {
		return 0, nil, 0, 0, err
	}

	rows, err := r.db.QueryContext(ctx, sqlReqs, argsReqs...)
	if err != nil {
		return 0, nil, 0, 0, err
	}
	defer rows.Close()

	var resourceReqs []group_dashboards.ResourceRequestsByFaculty
	for rows.Next() {
		var item group_dashboards.ResourceRequestsByFaculty
		if err := rows.Scan(&item.Faculty, &item.Count); err != nil {
			return 0, nil, 0, 0, err
		}
		resourceReqs = append(resourceReqs, item)
	}

	if err := rows.Err(); err != nil {
		return 0, nil, 0, 0, err
	}

	sqlActiveGroups, argsActiveGroups, err := queries.GetDeuActiveGroupsCount().ToSql()
	if err != nil {
		return 0, nil, 0, 0, err
	}
	var activeGroups int
	if err := r.db.QueryRowContext(ctx, sqlActiveGroups, argsActiveGroups...).Scan(&activeGroups); err != nil {
		return 0, nil, 0, 0, err
	}

	sqlInactiveGroups, argsInactiveGroups, err := queries.GetDeuInactiveGroupsCount().ToSql()
	if err != nil {
		return 0, nil, 0, 0, err
	}
	var inactiveGroups int
	if err := r.db.QueryRowContext(ctx, sqlInactiveGroups, argsInactiveGroups...).Scan(&inactiveGroups); err != nil {
		return 0, nil, 0, 0, err
	}

	return pendingDeu, resourceReqs, activeGroups, inactiveGroups, nil
}
