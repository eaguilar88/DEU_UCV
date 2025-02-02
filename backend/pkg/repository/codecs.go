package repository

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/eaguilar88/deu/pkg/repository/models"
)

func newUserFromEntity(user entities.User, isUpdate bool) models.User {
	ci, err := strconv.Atoi(user.CI)
	if err != nil {
		ci = 0
	}
	model := models.User{
		ID:        user.ID,
		CI:        ci,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		DateOfBirth: sql.NullString{
			String: user.DateOfBirth,
			Valid:  true,
		},
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
		CI:             fmt.Sprintf("%d", user.CI),
		Username:       user.Username,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		EducationLevel: user.EducationLevel,
		Password:       user.Password,
		CreatedAt:      user.CreatedAt,
	}

	if user.Gender.Valid {
		entity.Gender = user.Gender.String
	}
	if user.Address.Valid {
		entity.Address = user.Address.String
	}

	if user.DateOfBirth.Valid {
		entity.DateOfBirth = user.DateOfBirth.String
	}

	if user.ProviderCode.Valid {
		entity.ProviderCode = user.ProviderCode.String
	}

	entity.SetAge()
	return entity
}

func newEndorsmentFromModel(endorsement models.Endorsement) entities.Endorsements {
	result := entities.Endorsements{
		ID: endorsement.ID,
		User: entities.User{
			ID: endorsement.ID,
		},
		Status: entities.EndorsementStatus(endorsement.Status),

		CreatedAt:   endorsement.CreatedAt,
		UpdatedAtAt: endorsement.UpdatedAt,
	}

	if endorsement.Name.Valid {
		result.Name = endorsement.Name.String
	}

	if endorsement.Description.Valid {
		result.Description = endorsement.Description.String
	}

	if endorsement.Comments.Valid {
		result.Comments = endorsement.Comments.String
	}

	return result
}
