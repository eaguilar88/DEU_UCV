package queries

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/pkg/entities"
)

var (
	periodQuerySelectCommon = []string{
		"cp.id",
		"cp.course_id",
		"e.name",
		"e.description",
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
		Join(fmt.Sprintf("%s AS e ON cp.course_id = e.id", entitiesTableName)).
		Where(sq.Eq{"cp.id": periodID})
}

func GetCoursePeriods(page entities.PageScope) sq.SelectBuilder {
	return psql.Select(periodQuerySelectCommon...).
		From(fmt.Sprintf("%s AS cp", periodsTableName)).
		Join(fmt.Sprintf("%s AS c ON cp.course_id = c.id", coursesTableName)).
		Limit(uint64(page.PerPage)).
		Offset(uint64(page.Offset()))
}

func InsertCoursePeriod(coursePeriod entities.CoursePeriod) sq.InsertBuilder {
	return psql.Insert(periodsTableName).
		Columns(
			"course_id",
			"start_date",
			"end_date",
			"inscription_date",
		).
		Values(
			coursePeriod.Course.ID,
			coursePeriod.StartDate,
			coursePeriod.EndDate,
			coursePeriod.InscriptionDate,
		)
}

func UpdateCoursePeriod(periodID int, coursePeriod entities.CoursePeriod) sq.UpdateBuilder {
	return psql.Update(periodsTableName).
		Set("course_id", coursePeriod.Course.ID).
		Set("start_date", coursePeriod.StartDate).
		Set("end_date", coursePeriod.EndDate).
		Set("inscription_date", coursePeriod.InscriptionDate).
		Where(sq.Eq{"id": periodID})
}

func DeleteCoursePeriod(periodID int) sq.DeleteBuilder {
	return psql.Delete(periodsTableName).
		Where(sq.Eq{"id": periodID})
}
