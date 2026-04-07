package groups

import (
	"github.com/eaguilar88/deu/internal/entities"
)

type OwnerInfo struct {
	ID        string `json:"id,omitempty"`
	CI        string `json:"cedula,omitempty"`
	Email     string `json:"email,omitempty"`
	FirstName string `json:"nombres,omitempty"`
	LastName  string `json:"apellidos,omitempty"`
}

type GetGroupResponse struct {
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"nombre,omitempty"`
	Description string                 `json:"descripcion,omitempty"`
	Owner       *OwnerInfo             `json:"propietario,omitempty"`
	Objective   string                 `json:"objetivo,omitempty"`
	Location    string                 `json:"ubicacion,omitempty"`
	Active      bool                   `json:"activo,omitempty"`
	Members     []entities.GroupMember `json:"miembros,omitempty"`
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

// groupsToResponse converts a slice of ExtensionGroup entities to responses.
func groupsToResponse(groups []entities.ExtensionGroup) []GetGroupResponse {
	var res []GetGroupResponse
	for _, group := range groups {
		res = append(res, groupToResponse(group))
	}
	return res
}

// groupToResponse converts an ExtensionGroup entity to GetGroupResponse.
func groupToResponse(group entities.ExtensionGroup) GetGroupResponse {
	var owner *OwnerInfo
	if group.Owner != nil {
		owner = &OwnerInfo{
			ID:        group.Owner.ID,
			CI:        group.Owner.CI,
			Email:     group.Owner.Email,
			FirstName: group.Owner.FirstName,
			LastName:  group.Owner.LastName,
		}
	}

	return GetGroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		Owner:       owner,
		Objective:   group.Objective,
		Location:    group.Location,
		Active:      group.Active,
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
	}
}
