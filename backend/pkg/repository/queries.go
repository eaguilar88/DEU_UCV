package repository

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/pkg/entities"
)

var (
	psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
)

func getUserByID(userID int) sq.SelectBuilder {
	return psql.Select("u.id", "u.first_name", "u.last_name", "u.date_of_birth", "u.gender", "u.education", "u.address").
		From("deu.users u").
		Where(sq.Eq{"u.id": userID})
}

func getUsers(page entities.PageScope) sq.SelectBuilder {
	return psql.Select("u.id", "u.first_name", "u.last_name", "u.date_of_birth", "u.gender", "u.education", "u.address").
		From("users u").
		Limit(uint64(page.PerPage)).
		Offset(uint64(page.Offset()))
}

func insertUser(user entities.User) sq.InsertBuilder {
	return psql.Insert("users").
		Columns(
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
		)
}

func updateUserInfo(user entities.User, userID int) sq.UpdateBuilder {
	return psql.Update("users u").
		Set("u.first_name", user.FirstName).
		Set("u.last_name", user.LastName).
		Set("u.date_of_birth", user.DateOfBirth).
		Set("u.gender", user.Gender).
		Set("u.address", user.Address).
		Set("u.education", user.EducationLevel).
		Set("u.updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"u.id": userID})
}

func updateUsername(user entities.User, userID int) sq.UpdateBuilder {
	return psql.Update("users u").
		Set("u.updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"u.id": userID})
}

func updateUserPassword(user entities.User, userID int) sq.UpdateBuilder {
	return psql.Update("users u").
		Set("u.updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"u.id": userID})
}

func deleteUser(userID int) sq.DeleteBuilder {
	return psql.Delete("users").
		Where(sq.Eq{"id": userID})
}
