package users

import (
	"time"

	"github.com/eaguilar88/deu/internal/entities"
)

// userToResponse converts a User entity to GetUserResponse.
func userToResponse(user entities.User) GetUserResponse {
	dob := user.DateOfBirth
	if t, err := time.Parse(storageDateFormat, user.DateOfBirth); err == nil {
		dob = t.Format(clientDateFormat)
	}

	return GetUserResponse{
		ID:                user.ID,
		CI:                user.CI,
		Email:             user.Email,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		DateOfBirth:       dob,
		Age:               user.Age,
		Gender:            user.Gender,
		EducationLevel:    user.EducationLevel,
		Code:              user.ProviderCode,
		Address:           user.Address,
		CreatedAt:         user.CreatedAt,
		ProfilePictureURL: user.ProfilePictureURL,
	}
}

// usersToResponse converts a slice of User entities to responses.
func usersToResponse(users []entities.User) []GetUserResponse {
	out := make([]GetUserResponse, 0, len(users))
	for _, user := range users {
		out = append(out, userToResponse(user))
	}
	return out
}
