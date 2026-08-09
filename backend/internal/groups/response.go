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

type GroupMemberResponse struct {
	ID           string `json:"id,omitempty"`
	Name         string `json:"nombre,omitempty"`
	CI           int    `json:"cedula,omitempty"`
	Phone        string `json:"telefono,omitempty"`
	Email        string `json:"correo,omitempty"`
	Coordination string `json:"coordinacion,omitempty"`
	Year         string `json:"año,omitempty"`
	Faculty      string `json:"facultad,omitempty"`
	School       string `json:"escuela,omitempty"`
	Document     string `json:"documento,omitempty"`
	IsActive     bool   `json:"status"`
}

type GetGroupResponse struct {
	ID          string                `json:"id,omitempty"`
	Name        string                `json:"nombre,omitempty"`
	Description string                `json:"descripcion,omitempty"`
	Faculty     string                `json:"facultad,omitempty"`
	Foundation  string                `json:"fundacion,omitempty"`
	Type        string                `json:"tipo,omitempty"`
	LogoURL     string                `json:"imagen_url,omitempty"`
	Email       string                `json:"email,omitempty"`
	Phone       string                `json:"telefono,omitempty"`
	Owner       *OwnerInfo            `json:"propietario,omitempty"`
	Objective   string                `json:"objetivo,omitempty"`
	Location    string                `json:"ubicacion,omitempty"`
	Active      bool                  `json:"activo,omitempty"`
	Members     []GroupMemberResponse `json:"miembros,omitempty"`
	CreatedAt   string                `json:"creado_en,omitempty"`
	UpdatedAt   string                `json:"actualizado_en,omitempty"`
	DeletedAt   string                `json:"eliminado_en,omitempty"`
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

func memberToResponse(m entities.GroupMember) GroupMemberResponse {
	return GroupMemberResponse{
		ID:           m.ID,
		Name:         m.Name,
		CI:           m.CI,
		Phone:        m.Phone,
		Email:        m.Email,
		Coordination: m.Coordination,
		Year:         m.Year,
		Faculty:      string(m.Faculty),
		School:       m.School,
		Document:     m.Document,
		IsActive:     m.IsActive,
	}
}

func membersToResponse(members []entities.GroupMember) []GroupMemberResponse {
	res := make([]GroupMemberResponse, len(members))
	for i, m := range members {
		res[i] = memberToResponse(m)
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

	var logoURL string
	if group.Logo != nil {
		logoURL = group.Logo.URL
	}

	return GetGroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		Faculty:     string(group.Faculty),
		Foundation:  group.Foundation,
		Type:        string(group.Type),
		LogoURL:     logoURL,
		Email:       group.Email,
		Phone:       group.Phone,
		Owner:       owner,
		Objective:   group.Objective,
		Location:    group.Location,
		Active:      group.Active,
		Members:     membersToResponse(group.Members),
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
	}
}
