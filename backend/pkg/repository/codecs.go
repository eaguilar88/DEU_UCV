package repository

import (
	"database/sql"
	"time"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/eaguilar88/deu/pkg/repository/models"
)

func newUserFromEntity(user entities.User, isUpdate bool) models.User {
	model := models.User{
		ID:          user.ID,
		Username:    user.Username,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		DateOfBirth: user.DateOfBirth,
		Gender: sql.NullString{
			String: user.Gender,
			Valid:  true,
		},
		EducationLevel: user.EducationLevel,
		Address: sql.NullString{
			String: user.Address,
			Valid:  true,
		},
		Password:  user.Password,
		CreatedAt: time.Now().String(),
		UpdatedAt: time.Now().String(),
	}
	if isUpdate {
		model.CreatedAt = user.CreatedAt
	}
	return model
}

func newUserFromModel(user models.User) entities.User {
	entity := entities.User{
		ID:             user.ID,
		Username:       user.Username,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		DateOfBirth:    user.DateOfBirth,
		EducationLevel: user.EducationLevel,
		Password:       user.Password,
		CreatedAt:      time.Now().String(),
	}

	if user.Gender.Valid {
		entity.Gender = user.Gender.String
	}
	if user.Address.Valid {
		entity.Address = user.Address.String
	}
	entity.SetAge()
	return entity
}
