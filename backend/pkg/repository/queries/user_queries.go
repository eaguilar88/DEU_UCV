package queries

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/pkg/entities"
)

const (
	schema = "deu"
)

var (
	psql                  = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	userTableName         = fmt.Sprintf("%s.users", schema)
	userQuerySelectCommon = []string{
		"u.id", "u.ci_type", "u.ci", "u.username", "u.first_name", "u.last_name", "u.date_of_birth", "u.gender", "u.education", "u.address", "u.created_at",
	}
)

func GetUserByID(userID int) sq.SelectBuilder {
	return psql.Select(userQuerySelectCommon...).
		From(fmt.Sprintf("%s AS u", userTableName)).
		Where(sq.Eq{"u.id": userID})
}

func GetUsers(page entities.PageScope) sq.SelectBuilder {
	return psql.Select(userQuerySelectCommon...).
		From(fmt.Sprintf("%s AS u", userTableName)).
		Limit(uint64(page.PerPage)).
		Offset(uint64(page.Offset()))
}

func InsertUser(user entities.User) sq.InsertBuilder {
	ciType, ciNumber := splitUserCI(user)

	return psql.Insert(userTableName).
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

func UpdateUserInfo(user entities.User, userID int) sq.UpdateBuilder {
	return psql.Update(userTableName).
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
	return psql.Update(userTableName).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": userID})
}

func UpdateUserPassword(user entities.User, userID int) sq.UpdateBuilder {
	return psql.Update(fmt.Sprintf("%s AS u", userTableName)).
		Set("u.updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"u.id": userID})
}

func DeleteUser(userID int) sq.DeleteBuilder {
	return psql.Delete(userTableName).
		Where(sq.Eq{"id": userID})
}

func splitUserCI(user entities.User) (string, string) {
	ciType := strings.Split(user.CI, "-")[0]
	ciNumber := strings.Split(user.CI, "-")[1]
	return ciType, ciNumber
}
