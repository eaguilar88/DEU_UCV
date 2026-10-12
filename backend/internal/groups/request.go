package groups

import (
	"fmt"

	"github.com/eaguilar88/deu/internal/entities"
)

type GetGroupRequest struct {
	ID string `param:"id" validate:"required"`
}
type GetGroupsRequest struct {
	Faculty string `form:"facultad"`
	Active  *bool  `form:"activo"`
	Search  string `form:"q"`
	Page    int    `query:"page" validate:"required"`
	PerPage int    `query:"per_page" validate:"required"`
}
type CreateGroupRequest struct {
	Name                string           `validate:"required" form:"nombre"`
	Description         string           `form:"descripcion"`
	Type                []string         `validate:"required,min=1" form:"tipo"`
	Foundation          string           `form:"fundacion"`
	IsMultidisciplinary bool             `form:"es_multidisciplinario"`
	Faculty             []string         `validate:"required,min=1" form:"facultad"`
	Objective           string           `validate:"required" form:"objetivo"`
	Location            string           `form:"ubicacion"`
	Members             []GroupMemberDTO `form:"miembros" validate:"dive"`
	Email               string           `form:"correo" validate:"required"`
	Phone               string           `form:"telefono"`
}

type DeleteGroupRequest struct {
	ID      string `param:"id" validate:"required"`
	OwnerID string `validate:"required"`
}

type GroupMemberDTO struct {
	ID           string `json:"id,omitempty"`
	Name         string `json:"nombre" validate:"required"`
	CI           int    `json:"cedula" validate:"required"`
	Phone        string `json:"telefono" validate:"required"`
	Email        string `json:"correo" validate:"required,email"`
	Coordination string `json:"coordinacion" validate:"required"`
	Year         string `json:"año" validate:"required"`
	Faculty      string `json:"facultad" validate:"required"`
	School       string `json:"escuela" validate:"required"`
	IsLeader     bool   `json:"es_lider"`
	IsActive     bool   `json:"status"`
}

// groupMemberEntityFromRequest converts a GroupMemberDTO to a GroupMember entity.
func groupMemberEntityFromRequest(dto GroupMemberDTO) entities.GroupMember {
	return entities.GroupMember{
		ID:           dto.ID,
		Name:         dto.Name,
		CI:           dto.CI,
		Phone:        dto.Phone,
		Email:        dto.Email,
		Coordination: dto.Coordination,
		Year:         dto.Year,
		Faculty:      entities.Faculty(dto.Faculty),
		School:       dto.School,
		IsLeader:     dto.IsLeader,
		IsActive:     dto.IsActive,
	}
}

// groupMembersEntityFromRequest converts a slice of GroupMemberDTO to GroupMember entities.
func groupMembersEntityFromRequest(dtos []GroupMemberDTO) []entities.GroupMember {
	members := make([]entities.GroupMember, len(dtos))
	for i, dto := range dtos {
		members[i] = groupMemberEntityFromRequest(dto)
	}
	return members
}

// ValidateFaculty checks if the provided faculty string is valid and returns the Faculty type.
func ValidateFaculty(facultyStr string) (entities.Faculty, error) {
	faculty := entities.Faculty(facultyStr)
	if !faculty.IsValid() {
		return "", fmt.Errorf("invalid faculty: %s", facultyStr)
	}
	return faculty, nil
}

// createGroupEntityFromRequest converts CreateGroupRequest to an ExtensionGroup entity.
func createGroupEntityFromRequest(req CreateGroupRequest, ownerID string, faculties []string) entities.ExtensionGroup {
	validFaculties := make([]entities.Faculty, 0, len(faculties))
	for _, f := range faculties {
		fac := entities.Faculty(f)
		if fac.IsValid() {
			validFaculties = append(validFaculties, fac)
		}
	}

	types := make([]entities.GroupType, 0, len(req.Type))
	for _, t := range req.Type {
		gt := entities.GroupType(t)
		if gt.IsValid() {
			types = append(types, gt)
		}
	}

	return entities.ExtensionGroup{
		Name:                req.Name,
		Description:         req.Description,
		Foundation:          req.Foundation,
		IsMultidisciplinary: req.IsMultidisciplinary,
		Owner: &entities.User{
			ID: ownerID,
		},
		Type:      types,
		Faculty:   validFaculties,
		Objective: req.Objective,
		Location:  req.Location,
		Members:   groupMembersEntityFromRequest(req.Members),
		Email:     req.Email,
		Phone:     req.Phone,
		Active:    false,
	}
}
