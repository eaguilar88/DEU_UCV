package queries

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/pkg/entities"
)

var (
	courseTableName         = fmt.Sprintf("%s.endorsement_requests", schema)
	courseQuerySelectCommon = []string{
		"u.id", "u.ci_type", "u.ci", "u.username", "u.first_name", "u.last_name", "u.date_of_birth", "u.gender", "u.education", "u.address", "u.created_at",
	}
)

func GetCourseByID(userID int) sq.SelectBuilder {
	return psql.Select(courseQuerySelectCommon...).
		From(fmt.Sprintf("%s AS u", courseTableName)).
		Where(sq.Eq{"u.id": userID})
}

func GetCourses(page entities.PageScope) sq.SelectBuilder {
	return psql.Select(courseQuerySelectCommon...).
		From(fmt.Sprintf("%s AS u", courseTableName)).
		Limit(uint64(page.PerPage)).
		Offset(uint64(page.Offset()))
}

func InsertCourse(user entities.User) sq.InsertBuilder {
	ciType, ciNumber := splitUserCI(user)

	return psql.Insert(courseTableName).
		Columns(
			"ci",
			"ci_type",
			"username",
			"first_name",
			"last_name",
			"date_of_birth",
			"gender",
			"education",
			"address",
			"password",
			"created_at",
			"updated_at",
		).
		Values(
			ciNumber,
			ciType,
			user.Username,
			user.FirstName,
			user.LastName,
			user.DateOfBirth,
			user.Gender,
			user.EducationLevel,
			user.Address,
			user.Password,
			sq.Expr("NOW()"),
			sq.Expr("NOW()"),
		).Suffix("RETURNING id")
}

func DeleteCourse(endorsementID int) sq.DeleteBuilder {
	return psql.Delete(courseTableName).
		Where(sq.Eq{"id": endorsementID})
}
