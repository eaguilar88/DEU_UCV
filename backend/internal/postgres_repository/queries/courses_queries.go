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
	"c.rationale",
	"c.duration",
	"c.cost",
	"c.instructor_profile",
	"c.profiles",
	"c.requirements",
	"c.content",
	"c.evaluation",
	"c.schedule",
	"c.type",
	"c.faculty",
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
			"rationale",
			"duration",
			"cost",
			"instructor_profile",
			"profiles",
			"requirements",
			"content",
			"evaluation",
			"schedule",
			"type",
			"faculty",
			"location",
		).
		Values(
			course.Name,
			course.Description,
			course.OwnerID,
			course.Objectives,
			course.Rationale,
			course.Duration,
			course.Cost,
			course.InstructorProfile,
			course.Profiles,
			course.Requirements,
			course.Content,
			course.Evaluation,
			course.Schedule,
			course.Type,
			course.Faculty,
			course.Location,
		).Suffix("RETURNING id")
}

func UpdateCourse(courseID string, course models.Course) sq.UpdateBuilder {
	return psql.Update(coursesTableName).
		Set("name", course.Name).
		Set("description", course.Description).
		Set("objectives", course.Objectives).
		Set("rationale", course.Rationale).
		Set("duration", course.Duration).
		Set("cost", course.Cost).
		Set("instructor_profile", course.InstructorProfile).
		Set("profiles", course.Profiles).
		Set("requirements", course.Requirements).
		Set("content", course.Content).
		Set("evaluation", course.Evaluation).
		Set("schedule", course.Schedule).
		Set("type", course.Type).
		Set("faculty", course.Faculty).
		Set("location", course.Location).
		Where(sq.Eq{"id": courseID})
}

func DeleteCourse(courseID string) sq.DeleteBuilder {
	return psql.Delete(coursesTableName).
		Where(sq.Eq{"id": courseID})
}
