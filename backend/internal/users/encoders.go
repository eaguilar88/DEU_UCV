package users

import "github.com/eaguilar88/deu/internal/entities"

func UserEntityToGetUserResponse(user entities.User) GetUserResponse {
	return GetUserResponse{
		ID:             user.ID,
		CI:             user.CI,
		Username:       user.Username,
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

func UserEntitiesToGetUserResponse(users []entities.User) []GetUserResponse {
	var out = make([]GetUserResponse, 0, len(users))
	for _, user := range users {
		out = append(out, UserEntityToGetUserResponse(user))
	}
	return out
}
