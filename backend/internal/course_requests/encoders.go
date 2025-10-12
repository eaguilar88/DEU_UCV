package course_requests

import (
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/users"
)

func CourseRequestEntitiesToGetCourseRequestsResponse(
	courseRequests []entities.CourseRequest,
) []GetCourseRequestResponse {
	responses := make([]GetCourseRequestResponse, 0, len(courseRequests))
	for _, courseRequest := range courseRequests {
		responses = append(responses, EntitiesCourseRequestToGetCourseRequestResponse(courseRequest))
	}
	return responses
}

func EntitiesCourseRequestToGetCourseRequestResponse(
	courseRequest entities.CourseRequest,
) GetCourseRequestResponse {
	user := users.UserEntityToGetUserResponse(courseRequest.User)
	reviewer := users.UserEntityToGetUserResponse(courseRequest.Reviewer)
	return GetCourseRequestResponse{
		ID:          courseRequest.ID,
		User:        &user,
		Reviewer:    &reviewer,
		Status:      courseRequest.Status,
		Type:        courseRequest.Type,
		Name:        courseRequest.Name,
		Description: courseRequest.Description,
		Comments:    courseRequest.Comments,
		CreatedAt:   courseRequest.CreatedAt,
		UpdatedAtAt: courseRequest.UpdatedAtAt,
	}
}
