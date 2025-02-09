package queries

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/eaguilar88/deu/pkg/repository/models"
)

var (
	courseTableName         = fmt.Sprintf("%s.courses", schema)
	courseQuerySelectCommon = []string{
		"c.id", "c.user_id", "c.endorsement_id", "c.endorsed_by", "c.objectives", "c.content", "c.cost", "c.location", "c.created_at",
	}
)

func GetCourseByID(userID int) sq.SelectBuilder {
	return psql.Select(courseQuerySelectCommon...).
		From(courseTableName).
		Where(sq.Eq{"id": userID})
}

func GetCourses(page entities.PageScope) sq.SelectBuilder {
	return psql.Select(courseQuerySelectCommon...).
		From(courseTableName).
		Limit(uint64(page.PerPage)).
		Offset(uint64(page.Offset()))
}

func InsertCourse(course models.Course) sq.InsertBuilder {
	return psql.Insert(courseTableName).
		Columns(
			"user_id",
			"endorsement_id",
			"endorsed_by",
			"objectives",
			"content",
			"cost",
			"location",
			"created_at",
			"updated_at",
		).
		Values(
			course.UserID,
			course.EndorsementID,
			course.EndorsedBy,
			course.Objectives,
			course.Content,
			course.Cost,
			course.Location,
			sq.Expr("NOW()"),
			sq.Expr("NOW()"),
		).Suffix("RETURNING id")
}

func DeleteCourse(endorsementID int) sq.DeleteBuilder {
	return psql.Delete(courseTableName).
		Where(sq.Eq{"id": endorsementID})
}
