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
	psql                         = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	usersTableName               = fmt.Sprintf("%s.users", schema)
	rolesTableName               = fmt.Sprintf("%s.roles", schema)
	pivotTableName               = fmt.Sprintf("%s.user_roles", schema)
	entitiesTableName            = fmt.Sprintf("%s.entities", schema)
	coursesTableName             = fmt.Sprintf("%s.courses", schema)
	providersTableName           = fmt.Sprintf("%s.providers", schema)
	courseRequestsTableName      = fmt.Sprintf("%s.course_auth-requests", schema)
	groupsTableName              = fmt.Sprintf("%s.extension_groups", schema)
	periodsTableName             = fmt.Sprintf("%s.course_cycles", schema)
	participantsTableName        = fmt.Sprintf("%s.course_participants", schema)
	filesTableName               = fmt.Sprintf("%s.files", schema)
	courseParticipantsTableName  = fmt.Sprintf("%s.course_participants", schema)
	courseInstructorsTableName   = fmt.Sprintf("%s.course_instructors", schema)
	courseModulesTableName       = fmt.Sprintf("%s.course_modules", schema)
	moduleContentsTableName      = fmt.Sprintf("%s.module_contents", schema)
	moduleActivitiesTableName    = fmt.Sprintf("%s.module_activities", schema)
	activitySubmissionsTableName = fmt.Sprintf("%s.activity_submissions", schema)
	gradesTableName              = fmt.Sprintf("%s.grades", schema)
	certificatesTableName        = fmt.Sprintf("%s.certificates", schema)
)
