package groups

import (
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/users"
)

type GetGroupResponse struct {
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"nombre,omitempty"`
	Description string                 `json:"descripcion,omitempty"`
	Owner       *users.GetUserResponse `json:"propietario,omitempty"`
	Objective   string                 `json:"objetivo,omitempty"`
	Location    string                 `json:"ubicacion,omitempty"`
	Active      bool                   `json:"activo,omitempty"`
	CreatedAt   string                 `json:"creado_en,omitempty"`
	UpdatedAt   string                 `json:"actualizado_en,omitempty"`
	DeletedAt   string                 `json:"eliminado_en,omitempty"`
}
type GetGroupsResponse struct {
	Groups    []GetGroupResponse `json:"grupos"`
	PageScope entities.PageScope `json:"pagina"`
}
type CreateGroupResponse struct {
	ID   string `json:"id"`
	Code string `json:"codigo_proveedor"`
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
