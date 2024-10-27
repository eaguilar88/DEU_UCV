package courses

import "github.com/eaguilar88/deu/pkg/entities"

func entitiesCourseToGetCourseResponse(course entities.Course) GetCourseResponse {
	return GetCourseResponse{
		ID: course.ID,
	}
}
