package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUIRoles(t *testing.T) {
	tests := []struct {
		name      string
		roles     []string
		wantRoles []string
		wantRole  string
	}{
		{name: "approved provider", roles: []string{"course_admin"}, wantRoles: []string{"proveedor"}, wantRole: "proveedor"},
		{name: "faculty coordinator", roles: []string{"faculty_admin"}, wantRoles: []string{"coordinador"}, wantRole: "coordinador"},
		{name: "root and deu_admin are both admin", roles: []string{"root", "deu_admin"}, wantRoles: []string{"admin"}, wantRole: "admin"},
		{name: "visitor", roles: []string{"visitante"}, wantRoles: []string{"visitante"}, wantRole: "visitante"},
		{name: "most privileged wins", roles: []string{"visitante", "course_admin", "faculty_admin"}, wantRoles: []string{"visitante", "proveedor", "coordinador"}, wantRole: "coordinador"},
		{name: "group roles are kept as they are", roles: []string{"group_admin"}, wantRoles: []string{"group_admin"}, wantRole: "group_admin"},
		{name: "no roles", roles: nil, wantRoles: []string{}, wantRole: "visitante"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantRoles, UIRoles(tt.roles))
			assert.Equal(t, tt.wantRole, UIRole(tt.roles))
		})
	}
}
