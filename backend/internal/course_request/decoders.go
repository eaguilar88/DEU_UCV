package course_request

import (
	"github.com/eaguilar88/deu/internal/entities"
)

func createCourseRequestRequestToEntitiesCourseRequest(
	req CreateCourseRequestRequest,
) entities.CourseRequest {
	return entities.CourseRequest{
		User: entities.User{
			ID: req.UserID,
		},
		Status:      entities.CourseRequestStatus(req.Status),
		Type:        req.Type,
		Name:        req.Name,
		Description: req.Description,
		Comments:    req.Comments,
		CreatedAt:   req.CreatedAt,
		UpdatedAtAt: req.UpdatedAtAt,
	}
}

func updateCourseRequestRequestToEntitiesCourseRequest(
	req UpdateCourseRequestRequest,
) entities.CourseRequest {
	return entities.CourseRequest{
		ID: req.ID,
		User: entities.User{
			ID: req.UserID,
		},
		Status:      entities.CourseRequestStatus(req.Status),
		Type:        req.Type,
		Name:        req.Name,
		Description: req.Description,
		Comments:    req.Comments,
	}
}
