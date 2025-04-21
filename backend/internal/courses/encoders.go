package courses

import (
	"github.com/eaguilar88/deu/internal/endorsements"
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
	endorsement := endorsements.EntitiesEndorsementToGetEndorsementResponse(course.Endorsement)
	owner := users.UserEntityToGetUserResponse(course.Owner)
	return GetCourseResponse{
		ID:          course.ID,
		Content:     course.Content,
		Cost:        course.Cost,
		CreatedAt:   course.CreatedAt,
		Description: course.Description,
		Endorsement: &endorsement,
		Location:    course.Location,
		Name:        course.Name,
		Objectives:  course.Objectives,
		Owner:       &owner,
		UpdatedAt:   course.UpdatedAt,
	}
}
