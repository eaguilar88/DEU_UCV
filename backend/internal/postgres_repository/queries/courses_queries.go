package queries

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
)

var courseQuerySelectCommon = []string{
	"c.id",
	"c.name",
	"c.description",
	"c.endorsement_id",
	"requester.id",
	"requester.first_name",
	"requester.last_name",
	"reviewer.id",
	"reviewer.first_name",
	"reviewer.last_name",
	"c.content",
	"c.objectives",
	"c.cost",
	"c.location",
	"c.created_at",
	"c.updated_at",
	"c.deleted_at",
}

func GetCourseByID(courseID string) sq.SelectBuilder {
	return psql.Select(courseQuerySelectCommon...).
		From(fmt.Sprintf("%s AS c", coursesTableName)).
		Join(fmt.Sprintf("%s AS r ON c.endorsement_id = r.id", endorsementsTableName)).
		Join(fmt.Sprintf("%s AS requester ON c.user_id = requester.id", usersTableName)).
		Join(fmt.Sprintf("%s AS reviewer ON r.reviewer_id = reviewer.id", usersTableName)).
		Where(sq.Eq{"e.id": courseID})
}

func GetCourses(limit, offset int) sq.SelectBuilder {
	return psql.Select(courseQuerySelectCommon...).
		From(fmt.Sprintf("%s AS c", coursesTableName)).
		Join(fmt.Sprintf("%s AS r ON c.endorsement_id = r.id", endorsementsTableName)).
		Join(fmt.Sprintf("%s AS requester ON r.user_id = requester.id", usersTableName)).
		Join(fmt.Sprintf("%s AS reviewer ON r.reviewer_id = reviewer.id", usersTableName)).
		Limit(uint64(limit)).
		Offset(uint64(offset))
}

func InsertCourse(course models.Course) sq.InsertBuilder {
	return psql.Insert(entitiesTableName).
		Columns(
			"name",
			"description",
			"user_id",
			"endorsement_id",
			"objectives",
			"content",
			"cost",
			"location",
		).
		Values(
			course.Name,
			course.Description,
			course.OwnerID,
			course.EndorsementID,
			course.Objectives,
			course.Content,
			course.Cost,
			course.Location,
		).Suffix("RETURNING id")
}

func UpdateCourse(courseID string, course models.Course) sq.UpdateBuilder {
	return psql.Update(coursesTableName).
		Set("name", course.Name).
		Set("description", course.Description.String).
		Set("objectives", course.Objectives.String).
		Set("content", course.Content).
		Set("cost", course.Cost).
		Set("location", course.Location).
		Where(sq.Eq{"id": courseID})
}

func DeleteCourse(courseID string) sq.DeleteBuilder {
	return psql.Delete(coursesTableName).
		Where(sq.Eq{"id": courseID})
}
