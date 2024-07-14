package users

import "github.com/eaguilar88/deu/pkg/entities"

type GetUserResponse struct {
	User entities.User `json:"user"`
}

type GetUsersResponse struct {
	Users []entities.User    `json:"users,omitempty"`
	Pages entities.PageScope `json:"pages,omitempty"`
}

type CreateUsersResponse struct {
	ID string `json:"id,omitempty"`
}
