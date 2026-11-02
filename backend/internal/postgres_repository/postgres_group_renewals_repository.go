package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/group_requests"
	"github.com/eaguilar88/deu/internal/groups"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

// CreateGroupRenewal stores a group's proposed renewal and its approval requests in a single
// transaction. It fails with groups.ErrRenewalPending when the group already has a renewal under
// review.
func (r *PostgresRepository) CreateGroupRenewal(ctx context.Context, renewal entities.GroupRenewal, requests []entities.GroupRequest) (int64, error) {
	model, err := newGroupRenewalToModel(renewal)
	if err != nil {
		return -1, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("failed to begin transaction", zap.Error(err))
		return -1, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query, args, err := queries.InsertGroupRenewal(model).ToSql()
	if err != nil {
		return -1, fmt.Errorf("failed to build group renewal query: %w", err)
	}
	var renewalID int64
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&renewalID); err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == pgErrorCodeUniqueViolation {
			return -1, fmt.Errorf("%w: %w", groups.ErrRenewalPending, err)
		}
		r.logger.Error("failed to insert group renewal", zap.Error(err), zap.String("group_id", renewal.GroupID))
		return -1, fmt.Errorf("failed to insert group renewal: %w", err)
	}

	groupID, err := strconv.ParseInt(renewal.GroupID, 10, 64)
	if err != nil {
		return -1, err
	}
	for _, req := range requests {
		reqSQL, reqArgs, err := queries.InsertGroupRequest(models.GroupRequest{
			GroupID:   groupID,
			RenewalID: sql.NullInt64{Int64: renewalID, Valid: true},
			Status:    string(req.Status),
			Faculty:   string(req.Faculty),
			Comments:  sql.NullString{String: req.Comments, Valid: req.Comments != ""},
		}).ToSql()
		if err != nil {
			return -1, fmt.Errorf("failed to build request query: %w", err)
		}
		var reqID int64
		if err := tx.QueryRowContext(ctx, reqSQL, reqArgs...).Scan(&reqID); err != nil {
			r.logger.Error("failed to insert group renewal request", zap.Error(err), zap.String("group_id", renewal.GroupID))
			return -1, fmt.Errorf("failed to insert group renewal request: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return -1, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return renewalID, nil
}

// CancelGroupRenewal deletes a renewal and, by cascade, its requests. It undoes a submission whose
// files could not be stored.
func (r *PostgresRepository) CancelGroupRenewal(ctx context.Context, renewalID string) error {
	query, args, err := queries.DeleteGroupRenewal(renewalID).ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *PostgresRepository) GetGroupRenewal(ctx context.Context, renewalID string) (entities.GroupRenewal, error) {
	query, args, err := queries.GetGroupRenewalByID(renewalID).ToSql()
	if err != nil {
		return entities.GroupRenewal{}, err
	}
	var m models.GroupRenewal
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&m.ID, &m.GroupID, &m.Payload, &m.SubmittedBy, &m.Status, &m.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return entities.GroupRenewal{}, fmt.Errorf("%w", group_requests.ErrGroupRenewalNotFound)
	}
	if err != nil {
		return entities.GroupRenewal{}, err
	}
	return newGroupRenewalFromModel(m)
}

// ApplyGroupRenewal approves the renewal's last pending request (reqID) and replaces the group's
// data, contacts, members and files with the renewal's, all in a single transaction. The group is
// reactivated and its next renewal is due a year later.
func (r *PostgresRepository) ApplyGroupRenewal(ctx context.Context, reqID string, renewal entities.GroupRenewal) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("failed to begin transaction", zap.Error(err))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := execExpectingRows(ctx, tx, queries.ApproveGroupRequest(reqID), group_requests.ErrGroupRequestNotFound); err != nil {
		return err
	}
	if err := execExpectingRows(ctx, tx, queries.ReviewGroupRenewal(renewal.ID, entities.RequestStatus_APPROVED), group_requests.ErrGroupRenewalNotFound); err != nil {
		return err
	}

	groupID := renewal.GroupID
	if err := execExpectingRows(ctx, tx, queries.ApplyGroupRenewal(groupID, newGroupToModel(renewal.Group)), groups.ErrGroupNotFound); err != nil {
		return err
	}

	if err := r.replaceGroupContacts(ctx, tx, groupID, renewal.Group); err != nil {
		return err
	}

	oldMemberIDs, err := r.softDeleteGroupMembers(ctx, tx, groupID)
	if err != nil {
		return err
	}
	groupIDNum, err := strconv.ParseInt(groupID, 10, 64)
	if err != nil {
		return err
	}
	members := make([]entities.GroupMember, len(renewal.Group.Members))
	for i, m := range renewal.Group.Members {
		m.ID = "" // always insert: the renewal replaces every member
		members[i] = m
	}
	newMembers, err := r.upsertGroupMembers(ctx, tx, groupIDNum, members)
	if err != nil {
		return fmt.Errorf("failed to insert renewed group members: %w", err)
	}

	groupFilePurposes := []string{entities.GroupFileTypeLogo, entities.GroupFileTypeProject}
	fileUpdates := []sqlizer{
		queries.SoftDeleteFiles(entities.OwnerTypeExtensionGroup, []string{groupID}, groupFilePurposes...),
		queries.MoveRenewalFiles(renewal.ID, groupID, groupFilePurposes...),
	}
	if len(oldMemberIDs) > 0 {
		fileUpdates = append(fileUpdates, queries.SoftDeleteFiles(entities.OwnerTypeGroupMember, oldMemberIDs))
	}
	for i, m := range newMembers {
		fileUpdates = append(fileUpdates, queries.MoveRenewalMemberDocument(renewal.ID, i, m.ID))
	}
	for _, q := range fileUpdates {
		if err := execBuilder(ctx, tx, q); err != nil {
			return fmt.Errorf("failed to move renewal files: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// RejectGroupRenewal rejects reqID, the renewal's other pending requests and the renewal itself,
// so the group can submit a new one.
func (r *PostgresRepository) RejectGroupRenewal(ctx context.Context, reqID, renewalID, reason string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("failed to begin transaction", zap.Error(err))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := execExpectingRows(ctx, tx, queries.RejectGroupRequest(reqID, reason), group_requests.ErrGroupRequestNotFound); err != nil {
		return err
	}
	if err := execBuilder(ctx, tx, queries.RejectPendingRenewalRequests(renewalID, reason)); err != nil {
		return err
	}
	if err := execExpectingRows(ctx, tx, queries.ReviewGroupRenewal(renewalID, entities.RequestStatus_REJECTED), group_requests.ErrGroupRenewalNotFound); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (r *PostgresRepository) replaceGroupContacts(ctx context.Context, tx *sql.Tx, groupID string, group entities.ExtensionGroup) error {
	if err := execBuilder(ctx, tx, queries.SoftDeleteGroupContacts(groupID)); err != nil {
		return fmt.Errorf("failed to remove group contacts: %w", err)
	}
	contacts := []struct {
		kind  entities.ContactType
		value string
	}{
		{entities.ContactTypeEmail, group.Email},
		{entities.ContactTypePhone, group.Phone},
	}
	for _, c := range contacts {
		if c.value == "" {
			continue
		}
		q := queries.RestoreContact(string(c.kind), c.value, entities.OwnerTypeExtensionGroup.String(), groupID)
		if err := execBuilder(ctx, tx, q); err != nil {
			return fmt.Errorf("failed to save group contact: %w", err)
		}
	}
	return nil
}

func (r *PostgresRepository) softDeleteGroupMembers(ctx context.Context, tx *sql.Tx, groupID string) ([]string, error) {
	query, args, err := queries.SoftDeleteGroupMembers(groupID).ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to remove group members: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, strconv.FormatInt(id, 10))
	}
	return ids, rows.Err()
}

type sqlizer interface {
	ToSql() (string, []any, error)
}

func execBuilder(ctx context.Context, tx *sql.Tx, b sqlizer) error {
	query, args, err := b.ToSql()
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, query, args...)
	return err
}

// execExpectingRows runs b and returns notFound when it changes no row.
func execExpectingRows(ctx context.Context, tx *sql.Tx, b sqlizer, notFound error) error {
	query, args, err := b.ToSql()
	if err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("%w", notFound)
	}
	return nil
}

func newGroupRenewalToModel(renewal entities.GroupRenewal) (models.GroupRenewal, error) {
	g := renewal.Group
	payload := models.GroupRenewalPayload{
		Name:                g.Name,
		Description:         g.Description,
		Foundation:          g.Foundation,
		IsMultidisciplinary: g.IsMultidisciplinary,
		Faculty:             make([]string, len(g.Faculty)),
		Type:                make([]string, len(g.Type)),
		Objective:           g.Objective,
		Location:            g.Location,
		Email:               g.Email,
		Phone:               g.Phone,
		Members:             make([]models.GroupRenewalPayloadMember, len(g.Members)),
	}
	for i, f := range g.Faculty {
		payload.Faculty[i] = string(f)
	}
	for i, t := range g.Type {
		payload.Type[i] = string(t)
	}
	for i, m := range g.Members {
		payload.Members[i] = models.GroupRenewalPayloadMember{
			Name:         m.Name,
			CI:           m.CI,
			Phone:        m.Phone,
			Email:        m.Email,
			Coordination: m.Coordination,
			Year:         m.Year,
			Faculty:      string(m.Faculty),
			School:       m.School,
			IsLeader:     m.IsLeader,
			IsActive:     m.IsActive,
		}
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return models.GroupRenewal{}, fmt.Errorf("failed to encode group renewal: %w", err)
	}
	return models.GroupRenewal{
		GroupID:     renewal.GroupID,
		Payload:     raw,
		SubmittedBy: sql.NullString{String: renewal.SubmittedBy, Valid: renewal.SubmittedBy != ""},
	}, nil
}

func newGroupRenewalFromModel(m models.GroupRenewal) (entities.GroupRenewal, error) {
	var payload models.GroupRenewalPayload
	if err := json.Unmarshal(m.Payload, &payload); err != nil {
		return entities.GroupRenewal{}, fmt.Errorf("failed to decode group renewal %d: %w", m.ID, err)
	}

	group := entities.ExtensionGroup{
		ID:                  m.GroupID,
		Name:                payload.Name,
		Description:         payload.Description,
		Foundation:          payload.Foundation,
		IsMultidisciplinary: payload.IsMultidisciplinary,
		Faculty:             make([]entities.Faculty, len(payload.Faculty)),
		Type:                make([]entities.GroupType, len(payload.Type)),
		Objective:           payload.Objective,
		Location:            payload.Location,
		Email:               payload.Email,
		Phone:               payload.Phone,
		Members:             make([]entities.GroupMember, len(payload.Members)),
	}
	for i, f := range payload.Faculty {
		group.Faculty[i] = entities.Faculty(f)
	}
	for i, t := range payload.Type {
		group.Type[i] = entities.GroupType(t)
	}
	for i, pm := range payload.Members {
		group.Members[i] = entities.GroupMember{
			Name:         pm.Name,
			CI:           pm.CI,
			Phone:        pm.Phone,
			Email:        pm.Email,
			Coordination: pm.Coordination,
			Year:         pm.Year,
			Faculty:      entities.Faculty(pm.Faculty),
			School:       pm.School,
			IsLeader:     pm.IsLeader,
			IsActive:     pm.IsActive,
		}
	}

	return entities.GroupRenewal{
		ID:          strconv.FormatInt(m.ID, 10),
		GroupID:     m.GroupID,
		Group:       group,
		Status:      entities.RequestStatus(m.Status),
		SubmittedBy: m.SubmittedBy.String,
		CreatedAt:   m.CreatedAt.Format(time.RFC3339),
	}, nil
}
