package queries

import (
	"fmt"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
)

// activeProviderJoin joins the user's provider only while it is approved, so a pending or
// rejected provider has no provider code.
var activeProviderJoin = fmt.Sprintf("%s AS p ON p.user_id = u.id AND p.status = ? AND p.deleted_at IS NULL", providersTableName)

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
	return psql.Select("r.name", "pr.domain_type", "COALESCE(pr.faculty::TEXT, '')").
		From(fmt.Sprintf("%s AS r", rolesTableName)).
		LeftJoin(fmt.Sprintf("%s AS pr ON pr.role_id = r.id", pivotTableName)).
		Where(sq.Eq{"r.deleted_at": nil}).
		Where(sq.Eq{"pr.user_id": userID})
}

// GetEmailsByRole selects the emails of the users holding roleName. A non-empty faculty keeps only
// the assignments scoped to that faculty.
func GetEmailsByRole(roleName string, faculty entities.Faculty) sq.SelectBuilder {
	query := psql.Select("DISTINCT u.email").
		From(fmt.Sprintf("%s AS u", usersTableName)).
		Join(fmt.Sprintf("%s AS pr ON pr.user_id = u.id", pivotTableName)).
		Join(fmt.Sprintf("%s AS r ON r.id = pr.role_id", rolesTableName)).
		Where(sq.Eq{"u.deleted_at": nil}).
		Where(sq.Eq{"pr.deleted_at": nil}).
		Where(sq.Eq{"r.deleted_at": nil}).
		Where(sq.Eq{"r.name": roleName})
	if faculty != "" {
		query = query.Where(sq.Eq{"pr.faculty": string(faculty)})
	}
	return query
}

func GetUserByUsername(username string) sq.SelectBuilder {
	return psql.Select("u.id", "u.email", "u.first_name", "u.last_name", "u.password").
		From(fmt.Sprintf("%s AS u", usersTableName)).
		Where(sq.Eq{"u.deleted_at": nil}).
		Where(sq.Eq{"u.email": username})
}

// GetUserCredentialsByID selects the same columns as GetUserByUsername, including the password hash.
func GetUserCredentialsByID(userID string) sq.SelectBuilder {
	return psql.Select("u.id", "u.email", "u.first_name", "u.last_name", "u.password").
		From(fmt.Sprintf("%s AS u", usersTableName)).
		Where(sq.Eq{"u.deleted_at": nil}).
		Where(sq.Eq{"u.id": userID})
}

func GetUserByID(userID string) sq.SelectBuilder {
	return psql.Select(userQuerySelectCommon...).
		From(fmt.Sprintf("%s AS u", usersTableName)).
		LeftJoin(activeProviderJoin, entities.ProviderStatusActive).
		Where(sq.Eq{"u.deleted_at": nil}).
		Where(sq.Eq{"u.id": userID})
}

func GetUsers(limit, offset int) sq.SelectBuilder {
	return psql.Select(userQuerySelectCommon...).
		From(fmt.Sprintf("%s AS u", usersTableName)).
		Limit(uint64(limit)).
		LeftJoin(activeProviderJoin, entities.ProviderStatusActive).
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

// UpdateUserInfo sets only the fields present in changes. An empty gender or address is
// stored as NULL.
func UpdateUserInfo(userID string, changes entities.UserChanges) sq.UpdateBuilder {
	q := psql.Update(fmt.Sprintf("%s AS u", usersTableName))
	if changes.CI != nil {
		var ci any = *changes.CI
		if n, err := strconv.Atoi(*changes.CI); err == nil {
			ci = n
		}
		q = q.Set("ci", ci)
	}
	if changes.Email != nil {
		q = q.Set("email", *changes.Email)
	}
	if changes.FirstName != nil {
		q = q.Set("first_name", *changes.FirstName)
	}
	if changes.LastName != nil {
		q = q.Set("last_name", *changes.LastName)
	}
	if changes.DateOfBirth != nil {
		q = q.Set("date_of_birth", *changes.DateOfBirth)
	}
	if changes.Gender != nil {
		q = q.Set("gender", nullIfEmpty(*changes.Gender))
	}
	if changes.EducationLevel != nil {
		q = q.Set("education", *changes.EducationLevel)
	}
	if changes.Address != nil {
		q = q.Set("address", nullIfEmpty(*changes.Address))
	}
	return q.
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"u.deleted_at": nil}).
		Where(sq.Eq{"u.id": userID})
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
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
		Set("updated_at", sq.Expr("NOW()")).
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

func UpdateUserRole(userID, fromRoleName, toRoleName string) sq.UpdateBuilder {
	return psql.Update(pivotTableName).
		Set("role_id", sq.Expr(
			fmt.Sprintf("(SELECT id FROM %s WHERE name = ? AND deleted_at IS NULL)", rolesTableName), toRoleName,
		)).
		Where(sq.Eq{"user_id": userID}).
		Where(sq.Expr(
			fmt.Sprintf("role_id = (SELECT id FROM %s WHERE name = ? AND deleted_at IS NULL)", rolesTableName), fromRoleName,
		))
}
