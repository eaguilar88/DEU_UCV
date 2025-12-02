package queries

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
)

var userQuerySelectCommon = []string{
	"u.id",
	"u.ci",
	"u.email",
	"u.first_name",
	"u.last_name",
	"u.date_of_birth",
	"u.gender",
	"u.education",
	"u.address",
	"u.created_at",
	"p.code",
}

func GetRolesByUserID(userID string) sq.SelectBuilder {
	return psql.Select("r.name").
		From(fmt.Sprintf("%s AS r", rolesTableName)).
		LeftJoin(fmt.Sprintf("%s AS pr ON pr.role_id = r.id", pivotTableName)).
		Where(sq.Eq{"r.deleted_at": nil}).
		Where(sq.Eq{"pr.user_id": userID})
}

func GetUserByUsername(username string) sq.SelectBuilder {
	return psql.Select("u.id", "u.email", "u.first_name", "u.last_name", "u.password").
		From(fmt.Sprintf("%s AS u", usersTableName)).
		Where(sq.Eq{"u.deleted_at": nil}).
		Where(sq.Eq{"u.email": username})
}

func GetUserByID(userID string) sq.SelectBuilder {
	return psql.Select(userQuerySelectCommon...).
		From(fmt.Sprintf("%s AS u", usersTableName)).
		LeftJoin(fmt.Sprintf("%s AS p ON p.user_id = u.id", providersTableName)).
		Where(sq.Eq{"u.deleted_at": nil}).
		Where(sq.Eq{"u.id": userID})
}

func GetUsers(limit, offset int) sq.SelectBuilder {
	return psql.Select(userQuerySelectCommon...).
		From(fmt.Sprintf("%s AS u", usersTableName)).
		Limit(uint64(limit)).
		LeftJoin(fmt.Sprintf("%s AS p ON p.user_id = u.id", providersTableName)).
		Where(sq.Eq{"u.deleted_at": nil}).
		Offset(uint64(offset))
}

func GetUsersByCoursePeriodID(coursePeriodID string) sq.SelectBuilder {
	return psql.Select("u.id", "u.ci", "u.first_name", "u.last_name").
		From(fmt.Sprintf("%s AS u", usersTableName)).
		Join(fmt.Sprintf("%s AS up ON up.user_id = u.id", participantsTableName)).
		Where(sq.Eq{"u.deleted_at": nil}).
		Where(sq.Eq{"up.id": coursePeriodID})
}

func InsertUser(user models.User) sq.InsertBuilder {
	return psql.Insert(usersTableName).
		Columns(
			"ci",
			"email",
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
			user.Email,
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

func AddRoleToUser(userID string, role int) sq.InsertBuilder {
	return psql.Insert(pivotTableName).
		Columns("user_id", "role_id", "domain_type").
		Values(userID, role, "course").
		Suffix("RETURNING id")
}

func UpdateUserInfo(user entities.User, userID string) sq.UpdateBuilder {
	return psql.Update(usersTableName).
		Set("first_name", user.FirstName).
		Set("last_name", user.LastName).
		Set("date_of_birth", user.DateOfBirth).
		Set("gender", user.Gender).
		Set("address", user.Address).
		Set("education", user.EducationLevel).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"u.deleted_at": nil}).
		Where(sq.Eq{"id": userID})
}

func UpdateUsername(user entities.User, userID string) sq.UpdateBuilder {
	return psql.Update(usersTableName).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"u.deleted_at": nil}).
		Where(sq.Eq{"id": userID})
}

func UpdateUserPassword(userID string, password string) sq.UpdateBuilder {
	return psql.Update(fmt.Sprintf("%s AS u", usersTableName)).
		Set("password", password).
		Set("u.updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"u.deleted_at": nil}).
		Where(sq.Eq{"u.id": userID})
}

func SoftDeleteUser(userID string) sq.UpdateBuilder {
	return psql.Update(usersTableName).
		Set("deleted_at", sq.Expr("NOW()")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"deleted_at": nil}).
		Where(sq.Eq{"id": userID})
}
