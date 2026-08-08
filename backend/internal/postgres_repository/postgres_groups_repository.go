package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/groups"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

func (r *PostgresRepository) GetGroupByID(ctx context.Context, groupID string) (entities.ExtensionGroup, error) {
	query, args, err := queries.GetGroupByID(groupID).ToSql()
	if err != nil {
		return entities.ExtensionGroup{}, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return entities.ExtensionGroup{}, err
	}
	defer stmt.Close()
	var group models.ExtensionGroup
	row := stmt.QueryRowContext(ctx, args...)
	group, err = scanGroup(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return entities.ExtensionGroup{}, fmt.Errorf("%w: %w", groups.ErrGroupNotFound, err)
		}
		return entities.ExtensionGroup{}, err
	}
	eg := newGroupFromModel(group)
	eg.Members, err = r.getGroupMembers(ctx, groupID)
	if err != nil {
		return entities.ExtensionGroup{}, err
	}
	return eg, nil
}

func (r *PostgresRepository) GetGroups(ctx context.Context, filter entities.GroupFilter, pageScope entities.PageScope) ([]entities.ExtensionGroup, entities.PageScope, error) {
	query, args, err := queries.GetGroups(filter, pageScope.PerPage, pageScope.Offset()).ToSql()
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
	var groups []entities.ExtensionGroup
	for rows.Next() {
		group, err := scanGroup(rows)
		if err != nil {
			return nil, entities.PageScope{}, err
		}
		eg := newGroupFromModel(group)
		eg.Members, err = r.getGroupMembers(ctx, eg.ID)
		if err != nil {
			return nil, entities.PageScope{}, err
		}
		groups = append(groups, eg)
	}
	return groups, pageScope, nil
}

func (r *PostgresRepository) GetRandomActiveGroups(ctx context.Context, limit int) ([]entities.ExtensionGroup, error) {
	query, args, err := queries.GetRandomActiveGroups(limit).ToSql()
	if err != nil {
		return nil, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var groups []entities.ExtensionGroup
	for rows.Next() {
		group, err := scanGroup(rows)
		if err != nil {
			return nil, err
		}
		eg := newGroupFromModel(group)
		eg.Members, err = r.getGroupMembers(ctx, eg.ID)
		if err != nil {
			return nil, err
		}
		groups = append(groups, eg)
	}
	return groups, nil
}

func (r *PostgresRepository) CreateGroup(ctx context.Context, gr entities.ExtensionGroup) (int64, error) {
	query, args, err := queries.InsertGroup(newGroupToModel(gr)).ToSql()
	if err != nil {
		return 0, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	var lastInsertedID int64
	err = stmt.QueryRowContext(ctx, args...).Scan(&lastInsertedID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == pgErrorCodeUniqueViolation {
			r.logger.Error("error inserting group", zap.Error(err))
			return -1, httperrors.NewDuplicateEntryError(err)
		}
		r.logger.Error("error inserting group", zap.Error(err))
		return -1, httperrors.NewInternal(err)
	}
	if err := r.insertGroupMembers(ctx, lastInsertedID, gr.Members); err != nil {
		r.logger.Error("error inserting group members", zap.Error(err))
		return -1, err
	}
	return lastInsertedID, nil
}

// CreateGroupWithRequests creates a group and its authorization requests in a single transaction
func (r *PostgresRepository) CreateGroupWithRequests(ctx context.Context, group entities.ExtensionGroup, requests []entities.GroupRequest) (groupID int64, err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("failed to begin transaction", zap.Error(err))
		return -1, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Create the group
	groupSQL, groupArgs, err := queries.InsertGroup(newGroupToModel(group)).ToSql()
	if err != nil {
		r.logger.Error("failed to build group query", zap.Error(err))
		return -1, fmt.Errorf("failed to build group query: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, groupSQL)
	if err != nil {
		r.logger.Error("failed to prepare group statement", zap.Error(err))
		return -1, fmt.Errorf("failed to prepare group statement: %w", err)
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, groupArgs...).Scan(&groupID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == pgErrorCodeUniqueViolation {
			r.logger.Error("duplicate group entry", zap.Error(err))
			return -1, fmt.Errorf("duplicate group entry: %w ", err)
		}
		r.logger.Error("failed to insert group", zap.Error(err))
		return -1, fmt.Errorf("failed to insert group: %w", err)
	}

	// 2. Create contacts for the group
	if group.Email != "" {
		if err := r.insertContactInformation(ctx, tx, group.Email, entities.ContactTypeEmail, groupID); err != nil {
			return -1, fmt.Errorf("error saving group contact: %w", err)
		}
	}

	if group.Phone != "" {
		if err := r.insertContactInformation(ctx, tx, group.Phone, entities.ContactTypePhone, groupID); err != nil {
			return -1, fmt.Errorf("error saving group contact: %w", err)
		}
	}

	// 3. Create each authorization request
	for _, req := range requests {
		reqSQL, reqArgs, err := queries.InsertGroupRequest(models.GroupRequest{
			GroupID: groupID,
			Status:  string(req.Status),
			Faculty: string(req.Faculty),
			Comments: sql.NullString{
				String: req.Comments,
				Valid:  req.Comments != "",
			},
		}).ToSql()
		if err != nil {
			r.logger.Error("failed to build request query", zap.Error(err))
			return -1, fmt.Errorf("failed to build request query: %w", err)
		}

		reqStmt, err := tx.PrepareContext(ctx, reqSQL)
		if err != nil {
			r.logger.Error("failed to prepare request statement", zap.Error(err))
			return -1, fmt.Errorf("failed to prepare request statement: %w", err)
		}

		var reqID int64
		err = reqStmt.QueryRowContext(ctx, reqArgs...).Scan(&reqID)
		reqStmt.Close()
		if err != nil {
			if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == pgErrorCodeUniqueViolation {
				r.logger.Error("duplicate group request entry", zap.Error(err))
				return -1, httperrors.NewDuplicateEntryError(err)
			}
			r.logger.Error("failed to insert group request", zap.Error(err))
			return -1, fmt.Errorf("failed to insert group request: %w", err)
		}
	}

	// 4. Commit
	if err := tx.Commit(); err != nil {
		r.logger.Error("failed to commit transaction", zap.Error(err))
		return -1, fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("group and requests created successfully", zap.Int64("group_id", groupID))
	return groupID, nil
}

func (r *PostgresRepository) UpdateGroup(ctx context.Context, group entities.ExtensionGroup) error {
	query, args, err := queries.UpdateGroup(newGroupToModel(group)).ToSql()
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
		return fmt.Errorf("%w: %w", groups.ErrGroupNotFound, err)
	}
	return nil
}

func (r *PostgresRepository) DeleteGroup(ctx context.Context, groupID string) error {
	query, args, err := queries.DeleteGroup(groupID).ToSql()
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
		return fmt.Errorf("%w: %w", groups.ErrGroupNotFound, err)
	}
	return nil
}

func (r *PostgresRepository) CreateGroupRequest(ctx context.Context, req entities.GroupRequest) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return -1, err
	}
	defer tx.Rollback()

	groupID, err := strconv.Atoi(req.GroupID)
	if err != nil {
		r.logger.Warn("Non numeric group_id", zap.String("group_id", req.GroupID), zap.Error(err))
		return -1, err
	}

	query, args, err := queries.InsertGroupRequest(models.GroupRequest{
		GroupID: int64(groupID),
		Status:  string(req.Status),
		Faculty: string(req.Faculty),
		Comments: sql.NullString{
			String: req.Comments,
			Valid:  req.Comments != "",
		},
	}).ToSql()
	if err != nil {
		return -1, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()
	var lastInsertedID int64
	err = stmt.QueryRowContext(ctx, args...).Scan(&lastInsertedID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == pgErrorCodeUniqueViolation {
			r.logger.Error("error inserting group request", zap.Error(err), zap.String("group_id", req.GroupID))
			return -1, httperrors.NewDuplicateEntryError(err)
		}
		r.logger.Error("error inserting group request", zap.Error(err), zap.String("group_id", req.GroupID))
		return -1, httperrors.NewInternal(err)
	}

	if err := tx.Commit(); err != nil {
		return -1, err
	}
	return lastInsertedID, nil
}

func scanGroup(row scannable) (models.ExtensionGroup, error) {
	var group models.ExtensionGroup
	err := row.Scan(
		&group.ID,
		&group.UserID,
		&group.Name,
		&group.Description,
		&group.Faculty,
		&group.Foundation,
		&group.IsMultidisciplinary,
		&group.Objective,
		&group.Code,
		&group.Director,
		&group.Type,
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
		Faculty:   group.Faculty.String(),
		Foundation: sql.NullString{
			String: group.Foundation,
			Valid:  group.Foundation != "",
		},
		IsMultidisciplinary: group.IsMultidisciplinary, 
		UserID:    group.Owner.ID,
		Objective: group.Objective,
		Type:      string(group.Type),
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
	var description, objective, location, foundation string
	if group.Description.Valid {
		description = group.Description.String
	}

	if group.Location.Valid {
		location = group.Location.String
	}
	if group.Foundation.Valid {
		foundation = group.Foundation.String
	}

	return entities.ExtensionGroup{
		ID:          group.ID,
		Name:        group.Name,
		Description: description,
		Faculty:     entities.Faculty(group.Faculty),
		Foundation:  foundation,
		IsMultidisciplinary: group.IsMultidisciplinary,
		Type:        entities.GroupType(group.Type),
		Owner: &entities.User{
			ID: group.UserID,
		},
		Objective: objective,
		Location:  location,
		Active:    group.IsActive,
		CreatedAt: group.CreatedAt,
		UpdatedAt: group.UpdatedAt,
		DeletedAt: group.DeletedAt.String,
	}
}

func (r *PostgresRepository) insertGroupMembers(ctx context.Context, groupID int64, members []entities.GroupMember) error {
	for _, m := range members {
		model := models.GroupMember{
			Name:         m.Name,
			CI:           m.CI,
			Phone:        sql.NullString{String: m.Phone, Valid: m.Phone != ""},
			Email:        sql.NullString{String: m.Email, Valid: m.Email != ""},
			Coordination: sql.NullString{String: m.Coordination, Valid: m.Coordination != ""},
			Year:         sql.NullString{String: m.Year, Valid: m.Year != ""},
			Faculty:      string(m.Faculty),
			School:       sql.NullString{String: m.School, Valid: m.School != ""},
			Document:     sql.NullString{String: m.Document, Valid: m.Document != ""},
			IsActive:     m.IsActive,
		}
		query, args, err := queries.InsertGroupMember(groupID, model).ToSql()
		if err != nil {
			return err
		}
		stmt, err := r.db.PrepareContext(ctx, query)
		if err != nil {
			return err
		}
		var id int64
		err = stmt.QueryRowContext(ctx, args...).Scan(&id)
		stmt.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresRepository) getGroupMembers(ctx context.Context, groupID string) ([]entities.GroupMember, error) {
	query, args, err := queries.SelectGroupMembers(groupID).ToSql()
	if err != nil {
		return nil, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var members []entities.GroupMember
	for rows.Next() {
		var m models.GroupMember
		err := rows.Scan(&m.ID, &m.Name, &m.CI, &m.Phone, &m.Email, &m.Coordination, &m.Year, &m.Faculty, &m.School, &m.Document, &m.IsActive)
		if err != nil {
			return nil, err
		}
		members = append(members, groupMemberFromModel(m))
	}
	return members, nil
}

func groupMemberFromModel(m models.GroupMember) entities.GroupMember {
	return entities.GroupMember{
		ID:           m.ID,
		Name:         m.Name,
		CI:           m.CI,
		Phone:        m.Phone.String,
		Email:        m.Email.String,
		Coordination: m.Coordination.String,
		Year:         m.Year.String,
		Faculty:      entities.Faculty(m.Faculty),
		School:       m.School.String,
		Document:     m.Document.String,
		IsActive:     m.IsActive,
	}
}

func (r *PostgresRepository) insertContactInformation(ctx context.Context, tx *sql.Tx, contactValue string, contactType entities.ContactType, groupID int64) error {
	contactSQL, contactArgs, err := queries.InsertContact(
		string(contactType),
		contactValue,
		entities.OwnerTypeExtensionGroup.String(),
		groupID,
	).ToSql()
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	contactStmt, err := tx.PrepareContext(ctx, contactSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare contact statement: %w", err)
	}
	defer contactStmt.Close()

	var contactID string
	err = contactStmt.QueryRowContext(ctx, contactArgs...).Scan(&contactID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == pgErrorCodeUniqueViolation {
			return fmt.Errorf("duplicated contact information: %w", err)
		}
		return fmt.Errorf("failed to insert contact: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetContactsByOwner(ctx context.Context, ownerID string, ownerType entities.OwnerType) ([]entities.Contact, error) {
	query := `
		SELECT id, contact_type, contact_value, owner_type, owner_id
		FROM deu.contacts
		WHERE owner_id = $1 AND owner_type = $2 AND deleted_at IS NULL
	`
	rows, err := r.db.QueryContext(ctx, query, ownerID, string(ownerType))
	if err != nil {
		return nil, fmt.Errorf("failed to query contacts: %w", err)
	}
	defer rows.Close()

	var contacts []entities.Contact
	for rows.Next() {
		var c entities.Contact
		if err := rows.Scan(&c.ID, &c.Type, &c.Value, &c.OwnerType, &c.OwnerID); err != nil {
			return nil, fmt.Errorf("failed to scan contact: %w", err)
		}
		contacts = append(contacts, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return contacts, nil
}
