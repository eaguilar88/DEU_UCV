package courses

import (
	"github.com/eaguilar88/deu/pkg/entities"
)

func createCourseRequestToEntitiesCourse(req CreateCourseRequest) entities.Course {
	return entities.Course{}
}

func updateCourseRequestToEntitiesCourse(req UpdateCourseRequest, ID int) entities.Course {

	return entities.Course{
		ID: ID,
	}
}
