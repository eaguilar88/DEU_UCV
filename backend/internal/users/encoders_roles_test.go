package users

import (
	"encoding/json"
	"testing"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserToResponse_UIRoles(t *testing.T) {
	t.Run("full profile carries the UI role names", func(t *testing.T) {
		resp := userToResponse(entities.User{ID: "15", Roles: []string{"course_admin"}})

		assert.Equal(t, "proveedor", resp.Rol)
		assert.Equal(t, []string{"proveedor"}, resp.Roles)
	})

	t.Run("public profile has no roles", func(t *testing.T) {
		data, err := json.Marshal(userToResponse(entities.User{ID: "15", FirstName: "Ana"}))
		require.NoError(t, err)

		assert.NotContains(t, string(data), `"rol"`)
		assert.NotContains(t, string(data), `"roles"`)
	})
}
