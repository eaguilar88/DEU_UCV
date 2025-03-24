package queries

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/eaguilar88/deu/pkg/postgres_repository/models"
)

var (
	periodQuerySelectCommon = []string{
		"cp.id",
		"cp.course_id",
		"cp.start_date",
		"cp.end_date",
		"cp.is_active",
		"cp.inscription_date",
		"cp.created_at",
		"cp.updated_at",
		"cp.deleted_at",
	}
)

func GetCoursePeriodByID(periodID int) sq.SelectBuilder {
	return psql.Select(periodQuerySelectCommon...).
		From(fmt.Sprintf("%s AS cp", periodsTableName)).
		Where(sq.Eq{"cp.is_active": true}).
		Where(sq.Eq{"cp.id": periodID})
}

func GetCoursePeriods(courseID int, page entities.PageScope) sq.SelectBuilder {
	return psql.Select(periodQuerySelectCommon...).
		From(fmt.Sprintf("%s AS cp", periodsTableName)).
		Where(sq.Eq{"cp.course_id": courseID}).
		Where(sq.Eq{"cp.is_active": true}).
		Limit(uint64(page.PerPage)).
		Offset(uint64(page.Offset()))
}

func InsertCoursePeriod(coursePeriod models.CoursePeriod) sq.InsertBuilder {
	return psql.Insert(periodsTableName).
		Columns(
			"course_id",
			"start_date",
			"end_date",
			"inscription_date",
		).
		Values(
			coursePeriod.CourseID,
			coursePeriod.StartDate,
			coursePeriod.EndDate,
			coursePeriod.InscriptionDate,
		).Suffix("RETURNING id")
}

func UpdateCoursePeriod(periodID int, coursePeriod models.CoursePeriod) sq.UpdateBuilder {
	return psql.Update(periodsTableName).
		Set("course_id", coursePeriod.CourseID).
		Set("start_date", coursePeriod.StartDate).
		Set("end_date", coursePeriod.EndDate).
		Set("inscription_date", coursePeriod.InscriptionDate).
		Where(sq.Eq{"id": periodID})
}

func DeleteCoursePeriod(periodID int) sq.UpdateBuilder {
	return sq.Update(periodsTableName).
		Set("deleted_at", "NOW()").
		Set("is_active", false).
		Where(sq.Eq{"id": periodID})
}
