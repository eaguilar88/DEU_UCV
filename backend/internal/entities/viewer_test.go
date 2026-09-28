package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestViewer_CanManageGroup(t *testing.T) {
	group := ExtensionGroup{
		ID:      "1",
		Owner:   &User{ID: "10"},
		Faculty: []Faculty{FacultyIngenieria},
	}

	tests := []struct {
		name   string
		viewer Viewer
		want   bool
	}{
		{name: "anonymous", viewer: Viewer{}, want: false},
		{name: "owner", viewer: Viewer{UserID: "10", Roles: []string{"group_admin"}}, want: true},
		{name: "root", viewer: Viewer{UserID: "1", Roles: []string{"root"}}, want: true},
		{name: "deu_admin", viewer: Viewer{UserID: "2", Roles: []string{"deu_admin"}}, want: true},
		{
			name:   "faculty_admin of the group's faculty",
			viewer: Viewer{UserID: "3", Roles: []string{"faculty_admin"}, Faculty: FacultyIngenieria},
			want:   true,
		},
		{
			name:   "faculty_admin of another faculty",
			viewer: Viewer{UserID: "4", Roles: []string{"faculty_admin"}, Faculty: FacultyCiencias},
			want:   false,
		},
		{
			name:   "faculty_admin with DEU-wide claim",
			viewer: Viewer{UserID: "5", Roles: []string{"faculty_admin"}, Faculty: FacultyDEU},
			want:   true,
		},
		{name: "group_admin of another group", viewer: Viewer{UserID: "11", Roles: []string{"group_admin"}}, want: false},
		{name: "plain user", viewer: Viewer{UserID: "12", Roles: []string{"visitante"}}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.viewer.CanManageGroup(group))
		})
	}
}
