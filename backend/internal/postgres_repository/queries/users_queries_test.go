package queries

import (
	"testing"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateUserInfo(t *testing.T) {
	str := func(s string) *string { return &s }

	tests := []struct {
		name     string
		changes  entities.UserChanges
		wantSQL  string
		wantArgs []any
	}{
		{
			name:     "only the fields present are set",
			changes:  entities.UserChanges{FirstName: str("Ana"), Email: str("ana@example.com")},
			wantSQL:  "UPDATE deu.users AS u SET email = $1, first_name = $2, updated_at = NOW() WHERE u.deleted_at IS NULL AND u.id = $3",
			wantArgs: []any{"ana@example.com", "Ana", "7"},
		},
		{
			name: "every field, cedula as a number and empty gender/address as NULL",
			changes: entities.UserChanges{
				CI: str("12345678"), Email: str("ana@example.com"), FirstName: str("Ana"), LastName: str("Pérez"),
				DateOfBirth: str("1990-03-15"), Gender: str(""), EducationLevel: str("posgrado"), Address: str(""),
			},
			wantSQL: "UPDATE deu.users AS u SET ci = $1, email = $2, first_name = $3, last_name = $4, " +
				"date_of_birth = $5, gender = $6, education = $7, address = $8, updated_at = NOW() " +
				"WHERE u.deleted_at IS NULL AND u.id = $9",
			wantArgs: []any{12345678, "ana@example.com", "Ana", "Pérez", "1990-03-15", nil, "posgrado", nil, "7"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args, err := UpdateUserInfo("7", tt.changes).ToSql()
			require.NoError(t, err)
			assert.Equal(t, tt.wantSQL, sql)
			assert.Equal(t, tt.wantArgs, args)
		})
	}
}
