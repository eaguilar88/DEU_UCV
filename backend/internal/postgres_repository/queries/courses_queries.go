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
	"c.deleted_at",
}

func GetCourseByID(courseID string) sq.SelectBuilder {
	return psql.Select(courseQuerySelectCommon...).
		From(fmt.Sprintf("%s AS c", coursesTableName)).
		Where(sq.Eq{"c.deleted_at": nil}).
		Where(sq.Eq{"c.is_active": true}).
		Where(sq.Eq{"c.id": courseID})
}

func GetCourses(limit, offset int) sq.SelectBuilder {
	return psql.Select(courseQuerySelectCommon...).
		From(fmt.Sprintf("%s AS c", coursesTableName)).
		Where(sq.Eq{"c.deleted_at": nil}).
		Where(sq.Eq{"c.is_active": true}).
		Limit(uint64(limit)).
		Offset(uint64(offset))
}

func InsertCourse(course models.Course) sq.InsertBuilder {
	return psql.Insert(coursesTableName).
		Columns(
			"name",
			"description",
			"provider_id",
			"objectives",
			"duration",
			"content",
			"faculty",
			"cost",
			"location",
		).
		Values(
			course.Name,
			course.Description,
			course.OwnerID,
			course.Objectives,
			course.Duration,
			course.Content,
			course.Faculty,
			course.Cost,
			course.Location,
		).Suffix("RETURNING id")
}

func UpdateCourse(courseID string, course models.Course) sq.UpdateBuilder {
	return psql.Update(coursesTableName).
		Set("name", course.Name).
		Set("description", course.Description.String).
		Set("objectives", course.Objectives.String).
		Set("duration", course.Duration).
		Set("content", course.Content).
		Set("type", course.Type).
		Set("faculty", course.Faculty).
		Set("cost", course.Cost).
		Set("location", course.Location).
		Where(sq.Eq{"id": courseID})
}

func DeleteCourse(courseID string) sq.DeleteBuilder {
	return psql.Delete(coursesTableName).
		Where(sq.Eq{"id": courseID})
}
