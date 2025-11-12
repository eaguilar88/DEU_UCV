package queries

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
)

var periodQuerySelectCommon = []string{
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

func GetCoursePeriodByID(periodID string) sq.SelectBuilder {
	return psql.Select(periodQuerySelectCommon...).
		From(fmt.Sprintf("%s AS cp", periodsTableName)).
		Where(sq.Eq{"cp.is_active": true}).
		Where(sq.Eq{"cp.id": periodID})
}

func GetCoursePeriods(courseID string, page entities.PageScope) sq.SelectBuilder {
	return psql.Select(periodQuerySelectCommon...).
		From(fmt.Sprintf("%s AS cp", periodsTableName)).
		Where(sq.Eq{"cp.course_id": courseID}).
		Where(sq.Eq{"cp.is_active": true}).
		Limit(uint64(page.PerPage)).
		Offset(uint64(page.Offset()))
}

func GetLatestCoursePeriod(courseID string) sq.SelectBuilder {
	return psql.Select("cp.id", "cp.start_date", "cp.end_date", "cp.inscription_date").
		From(fmt.Sprintf("%s AS cp", periodsTableName)).
		Where(sq.Eq{"cp.course_id": courseID}).
		Where(sq.Eq{"cp.is_active": true}).
		OrderBy("cp.created_at DESC").
		Limit(1)
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

func UpdateCoursePeriod(periodID string, coursePeriod models.CoursePeriod) sq.UpdateBuilder {
	return psql.Update(periodsTableName).
		Set("course_id", coursePeriod.CourseID).
		Set("start_date", coursePeriod.StartDate).
		Set("end_date", coursePeriod.EndDate).
		Set("inscription_date", coursePeriod.InscriptionDate).
		Where(sq.Eq{"id": periodID})
}

func DeleteCoursePeriod(periodID string) sq.UpdateBuilder {
	return sq.Update(periodsTableName).
		Set("deleted_at", "NOW()").
		Set("is_active", false).
		Where(sq.Eq{"id": periodID})
}

// Announcement queries
var announcementQuerySelectCommon = []string{
	"a.id",
	"a.course_cycle_id",
	"a.title",
	"a.content",
	"a.created_at",
	"a.updated_at",
	"a.deleted_at",
}

func GetAnnouncementByID(announcementID string) sq.SelectBuilder {
	return psql.Select(announcementQuerySelectCommon...).
		From(fmt.Sprintf("%s AS a", announcementsTableName)).
		Where(sq.Eq{"a.id": announcementID}).
		Where(sq.Eq{"a.deleted_at": nil})
}

func GetAnnouncementsByCoursePeriodID(coursePeriodID string) sq.SelectBuilder {
	return psql.Select(announcementQuerySelectCommon...).
		From(fmt.Sprintf("%s AS a", announcementsTableName)).
		Where(sq.Eq{"a.course_cycle_id": coursePeriodID}).
		Where(sq.Eq{"a.deleted_at": nil}).
		OrderBy("a.created_at DESC")
}

func InsertAnnouncement(announcement models.Announcement) sq.InsertBuilder {
	return psql.Insert(announcementsTableName).
		Columns(
			"course_cycle_id",
			"title",
			"content",
		).
		Values(
			announcement.CourseCycleID,
			announcement.Title,
			announcement.Content,
		).Suffix("RETURNING id")
}

func UpdateAnnouncement(announcementID string, announcement models.Announcement) sq.UpdateBuilder {
	return psql.Update(announcementsTableName).
		Set("title", announcement.Title).
		Set("content", announcement.Content).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": announcementID}).
		Where(sq.Eq{"deleted_at": nil})
}

func DeleteAnnouncement(announcementID string) sq.UpdateBuilder {
	return psql.Update(announcementsTableName).
		Set("deleted_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": announcementID})
}
