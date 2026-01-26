package courses

import (
	"github.com/eaguilar88/deu/internal/entities"
)

func EntitiesCoursesToGetCoursesResponse(courses []entities.Course) []GetCourseResponse {
	var res []GetCourseResponse
	for _, course := range courses {
		res = append(res, EntitiesCourseToGetCourseResponse(course))
	}
	return res
}

func EntitiesCourseToGetCourseResponse(course entities.Course) GetCourseResponse {
	return GetCourseResponse{
		ID:          course.ID,
		Content:     course.Content,
		Cost:        course.Cost,
		CreatedAt:   course.CreatedAt,
		Description: course.Description,
		Duration:    course.Duration,
		Faculty:     string(course.Faculty),
		Name:        course.Name,
		Objectives:  course.Objectives,
		Type:        course.Type.String(),
		UpdatedAt:   course.UpdatedAt,
	}
}
