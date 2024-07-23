package users

import "github.com/eaguilar88/deu/pkg/entities"

func entitiesUserToGetUserResponse(user entities.User) GetUserResponse {
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
		Address:        user.Address,
		Password:       user.Password,
		CreatedAt:      user.CreatedAt,
	}
}
