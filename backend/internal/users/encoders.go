package users

import "github.com/eaguilar88/deu/internal/entities"

// userToResponse converts a User entity to GetUserResponse.
func userToResponse(user entities.User) GetUserResponse {
	return GetUserResponse{
		ID:             user.ID,
		CI:             user.CI,
		Email:          user.Email,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		DateOfBirth:    user.DateOfBirth,
		Age:            user.Age,
		Gender:         user.Gender,
		EducationLevel: user.EducationLevel,
		Code:           user.ProviderCode,
		Address:        user.Address,
		CreatedAt:      user.CreatedAt,
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
