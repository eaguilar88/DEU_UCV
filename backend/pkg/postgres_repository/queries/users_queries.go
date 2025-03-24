package queries

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/eaguilar88/deu/pkg/postgres_repository/models"
)

const (
	schema = "deu"
)

var (
	userQuerySelectCommon = []string{
		"u.id", "u.ci", "u.username", "u.first_name", "u.last_name", "u.date_of_birth", "u.gender", "u.education", "u.address", "u.created_at", "p.code",
	}
)

func GetRolesByUserID(userID int) sq.SelectBuilder {
	return psql.Select("r.name").
		From(fmt.Sprintf("%s AS r", rolesTableName)).
		LeftJoin(fmt.Sprintf("%s AS pr ON pr.role_id = r.id", pivotTableName)).
		Where(sq.Eq{"pr.user_id": userID})
}

func GetUserByUsername(username string) sq.SelectBuilder {
	return psql.Select("u.id", "u.username", "u.first_name", "u.last_name", "u.password").
		From(fmt.Sprintf("%s AS u", usersTableName)).
		Where(sq.Eq{"u.username": username})
}

func GetUserByID(userID int) sq.SelectBuilder {
	return psql.Select(userQuerySelectCommon...).
		From(fmt.Sprintf("%s AS u", usersTableName)).
		LeftJoin(fmt.Sprintf("%s AS p ON p.user_id = u.id", providersTableName)).
		Where(sq.Eq{"u.id": userID})
}

func GetUsers(page entities.PageScope) sq.SelectBuilder {
	return psql.Select(userQuerySelectCommon...).
		From(fmt.Sprintf("%s AS u", usersTableName)).
		Limit(uint64(page.PerPage)).
		LeftJoin(fmt.Sprintf("%s AS p ON p.user_id = u.id", providersTableName)).
		Offset(uint64(page.Offset()))
}

func GetUsersByCoursePeriodID(coursePeriodID int) sq.SelectBuilder {
	return psql.Select(userQuerySelectCommon...).
		From(fmt.Sprintf("%s AS u", usersTableName)).
		Join(fmt.Sprintf("%s AS up ON up.user_id = u.id", participantsTableName)).
		Where(sq.Eq{"up.course_period_id": coursePeriodID})
}

func InsertUser(user models.User) sq.InsertBuilder {
	return psql.Insert(usersTableName).
		Columns(
			"ci",
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
			user.CI,
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

func AddRoleToUser(userID, role int) sq.InsertBuilder {
	return psql.Insert(pivotTableName).
		Columns("user_id", "role_id").
		Values(userID, role).
		Suffix("RETURNING id")
}

func UpdateUserInfo(user entities.User, userID int) sq.UpdateBuilder {
	return psql.Update(usersTableName).
		Set("first_name", user.FirstName).
		Set("last_name", user.LastName).
		Set("date_of_birth", user.DateOfBirth).
		Set("gender", user.Gender).
		Set("address", user.Address).
		Set("education", user.EducationLevel).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": userID})
}

func UpdateUsername(user entities.User, userID int) sq.UpdateBuilder {
	return psql.Update(usersTableName).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": userID})
}

func UpdateUserPassword(user entities.User, userID int) sq.UpdateBuilder {
	return psql.Update(fmt.Sprintf("%s AS u", usersTableName)).
		Set("u.updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"u.id": userID})
}

func DeleteUser(userID int) sq.DeleteBuilder {
	return psql.Delete(usersTableName).
		Where(sq.Eq{"id": userID})
}

func splitUserCI(user entities.User) (string, string) {
	ciType := strings.Split(user.CI, "-")[0]
	ciNumber := strings.Split(user.CI, "-")[1]
	return ciType, ciNumber
}
