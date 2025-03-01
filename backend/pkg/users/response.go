package users

import "github.com/eaguilar88/deu/pkg/entities"

type GetUserResponse struct {
	ID             int    `json:"id"`
	CI             string `json:"ci"`
	Username       string `json:"username"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	DateOfBirth    string `json:"date_of_birth"`
	Age            int    `json:"age"`
	Gender         string `json:"gender,omitempty"`
	EducationLevel string `json:"education_level,omitempty"`
	Code           string `json:"code,omitempty"`
	Address        string `json:"address,omitempty"`
	CreatedAt      string `json:"created_at"`
}

type GetUsersResponse struct {
	Users []GetUserResponse  `json:"users,omitempty"`
	Pages entities.PageScope `json:"pages,omitempty"`
}

type CreateUsersResponse struct {
	ID string `json:"id,omitempty"`
}

func userEntitiesToUserDTO(users []entities.User) []GetUserResponse {
	var out = make([]GetUserResponse, 0, len(users))
	for _, user := range users {
		out = append(out, GetUserResponse{
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
		})
	}
	return out
}
