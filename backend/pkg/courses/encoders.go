package courses

import (
	"github.com/eaguilar88/deu/pkg/endorsements"
	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/eaguilar88/deu/pkg/users"
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
		Endorsement: endorsements.EntitiesEndorsementToGetEndorsementResponse(course.Endorsement),
		Location:    course.Location,
		Name:        course.Name,
		Objectives:  course.Objectives,
		Owner:       users.UserEntityToGetUserResponse(course.Owner),
		UpdatedAt:   course.UpdatedAt,
	}
}
