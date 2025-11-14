package courses

import (
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/users"
)

func EntitiesCoursesToGetCoursesResponse(courses []entities.Course) []GetCourseResponse {
	var res []GetCourseResponse
	for _, course := range courses {
		res = append(res, EntitiesCourseToGetCourseResponse(course))
	}
	return res
}

func EntitiesCourseToGetCourseResponse(course entities.Course) GetCourseResponse {
	// Using the new course_request package
	owner := users.UserEntityToGetUserResponse(course.Owner)
	return GetCourseResponse{
		ID:          course.ID,
		Content:     course.Content,
		Cost:        course.Cost,
		CreatedAt:   course.CreatedAt,
		Description: course.Description,
		Duration:    course.Duration,
		Faculty:     string(course.Faculty),
		Location:    course.Location,
		Name:        course.Name,
		Objectives:  course.Objectives,
		Owner:       &owner,
		Type:        course.Type.String(),
		UpdatedAt:   course.UpdatedAt,
	}
}
