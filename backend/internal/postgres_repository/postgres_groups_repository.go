package repository

import (
	"context"
	"database/sql"

	"github.com/eaguilar88/deu/internal/entities"
	errs "github.com/eaguilar88/deu/internal/errors"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
)

func (r *PostgresRepository) GetGroupByID(
	ctx context.Context,
	groupID string,
) (entities.ExtensionGroup, error) {
	sql, args, err := queries.GetGroupByID(groupID).ToSql()
	if err != nil {
		return entities.ExtensionGroup{}, err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return entities.ExtensionGroup{}, err
	}
	defer stmt.Close()
	var group models.ExtensionGroup
	row := stmt.QueryRowContext(ctx, args...)
	group, err = scanGroup(row)
	if err != nil {
		return entities.ExtensionGroup{}, err
	}
	return newGroupFromModel(group), nil
}

func (r *PostgresRepository) GetGroups(
	ctx context.Context,
	pageScope entities.PageScope,
) ([]entities.ExtensionGroup, entities.PageScope, error) {
	sql, args, err := queries.GetGroups(pageScope.PerPage, pageScope.Offset()).ToSql()
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
	var groups []entities.ExtensionGroup
	for rows.Next() {
		group, err := scanGroup(rows)
		if err != nil {
			return nil, entities.PageScope{}, err
		}
		groups = append(groups, newGroupFromModel(group))
	}
	return groups, pageScope, nil
}

func (r *PostgresRepository) CreateGroup(
	ctx context.Context,
	group entities.ExtensionGroup,
) (int64, error) {
	sql, args, err := queries.InsertGroup(newGroupToModel(group)).ToSql()
	if err != nil {
		return 0, err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *PostgresRepository) UpdateGroup(ctx context.Context, group entities.ExtensionGroup) error {
	sql, args, err := queries.UpdateGroup(newGroupToModel(group)).ToSql()
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

func (r *PostgresRepository) DeleteGroup(ctx context.Context, groupID string) error {
	sql, args, err := queries.DeleteGroup(groupID).ToSql()
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

func scanGroup(row scannable) (models.ExtensionGroup, error) {
	var group models.ExtensionGroup
	err := row.Scan(
		&group.ID,
		&group.Name,
		&group.Description,
		&group.OwnerID,
		&group.OwnerFirstName,
		&group.OwnerLastName,
		&group.RequestID,
		&group.RequesterID,
		&group.RequesterFirstName,
		&group.RequesterLastName,
		&group.Objective,
		&group.Location,
		&group.IsActive,
		&group.CreatedAt,
		&group.UpdatedAt,
		&group.DeletedAt,
	)
	if err != nil {
		return models.ExtensionGroup{}, err
	}
	return group, nil
}

func newGroupToModel(group entities.ExtensionGroup) models.ExtensionGroup {
	return models.ExtensionGroup{
		ID:   group.ID,
		Name: group.Name,
		Description: sql.NullString{
			String: group.Description,
			Valid:  true,
		},
		OwnerID:   group.Owner.ID,
		RequestID: group.CourseRequest.ID,
		Objective: sql.NullString{
			String: group.Objective,
			Valid:  true,
		},
		Location: sql.NullString{
			String: group.Location,
			Valid:  true,
		},
		IsActive:  group.Active,
		CreatedAt: group.CreatedAt,
		UpdatedAt: group.UpdatedAt,
	}
}

func newGroupFromModel(group models.ExtensionGroup) entities.ExtensionGroup {
	var description, objective, location string
	if group.Description.Valid {
		description = group.Description.String
	}
	if group.Objective.Valid {
		objective = group.Objective.String
	}
	if group.Location.Valid {
		location = group.Location.String
	}

	return entities.ExtensionGroup{
		ID:          group.ID,
		Name:        group.Name,
		Description: description,
		Owner: entities.User{
			ID:        group.OwnerID,
			FirstName: group.OwnerFirstName,
			LastName:  group.OwnerLastName,
		},
		CourseRequest: entities.CourseRequest{
			ID: group.RequestID,
			User: entities.User{
				ID:        group.RequesterID,
				FirstName: group.RequesterFirstName,
				LastName:  group.RequesterLastName,
			},
		},
		Objective: objective,
		Location:  location,
		Active:    group.IsActive,
		CreatedAt: group.CreatedAt,
		UpdatedAt: group.UpdatedAt,
		DeletedAt: group.DeletedAt,
	}
}
