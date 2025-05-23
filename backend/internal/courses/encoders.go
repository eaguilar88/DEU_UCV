package courses

import (
	"github.com/eaguilar88/deu/internal/course_request"
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
	endorsement := course_request.EntitiesCourseRequestToGetCourseRequestResponse(entities.CourseRequest{
		ID:          course.CourseRequest.ID,
		User:        course.CourseRequest.User,
		Reviewer:    course.CourseRequest.Reviewer,
		Status:      course.CourseRequest.Status,
		Type:        course.CourseRequest.Type,
		Name:        course.CourseRequest.Name,
		Description: course.CourseRequest.Description,
		Comments:    course.CourseRequest.Comments,
		ReviewedAt:  course.CourseRequest.ReviewedAt,
		CreatedAt:   course.CourseRequest.CreatedAt,
		UpdatedAtAt: course.CourseRequest.UpdatedAtAt,
	})
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
