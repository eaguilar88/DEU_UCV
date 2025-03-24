package queries

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/eaguilar88/deu/pkg/postgres_repository/models"
)

var (
	courseQuerySelectCommon = []string{
		"c.id",
		"e.name",
		"e.description",
		"e.endorsement_id",
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
		"e.created_at",
		"e.updated_at",
	}
)

func GetCourseByID(courseID int) sq.SelectBuilder {
	return psql.Select(courseQuerySelectCommon...).
		From(fmt.Sprintf("%s AS e", entitiesTableName)).
		Join(fmt.Sprintf("%s AS c ON e.id = c.entity_id", coursesTableName)).
		Join(fmt.Sprintf("%s AS r ON e.endorsement_id = r.id", endorsementsTableName)).
		Join(fmt.Sprintf("%s AS requester ON r.user_id = requester.id", usersTableName)).
		Join(fmt.Sprintf("%s AS reviewer ON r.reviewer_id = reviewer.id", usersTableName)).
		Where(sq.Eq{"e.id": courseID})
}

func GetCourses(page entities.PageScope) sq.SelectBuilder {
	return psql.Select(courseQuerySelectCommon...).
		From(fmt.Sprintf("%s AS e", entitiesTableName)).
		Join(fmt.Sprintf("%s AS c ON e.id = c.entity_id", coursesTableName)).
		Join(fmt.Sprintf("%s AS r ON e.endorsement_id = r.id", endorsementsTableName)).
		Join(fmt.Sprintf("%s AS requester ON r.user_id = requester.id", usersTableName)).
		Join(fmt.Sprintf("%s AS reviewer ON r.reviewer_id = reviewer.id", usersTableName)).
		Limit(uint64(page.PerPage)).
		Offset(uint64(page.Offset()))
}

func InsertCourse(course models.Course) sq.InsertBuilder {
	return psql.Insert(coursesTableName).
		Columns(
			"user_id",
			"endorsement_id",
			"endorsed_by",
			"objectives",
			"content",
			"cost",
			"location",
		).
		Values(
			course.OwnerID,
			course.EndorsementID,
			course.EndorserID,
			course.Objectives,
			course.Content,
			course.Cost,
			course.Location,
		).Suffix("RETURNING id")
}

func DeleteCourse(endorsementID int) sq.DeleteBuilder {
	return psql.Delete(coursesTableName).
		Where(sq.Eq{"id": endorsementID})
}
