package queries

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
)

func InsertGroupRenewal(r models.GroupRenewal) sq.InsertBuilder {
	return psql.Insert(groupRenewalsTableName).
		Columns("group_id", "payload", "submitted_by").
		Values(r.GroupID, string(r.Payload), r.SubmittedBy).
		Suffix("RETURNING id")
}

func GetGroupRenewalByID(renewalID string) sq.SelectBuilder {
	return psql.Select("id", "group_id", "payload", "submitted_by", "status", "created_at").
		From(groupRenewalsTableName).
		Where(sq.Eq{"id": renewalID})
}

func DeleteGroupRenewal(renewalID string) sq.DeleteBuilder {
	return psql.Delete(groupRenewalsTableName).
		Where(sq.Eq{"id": renewalID})
}

// ReviewGroupRenewal sets the status of a renewal still under review.
func ReviewGroupRenewal(renewalID string, status entities.RequestStatus) sq.UpdateBuilder {
	return psql.Update(groupRenewalsTableName).
		Set("status", string(status)).
		Set("reviewed_at", sq.Expr("NOW()")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": renewalID, "status": "under_review"})
}

// RejectPendingRenewalRequests rejects the requests of a renewal that are still under review.
func RejectPendingRenewalRequests(renewalID, reason string) sq.UpdateBuilder {
	return psql.Update(groupRequestsTableName).
		Set("status", "rejected").
		Set("comments", reason).
		Set("reviewed_at", sq.Expr("NOW()")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"renewal_id": renewalID, "status": "under_review"})
}

// ApplyGroupRenewal replaces the group's data with the renewal's and starts a new yearly period:
// from the current due date when renewed on time, or from now when overdue.
func ApplyGroupRenewal(groupID string, g models.ExtensionGroup) sq.UpdateBuilder {
	return psql.Update(groupsTableName).
		Set("name", g.Name).
		Set("description", g.Description).
		Set("faculty", g.Faculty).
		Set("foundation", g.Foundation).
		Set("is_multidisciplinary", g.IsMultidisciplinary).
		Set("objective", g.Objective).
		Set("type", g.Type).
		Set("location", g.Location).
		Set("is_active", true).
		Set("renewal_due_at", sq.Expr("GREATEST(COALESCE(renewal_due_at, NOW()), NOW()) + INTERVAL '1 year'")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": groupID})
}

func SoftDeleteGroupContacts(groupID string) sq.UpdateBuilder {
	return psql.Update(contactsTableName).
		Set("deleted_at", sq.Expr("NOW()")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"owner_type": string(entities.OwnerTypeExtensionGroup), "owner_id": groupID, "deleted_at": nil})
}

// RestoreContact inserts a contact, or brings back a soft-deleted one with the same value.
func RestoreContact(contactType, value, ownerType string, ownerID string) sq.InsertBuilder {
	return psql.Insert(contactsTableName).
		Columns("contact_type", "contact_value", "owner_type", "owner_id").
		Values(contactType, value, ownerType, ownerID).
		Suffix("ON CONFLICT (contact_type, contact_value, owner_type, owner_id) DO UPDATE SET deleted_at = NULL, updated_at = NOW()")
}

// SoftDeleteGroupMembers removes the group's current members and returns their IDs.
func SoftDeleteGroupMembers(groupID string) sq.UpdateBuilder {
	return psql.Update(groupMembersTableName).
		Set("deleted_at", sq.Expr("NOW()")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"group_id": groupID, "deleted_at": nil}).
		Suffix("RETURNING id")
}

// SoftDeleteFiles removes the live files of an owner, optionally only those with the given purposes.
func SoftDeleteFiles(ownerType entities.OwnerType, ownerIDs []string, purposes ...string) sq.UpdateBuilder {
	q := psql.Update(filesTableName).
		Set("deleted_at", sq.Expr("NOW()")).
		Where(sq.Eq{"owner_type": string(ownerType), "owner_id": ownerIDs, "deleted_at": nil})
	if len(purposes) > 0 {
		q = q.Where(sq.Eq{"purpose": purposes})
	}
	return q
}

// MoveRenewalFiles reassigns the renewal's proposed files with the given purposes to the group.
func MoveRenewalFiles(renewalID, groupID string, purposes ...string) sq.UpdateBuilder {
	return psql.Update(filesTableName).
		Set("owner_type", string(entities.OwnerTypeExtensionGroup)).
		Set("owner_id", groupID).
		Where(sq.Eq{"owner_type": string(entities.OwnerTypeGroupRenewal), "owner_id": renewalID, "purpose": purposes, "deleted_at": nil})
}

// MoveRenewalMemberDocument reassigns the proposed document of the renewal's memberIndex-th
// member to the member created for it.
func MoveRenewalMemberDocument(renewalID string, memberIndex int, memberID string) sq.UpdateBuilder {
	return psql.Update(filesTableName).
		Set("owner_type", string(entities.OwnerTypeGroupMember)).
		Set("owner_id", memberID).
		Set("metadata", sq.Expr("metadata || jsonb_build_object('member_id', ?::text)", memberID)).
		Where(sq.Eq{
			"owner_type": string(entities.OwnerTypeGroupRenewal),
			"owner_id":   renewalID,
			"purpose":    entities.GroupMemberFileTypeDocument,
			"deleted_at": nil,
		}).
		Where(sq.Expr(fmt.Sprintf("metadata->>'%s' = ?", entities.GroupRenewalMemberIndexKey), fmt.Sprintf("%d", memberIndex)))
}
