package groups

import "github.com/eaguilar88/deu/internal/entities"

// TODO: Implement request.go logic
type GetGroupRequest struct {
	ID string `param:"id" validate:"required"`
}
type GetGroupsRequest struct {
	Page    int `query:"page" validate:"required"`
	PerPage int `query:"per_page" validate:"required"`
}
type CreateGroupRequest struct {
	OwnerID       string           `validate:"required"`
	Name          string           `json:"name" validate:"required"`
	Description   string           `json:"description"`
	EndorsementID string           `json:"endorsement_id" validate:"required"`
	Objective     string           `json:"objective" validate:"required"`
	Location      string           `json:"location" validate:"required"`
	Members       []GroupMemberDTO `json:"members"`
}
type UpdateGroupRequest struct {
	ID            string           `param:"id" validate:"required"`
	OwnerID       string           `validate:"required"`
	Name          string           `json:"name" validate:"required"`
	Description   string           `json:"description"`
	EndorsementID string           `json:"endorsement_id" validate:"required"`
	Objective     string           `json:"objective" validate:"required"`
	Location      string           `json:"location" validate:"required"`
	Members       []GroupMemberDTO `json:"members"`
}
type DeleteGroupRequest struct {
	ID      string `param:"id" validate:"required"`
	OwnerID string `validate:"required"`
}

type GroupMemberDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name" validate:"required"`
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

func createGroupEntityFromRequest(req CreateGroupRequest) entities.ExtensionGroup {
	return entities.ExtensionGroup{
		Name:        req.Name,
		Description: req.Description,
		Owner: entities.User{
			ID: req.OwnerID,
		},
		Endorsement: entities.Endorsement{
			ID: req.EndorsementID,
		},
		Objective: req.Objective,
		Location:  req.Location,
		Members:   groupMembersEntityFromRequest(req.Members),
	}
}

func updateGroupEntityFromRequest(req UpdateGroupRequest) entities.ExtensionGroup {
	return entities.ExtensionGroup{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		Owner: entities.User{
			ID: req.OwnerID,
		},
		Endorsement: entities.Endorsement{
			ID: req.EndorsementID,
		},
		Objective: req.Objective,
		Location:  req.Location,
		Members:   groupMembersEntityFromRequest(req.Members),
	}
}
