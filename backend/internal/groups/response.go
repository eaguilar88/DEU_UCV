package groups

import (
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/users"
)

type GetGroupResponse struct {
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Owner       *users.GetUserResponse `json:"owner,omitempty"`
	Objective   string                 `json:"objective,omitempty"`
	Location    string                 `json:"location,omitempty"`
	Active      bool                   `json:"active,omitempty"`
	CreatedAt   string                 `json:"created_at,omitempty"`
	UpdatedAt   string                 `json:"updated_at,omitempty"`
	DeletedAt   string                 `json:"deleted_at,omitempty"`
}
type GetGroupsResponse struct {
	Groups    []GetGroupResponse `json:"groups"`
	PageScope entities.PageScope `json:"page"`
}
type CreateGroupResponse struct {
	ID   string `json:"id"`
	Code string `json:"provider_code"`
}
type UpdateGroupResponse struct{}

type DeleteGroupResponse struct{}

func EntitiesGroupsToGetGroupsResponse(groups []entities.ExtensionGroup) []GetGroupResponse {
	var res []GetGroupResponse
	for _, group := range groups {
		res = append(res, EntitiesGroupToGetGroupResponse(group))
	}
	return res
}

func EntitiesGroupToGetGroupResponse(group entities.ExtensionGroup) GetGroupResponse {
	owner := users.UserEntityToGetUserResponse(*group.Owner)
	// Using the new course_request package

	return GetGroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		Owner:       &owner,
		Objective:   group.Objective,
		Location:    group.Location,
		Active:      group.Active,
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
	}
}
