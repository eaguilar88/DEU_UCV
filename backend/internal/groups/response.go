package groups

import (
	"github.com/eaguilar88/deu/internal/course_request"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/users"
)

type GetGroupResponse struct {
	ID          string                                   `json:"id,omitempty"`
	Name        string                                   `json:"name,omitempty"`
	Description string                                   `json:"description,omitempty"`
	Owner       *users.GetUserResponse                   `json:"owner,omitempty"`
	Endorsement *course_request.GetCourseRequestResponse `json:"endorsement,omitempty"`
	Objective   string                                   `json:"objective,omitempty"`
	Location    string                                   `json:"location,omitempty"`
	Active      bool                                     `json:"active,omitempty"`
	CreatedAt   string                                   `json:"created_at,omitempty"`
	UpdatedAt   string                                   `json:"updated_at,omitempty"`
	DeletedAt   string                                   `json:"deleted_at,omitempty"`
}
type GetGroupsResponse struct {
	Groups    []GetGroupResponse `json:"groups"`
	PageScope entities.PageScope `json:"page"`
}
type CreateGroupResponse struct {
	ID string `json:"id"`
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
	owner := users.UserEntityToGetUserResponse(group.Owner)
	// Using the new course_request package
	endorsement := course_request.EntitiesCourseRequestToGetCourseRequestResponse(entities.CourseRequest{
		ID:          group.CourseRequest.ID,
		User:        group.CourseRequest.User,
		Reviewer:    group.CourseRequest.Reviewer,
		Status:      group.CourseRequest.Status,
		Type:        group.CourseRequest.Type,
		Name:        group.CourseRequest.Name,
		Description: group.CourseRequest.Description,
		Comments:    group.CourseRequest.Comments,
		ReviewedAt:  group.CourseRequest.ReviewedAt,
		CreatedAt:   group.CourseRequest.CreatedAt,
		UpdatedAtAt: group.CourseRequest.UpdatedAtAt,
	})
	return GetGroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		Owner:       &owner,
		Endorsement: &endorsement,
		Objective:   group.Objective,
		Location:    group.Location,
		Active:      group.Active,
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
	}
}
