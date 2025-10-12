package queries

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

const (
	schema = "deu"
)

// Table names
var (
	psql                    = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	usersTableName          = fmt.Sprintf("%s.users", schema)
	rolesTableName          = fmt.Sprintf("%s.roles", schema)
	pivotTableName          = fmt.Sprintf("%s.user_roles", schema)
	entitiesTableName       = fmt.Sprintf("%s.entities", schema)
	coursesTableName        = fmt.Sprintf("%s.courses", schema)
	providersTableName      = fmt.Sprintf("%s.providers", schema)
	courseRequestsTableName = fmt.Sprintf("%s.course_auth_requests", schema)
	groupsTableName         = fmt.Sprintf("%s.extension_groups", schema)
	groupRequestsTableName  = fmt.Sprintf("%s.group_auth_requests", schema)
	periodsTableName        = fmt.Sprintf("%s.course_cycles", schema)
	participantsTableName   = fmt.Sprintf("%s.course_participants", schema)
	filesTableName          = fmt.Sprintf("%s.files", schema)
)
