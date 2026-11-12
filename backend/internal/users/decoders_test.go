package users

import (
	"testing"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/stretchr/testify/assert"
)

func TestToUserEntity(t *testing.T) {
	tests := []struct {
		name        string
		dateOfBirth string
		wantDOB     string
		wantErr     bool
	}{
		{
			name:        "valid date DD-MM-YYYY",
			dateOfBirth: "15-03-1990",
			wantDOB:     "1990-03-15",
		},
		// {
		// 	name:        "valid date single digit day and month",
		// 	dateOfBirth: "01-01-2000",
		// 	wantDOB:     "2000-01-01",
		// },
		// {
		// 	name:        "valid date two-digit day and month",
		// 	dateOfBirth: "25-12-1985",
		// 	wantDOB:     "1985-12-25",
		// },
		// {
		// 	name:        "wrong order DD-MM-YYYY with impossible month",
		// 	dateOfBirth: "03-15-1990",
		// 	wantErr:     true,
		// },
		// {
		// 	name:        "empty string",
		// 	dateOfBirth: "",
		// 	wantErr:     true,
		// },
		// {
		// 	name:        "ISO format YYYY-MM-DD is rejected",
		// 	dateOfBirth: "1990-03-15",
		// 	wantErr:     true,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateUserRequest{
				Document:    "12345678",
				Email:       "test@test.com",
				FirstName:   "John",
				LastName:    "Doe",
				DateOfBirth: tt.dateOfBirth,
				Password:    "password123",
			}

			user, err := toUserEntity(req)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantDOB, user.DateOfBirth)
		})
	}
}

func TestToUserChanges(t *testing.T) {
	str := func(s string) *string { return &s }

	tests := []struct {
		name    string
		req     UpdateUserRequest
		want    entities.UserChanges
		wantErr string
	}{
		{
			name: "only the fields sent are set",
			req:  UpdateUserRequest{ID: "1", FirstName: str("María"), Address: str("Caracas")},
			want: entities.UserChanges{FirstName: str("María"), Address: str("Caracas")},
		},
		{
			name: "every field, date converted to storage format",
			req: UpdateUserRequest{
				ID: "1", Document: str(" 12345678 "), Email: str(" ana@example.com "),
				FirstName: str("Ana"), LastName: str("Pérez"), DateOfBirth: str("15-03-1990"),
				Gender: str("femenino"), EducationLevel: str("posgrado"), Address: str("Caracas"),
			},
			want: entities.UserChanges{
				CI: str("12345678"), Email: str("ana@example.com"),
				FirstName: str("Ana"), LastName: str("Pérez"), DateOfBirth: str("1990-03-15"),
				Gender: str("femenino"), EducationLevel: str("posgrado"), Address: str("Caracas"),
			},
		},
		{
			name: "genero and direccion can be cleared",
			req:  UpdateUserRequest{ID: "1", Gender: str(""), Address: str("")},
			want: entities.UserChanges{Gender: str(""), Address: str("")},
		},
		{
			name: "nothing sent",
			req:  UpdateUserRequest{ID: "1"},
			want: entities.UserChanges{},
		},
		{name: "empty nombres", req: UpdateUserRequest{FirstName: str(" ")}, wantErr: "nombres no puede estar vacío"},
		{name: "empty apellidos", req: UpdateUserRequest{LastName: str("")}, wantErr: "apellidos no puede estar vacío"},
		{name: "empty email", req: UpdateUserRequest{Email: str("")}, wantErr: "email no puede estar vacío"},
		{name: "empty cedula", req: UpdateUserRequest{Document: str("")}, wantErr: "cedula no puede estar vacío"},
		{name: "invalid email", req: UpdateUserRequest{Email: str("not-an-email")}, wantErr: "email no es válido"},
		{name: "non numeric cedula", req: UpdateUserRequest{Document: str("V-123")}, wantErr: "cedula debe ser un número positivo"},
		{name: "negative cedula", req: UpdateUserRequest{Document: str("-5")}, wantErr: "cedula debe ser un número positivo"},
		{name: "ISO date is rejected", req: UpdateUserRequest{DateOfBirth: str("1990-03-15")}, wantErr: "fecha"},
		{name: "empty date is rejected", req: UpdateUserRequest{DateOfBirth: str("")}, wantErr: "fecha"},
		{name: "unknown nivel_educativo", req: UpdateUserRequest{EducationLevel: str("doctorado")}, wantErr: "nivel_educativo debe ser uno de"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toUserChanges(tt.req)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUserToResponse(t *testing.T) {
	tests := []struct {
		name        string
		dateOfBirth string
		wantDOB     string
	}{
		{
			name:        "formats stored YYYY-MM-DD to D-M-YYYY",
			dateOfBirth: "1990-03-15",
			wantDOB:     "15-03-1990",
		},
		{
			name:        "formats leading zeros correctly",
			dateOfBirth: "2000-01-01",
			wantDOB:     "01-01-2000",
		},
		{
			name:        "returns raw value if not parseable",
			dateOfBirth: "not-a-date",
			wantDOB:     "not-a-date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := newTestUser()
			user.DateOfBirth = tt.dateOfBirth

			resp := userToResponse(*user)
			assert.Equal(t, tt.wantDOB, resp.DateOfBirth)
		})
	}
}
