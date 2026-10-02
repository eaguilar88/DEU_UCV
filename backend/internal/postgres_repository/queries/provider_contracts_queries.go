package queries

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

var providerContractQuerySelectCommon = []string{
	"pc.id",
	"pc.provider_id",
	"pc.type",
	"pc.created_at",
}

func InsertProviderContract(providerID, contractType string) sq.InsertBuilder {
	return psql.Insert(providerContractsTableName).
		Columns("provider_id", "type").
		Values(providerID, contractType).
		Suffix("RETURNING id")
}

func InsertProviderContractCourses(contractID string, courseIDs []string) sq.InsertBuilder {
	q := psql.Insert(providerContractCoursesTableName).
		Columns("contract_id", "course_id")
	for _, courseID := range courseIDs {
		q = q.Values(contractID, courseID)
	}
	return q
}

func CountInitialContracts(providerID string) sq.SelectBuilder {
	return psql.Select("COUNT(*)").
		From(fmt.Sprintf("%s AS pc", providerContractsTableName)).
		Where(sq.Eq{"pc.provider_id": providerID}).
		Where(sq.Eq{"pc.type": "inicial"}).
		Where(sq.Eq{"pc.deleted_at": nil})
}

func GetUncoveredCourseIDs(providerID string) sq.SelectBuilder {
	return psql.Select("c.id").
		From(fmt.Sprintf("%s AS c", coursesTableName)).
		Where(sq.Eq{"c.provider_id": providerID}).
		Where(sq.Eq{"c.is_active": true}).
		Where(sq.Eq{"c.has_documentation": false}).
		Where(sq.Eq{"c.deleted_at": nil})
}

func GetProviderContracts(providerID string) sq.SelectBuilder {
	return psql.Select(providerContractQuerySelectCommon...).
		From(fmt.Sprintf("%s AS pc", providerContractsTableName)).
		Where(sq.Eq{"pc.provider_id": providerID}).
		Where(sq.Eq{"pc.deleted_at": nil}).
		OrderBy("pc.created_at ASC")
}

func GetContractCourseIDs(contractID string) sq.SelectBuilder {
	return psql.Select("course_id").
		From(providerContractCoursesTableName).
		Where(sq.Eq{"contract_id": contractID})
}

func SetCoursesDocumented(courseIDs []string) sq.UpdateBuilder {
	return psql.Update(coursesTableName).
		Set("has_documentation", true).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": courseIDs})
}
