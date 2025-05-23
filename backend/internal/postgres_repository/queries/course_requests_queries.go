package queries

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
)

var courseRequestQuerySelectCommon = []string{
	"r.id",
	"r.name",
	"r.description",
	"r.type",
	"r.status",
	"r.comments",
	"owner.id owner_id",
	"owner.first_name",
	"owner.last_name",
	"reviewer.id reviewer_id",
	"reviewer.first_name",
	"reviewer.last_name",
	"r.reviewed_at",
	"r.created_at",
	"r.updated_at",
}

func GetCourseRequestByID(requestID string) sq.SelectBuilder {
	return psql.Select(courseRequestQuerySelectCommon...).
		From(fmt.Sprintf("%s AS r", courseRequestsTableName)).
		Join(fmt.Sprintf("%s AS owner ON r.user_id = owner.id", usersTableName)).
		Join(fmt.Sprintf("%s AS reviewer ON r.reviewer_id = reviewer.id", usersTableName)).
		Where(sq.Eq{"r.id": requestID})
}

func GetCourseRequests(page entities.PageScope) sq.SelectBuilder {
	return psql.Select(courseRequestQuerySelectCommon...).
		From(fmt.Sprintf("%s AS r", courseRequestsTableName)).
		Join(fmt.Sprintf("%s AS owner ON r.user_id = owner.id", usersTableName)).
		Join(fmt.Sprintf("%s AS reviewer ON r.reviewer_id = reviewer.id", usersTableName)).
		Limit(uint64(page.PerPage)).
		Offset(uint64(page.Offset()))
}

func InsertCourseRequest(request models.CourseRequest) sq.InsertBuilder {
	return psql.Insert(courseRequestsTableName).
		Columns(
			"user_id",
			"type",
			"name",
			"description",
			"status",
		).
		Values(
			request.UserID,
			request.Type,
			request.Name.String,
			request.Description.String,
			entities.CourseRequestStatus_CREATED,
		).Suffix("RETURNING id")
}

func UpdateCourseRequestInfo(
	request models.CourseRequest,
	requestID string,
) sq.UpdateBuilder {
	return psql.Update(courseRequestsTableName).
		Set("user_id", request.UserID).
		Set("status", request.Status).
		Set("type", request.Type).
		Set("name", request.Name.String).
		Set("description", request.Description.String).
		Set("comments", request.Comments.String).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"r.id": requestID})
}

func DeleteCourseRequest(requestID string) sq.DeleteBuilder {
	return psql.Delete(courseRequestsTableName).
		Where(sq.Eq{"r.id": requestID})
}
