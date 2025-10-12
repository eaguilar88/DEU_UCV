package groups

import (
	"fmt"

	"github.com/eaguilar88/deu/internal/entities"
)

type GetGroupRequest struct {
	ID string `param:"id" validate:"required"`
}
type GetGroupsRequest struct {
	Page    int `query:"page"     validate:"required"`
	PerPage int `query:"per_page" validate:"required"`
}
type CreateGroupRequest struct {
	Name        string           `validate:"required" form:"name"`
	Description string           `                    form:"description"`
	LeaderName  string           `validate:"required" form:"leader_name"`
	Type        string           `validate:"required" form:"type"`
	Faculty     string           `validate:"required" form:"faculty"`
	Objective   string           `validate:"required" form:"objective"`
	Location    string           `validate:"required" form:"location"`
	Members     []GroupMemberDTO `                    form:"members"`
}

type UpdateGroupRequest struct {
	ID            string           `param:"id" validate:"required"`
	OwnerID       string           `           validate:"required"`
	Name          string           `           validate:"required" form:"name"`
	Description   string           `                               form:"description"`
	EndorsementID string           `           validate:"required" form:"endorsement_id"`
	Objective     string           `           validate:"required" form:"objective"`
	Location      string           `           validate:"required" form:"location"`
	Members       []GroupMemberDTO `                               form:"members"`
}
type DeleteGroupRequest struct {
	ID      string `param:"id" validate:"required"`
	OwnerID string `           validate:"required"`
}

type GroupMemberDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"  validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

func groupMemberEntityFromRequest(dto GroupMemberDTO) entities.GroupMember {
	return entities.GroupMember{
		ID:    dto.ID,
		Name:  dto.Name,
		Email: dto.Email,
	}
}

func groupMembersEntityFromRequest(dtos []GroupMemberDTO) []entities.GroupMember {
	members := make([]entities.GroupMember, len(dtos))
	for i, dto := range dtos {
		members[i] = groupMemberEntityFromRequest(dto)
	}
	return members
}

// ValidateFaculty checks if the provided faculty string is valid and returns the Faculty type
func ValidateFaculty(facultyStr string) (entities.Faculty, error) {
	faculty := entities.Faculty(facultyStr)
	if !faculty.IsValid() {
		return "", fmt.Errorf("invalid faculty: %s", facultyStr)
	}
	return faculty, nil
}

func updateGroupEntityFromRequest(req UpdateGroupRequest) entities.ExtensionGroup {
	return entities.ExtensionGroup{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		Owner: &entities.User{
			ID: req.OwnerID,
		},
		// CourseRequest: entities.CourseRequest{
		// 	ID: req.EndorsementID,
		// },
		Objective: req.Objective,
		Location:  req.Location,
		Members:   groupMembersEntityFromRequest(req.Members),
	}
}

func createGroupEntityFromRequest(req CreateGroupRequest, ownerID, faculty string) entities.ExtensionGroup {
	eg := entities.ExtensionGroup{
		Name:        req.Name,
		Description: req.Description,
		Owner: &entities.User{
			ID: ownerID,
		},
		LeadName:  req.LeaderName,
		Type:      entities.GroupType(req.Type),
		Objective: req.Objective,
		Location:  req.Location,
		Members:   groupMembersEntityFromRequest(req.Members),
	}

	if entities.Faculty(faculty).IsValid() {
		eg.Faculty = entities.Faculty(faculty)
	}

	return eg
}
