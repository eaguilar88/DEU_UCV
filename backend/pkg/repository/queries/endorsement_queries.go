package queries

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/pkg/entities"
)

var (
	endorsementTableName         = fmt.Sprintf("%s.endorsement_requests", schema)
	endorsementQuerySelectCommon = []string{
		"u.id", "u.ci_type", "u.ci", "u.username", "u.first_name", "u.last_name", "u.date_of_birth", "u.gender", "u.education", "u.address", "u.created_at",
	}
)

func GetEndorsementByID(endorsementID int) sq.SelectBuilder {
	return psql.Select(endorsementQuerySelectCommon...).
		From(fmt.Sprintf("%s AS u", endorsementTableName)).
		Where(sq.Eq{"u.id": endorsementID})
}

func GetEndorsements(page entities.PageScope) sq.SelectBuilder {
	return psql.Select(endorsementQuerySelectCommon...).
		From(fmt.Sprintf("%s AS u", endorsementTableName)).
		Limit(uint64(page.PerPage)).
		Offset(uint64(page.Offset()))
}

func InsertEndorsement(user entities.User) sq.InsertBuilder {
	ciType, ciNumber := splitUserCI(user)

	return psql.Insert(endorsementTableName).
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

func UpdateEndorsementInfo(user entities.User, userID int) sq.UpdateBuilder {
	return psql.Update(endorsementTableName).
		Set("first_name", user.FirstName).
		Set("last_name", user.LastName).
		Set("date_of_birth", user.DateOfBirth).
		Set("gender", user.Gender).
		Set("address", user.Address).
		Set("education", user.EducationLevel).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": userID})
}

func DeleteEndorsement(endorsementID int) sq.DeleteBuilder {
	return psql.Delete(endorsementTableName).
		Where(sq.Eq{"id": endorsementID})
}
