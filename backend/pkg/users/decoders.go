package users

import (
	"github.com/eaguilar88/deu/pkg/entities"
)

func createUserRequestToEntitiesUser(req CreateUserRequest) entities.User {
	return entities.User{
		CI:             req.Document,
		Username:       req.Username,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		DateOfBirth:    req.DateOfBirth,
		Gender:         req.Gender,
		EducationLevel: req.EducationLevel,
		Address:        req.Address,
		Password:       req.Password,
	}
}

func updateUserRequestToEntitiesUser(req UpdateUserRequest, ID int) entities.User {

	return entities.User{
		ID:             ID,
		CI:             req.Document,
		Username:       req.Username,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		DateOfBirth:    req.DateOfBirth,
		Gender:         req.Gender,
		EducationLevel: req.EducationLevel,
		Address:        req.Address,
		Password:       req.Password,
	}
}
