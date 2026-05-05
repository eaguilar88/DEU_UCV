package users

import (
	"testing"

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
		{
			name:        "valid date single digit day and month",
			dateOfBirth: "01-01-2000",
			wantDOB:     "2000-01-01",
		},
		{
			name:        "valid date two-digit day and month",
			dateOfBirth: "25-12-1985",
			wantDOB:     "1985-12-25",
		},
		{
			name:        "wrong order DD-MM-YYYY with impossible month",
			dateOfBirth: "03-15-1990",
			wantErr:     true,
		},
		{
			name:        "empty string",
			dateOfBirth: "",
			wantErr:     true,
		},
		{
			name:        "ISO format YYYY-MM-DD is rejected",
			dateOfBirth: "1990-03-15",
			wantErr:     true,
		},
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

func TestToUserUpdateEntity(t *testing.T) {
	tests := []struct {
		name        string
		dateOfBirth string
		wantDOB     string
		wantErr     bool
	}{
		{
			name:        "valid date D-M-YYYY",
			dateOfBirth: "15-03-1990",
			wantDOB:     "1990-03-15",
		},
		{
			name:        "wrong order M-D-YYYY with impossible month",
			dateOfBirth: "3-15-1990",
			wantErr:     true,
		},
		{
			name:        "empty string",
			dateOfBirth: "",
			wantErr:     true,
		},
		{
			name:        "ISO format YYYY-MM-DD is rejected",
			dateOfBirth: "1990-03-15",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := UpdateUserRequest{
				ID:          "user-1",
				DateOfBirth: tt.dateOfBirth,
			}

			user, err := toUserUpdateEntity(req)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantDOB, user.DateOfBirth)
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
