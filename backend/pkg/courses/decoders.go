package courses

import (
	"github.com/eaguilar88/deu/pkg/entities"
)

func createCourseRequestToEntitiesCourse(req CreateCourseRequest) entities.Course {
	return entities.Course{
		Owner: entities.User{
			ID: req.UserID,
		},
		Endorsement: entities.Endorsement{
			ID: req.EndorsementID,
		},
		Name:        req.Name,
		Description: req.Description,
		Objectives:  req.Objectives,
		Content:     req.Content,
		Cost:        req.Cost,
		Location:    req.Location,
	}
}

func updateCourseRequestToEntitiesCourse(req UpdateCourseRequest, ID string) entities.Course {

	return entities.Course{
		ID:          ID,
		Name:        req.Name,
		Description: req.Description,
		Objectives:  req.Objectives,
		Content:     req.Content,
		Cost:        req.Cost,
		Location:    req.Location,
	}
}
