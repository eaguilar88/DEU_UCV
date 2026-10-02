package queries

import (
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
)

var courseRequestQuerySelectCommon = []string{
	"r.id",
	"r.course_id",
	"r.status",
	"r.reviewer_id",
	"r.comments",
	"r.reviewed_at",
	"r.created_at",
	"r.updated_at",
	"r.deleted_at",
	"r.score",
	"r.classification",
}

var courseRequestWithCourseSelect = []string{
	"r.id",
	"r.course_id",
	"r.status",
	"r.reviewer_id",
	"r.comments",
	"r.reviewed_at",
	"r.created_at",
	"r.updated_at",
	"r.deleted_at",
	"r.score",
	"r.classification",
	// Course fields (excluding deleted_at)
	"c.id",
	"c.name",
	"c.description",
	"c.provider_id",
	"c.objectives",
	"c.duration",
	"c.content",
	"c.type",
	"c.faculty",
	"c.cost",
	"c.location",
	"c.is_active",
	"c.created_at",
	"c.updated_at",
	"c.has_documentation",
	// The course owner's user account.
	"p.user_id",
}

// fromCourseRequestsWithCourse joins each request (r) with its course (c) and provider (p).
func fromCourseRequestsWithCourse(q sq.SelectBuilder) sq.SelectBuilder {
	return q.From(fmt.Sprintf("%s AS r", courseRequestsTableName)).
		Join(fmt.Sprintf("%s AS c ON c.id = r.course_id", coursesTableName)).
		Join(fmt.Sprintf("%s AS p ON p.id = c.provider_id", providersTableName))
}

func GetCourseRequestByID(requestID string) sq.SelectBuilder {
	return fromCourseRequestsWithCourse(psql.Select(courseRequestWithCourseSelect...)).
		Where(sq.Eq{"r.deleted_at": nil}).
		Where(sq.Eq{"r.id": requestID})
}

// GetCourseRequestsByFaculty lists requests with their course. An empty faculty lists every
// faculty (DEU-wide admin).
func GetCourseRequestsByFaculty(faculty string, limit, offset int) sq.SelectBuilder {
	q := fromCourseRequestsWithCourse(psql.Select(courseRequestWithCourseSelect...)).
		Where(sq.Eq{"r.deleted_at": nil})
	if faculty != "" {
		q = q.Where(sq.Eq{"c.faculty": faculty})
	}
	return q.OrderBy("r.created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))
}

func CountCourseRequestsByFaculty(faculty string) sq.SelectBuilder {
	q := psql.Select("COUNT(*)").
		From(fmt.Sprintf("%s AS r", courseRequestsTableName)).
		Join(fmt.Sprintf("%s AS c ON c.id = r.course_id", coursesTableName)).
		Where(sq.Eq{"r.deleted_at": nil})
	if faculty != "" {
		q = q.Where(sq.Eq{"c.faculty": faculty})
	}
	return q
}

func GetCourseRequestsByProvider(providerID string, limit, offset int) sq.SelectBuilder {
	return fromCourseRequestsWithCourse(psql.Select(courseRequestWithCourseSelect...)).
		Where(sq.Eq{"c.provider_id": providerID}).
		Where(sq.Eq{"r.deleted_at": nil}).
		OrderBy("r.created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))
}

func CountCourseRequestsByProvider(providerID string) sq.SelectBuilder {
	return psql.Select("COUNT(*)").
		From(fmt.Sprintf("%s AS r", courseRequestsTableName)).
		Join(fmt.Sprintf("%s AS c ON c.id = r.course_id", coursesTableName)).
		Where(sq.Eq{"c.provider_id": providerID}).
		Where(sq.Eq{"r.deleted_at": nil})
}

func ApproveCourseRequest(requestID, reviewerID, comments string, score sql.NullFloat64, classification sql.NullString) sq.UpdateBuilder {
	return psql.Update(courseRequestsTableName).
		Set("status", "approved").
		Set("reviewer_id", reviewerID).
		Set("comments", comments).
		Set("score", score).
		Set("classification", classification).
		Set("reviewed_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": requestID, "status": "under_review"})
}

func RejectCourseRequest(requestID, reviewerID, comments string) sq.UpdateBuilder {
	return psql.Update(courseRequestsTableName).
		Set("status", "rejected").
		Set("reviewer_id", reviewerID).
		Set("comments", comments).
		Set("reviewed_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": requestID, "status": "under_review"})
}

func RedirectCourseRequest(requestID, reviewerID, reason string) sq.UpdateBuilder {
	return psql.Update(courseRequestsTableName).
		Set("reviewer_id", reviewerID).
		Set("comments", reason).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": requestID, "status": "under_review"})
}

func UpdateCourseType(courseID, courseType string, active bool) sq.UpdateBuilder {
	return psql.Update(coursesTableName).
		Set("type", courseType).
		Set("is_active", active).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": courseID})
}

func UpdateCourseFaculty(courseID, faculty string) sq.UpdateBuilder {
	return psql.Update(coursesTableName).
		Set("faculty", faculty).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": courseID})
}

func GetCourseRequestByCourseID(courseID string) sq.SelectBuilder {
	return psql.Select(courseRequestQuerySelectCommon...).
		From(fmt.Sprintf("%s AS r", courseRequestsTableName)).
		Where(sq.Eq{"r.course_id": courseID})
}

func GetCourseIDFromRequest(requestID string) sq.SelectBuilder {
	return psql.Select("course_id").
		From(courseRequestsTableName).
		Where(sq.Eq{"id": requestID})
}

func InsertCourseRequest(request models.CourseRequest) sq.InsertBuilder {
	return psql.Insert(courseRequestsTableName).
		Columns(
			"course_id",
			"status",
		).
		Values(
			request.CourseID,
			request.Status,
		).Suffix("RETURNING id")
}
