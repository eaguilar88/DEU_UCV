package queries

import (
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/internal/entities"
)

// CountStaleCourseRequestsByFaculty counts the course requests still under review that were
// created before createdBefore, grouped by the faculty that has to review them.
func CountStaleCourseRequestsByFaculty(createdBefore time.Time) sq.SelectBuilder {
	return psql.Select("c.faculty", "COUNT(*)").
		From(fmt.Sprintf("%s AS r", courseRequestsTableName)).
		Join(fmt.Sprintf("%s AS c ON c.id = r.course_id", coursesTableName)).
		Where(sq.Eq{"r.deleted_at": nil, "c.deleted_at": nil, "r.status": "under_review"}).
		Where(sq.NotEq{"c.faculty": nil}).
		Where(sq.Lt{"r.created_at": createdBefore}).
		GroupBy("c.faculty")
}

// CountStaleGroupRequestsByFaculty counts the group requests still under review that were created
// before createdBefore, grouped by the faculty that has to review them.
func CountStaleGroupRequestsByFaculty(createdBefore time.Time) sq.SelectBuilder {
	return psql.Select("gr.faculty", "COUNT(*)").
		From(fmt.Sprintf("%s AS gr", groupRequestsTableName)).
		Join(fmt.Sprintf("%s AS g ON g.id = gr.group_id", groupsTableName)).
		Where(sq.Eq{"g.deleted_at": nil, "gr.status": "under_review"}).
		Where(sq.Lt{"gr.created_at": createdBefore}).
		GroupBy("gr.faculty")
}

// GetGroupsDueForRenewal lists the approved groups whose renewal is due before dueBefore,
// including the ones already overdue (and deactivated for it). Groups that already submitted a
// renewal under review are left out.
func GetGroupsDueForRenewal(dueBefore time.Time) sq.SelectBuilder {
	return psql.Select("g.id", "g.name", "g.renewal_due_at").
		From(fmt.Sprintf("%s AS g", groupsTableName)).
		Where(sq.Eq{"g.deleted_at": nil}).
		Where(sq.NotEq{"g.renewal_due_at": nil}).
		Where(sq.Lt{"g.renewal_due_at": dueBefore}).
		Where(fmt.Sprintf("NOT EXISTS (SELECT 1 FROM %s r WHERE r.group_id = g.id AND r.status = 'under_review')", groupRenewalsTableName)).
		OrderBy("g.renewal_due_at")
}

// DeactivateExpiredGroups deactivates the active groups whose renewal was due before now.
func DeactivateExpiredGroups(now time.Time) sq.UpdateBuilder {
	return psql.Update(groupsTableName).
		Set("is_active", false).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"deleted_at": nil, "is_active": true}).
		Where(sq.Lt{"renewal_due_at": now}).
		Suffix("RETURNING id, name, renewal_due_at")
}

// CountNotificationsSince counts the reminders of kind about target sent to recipient after since.
func CountNotificationsSince(kind, target, recipient string, since time.Time) sq.SelectBuilder {
	return psql.Select("COUNT(*)").
		From(notificationsLogTableName).
		Where(sq.Eq{"kind": kind, "target": target, "recipient": recipient}).
		Where(sq.Gt{"sent_at": since})
}

func InsertNotificationLog(n entities.NotificationLog) sq.InsertBuilder {
	return psql.Insert(notificationsLogTableName).
		Columns("kind", "target", "recipient").
		Values(n.Kind, n.Target, n.Recipient)
}
