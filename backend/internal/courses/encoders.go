package courses

import (
	"github.com/eaguilar88/deu/internal/entities"
)

// coursesToResponse converts a slice of Course entities to responses.
func coursesToResponse(courses []entities.Course) []GetCourseResponse {
	var res []GetCourseResponse
	for _, course := range courses {
		res = append(res, courseToResponse(course))
	}
	return res
}

// courseToResponse converts a Course entity to GetCourseResponse.
func courseToResponse(course entities.Course) GetCourseResponse {
	var coverURL string
	if course.Cover != nil {
		coverURL = course.Cover.URL
	}

	return GetCourseResponse{
		ID:                course.ID,
		Name:              course.Name,
		Description:       course.Description,
		Cover:             coverURL,
		Objectives:        course.Objectives,
		Rationale:         course.Rationale,
		Duration:          course.Duration,
		Cost:              course.Cost,
		InstructorProfile: course.InstructorProfile,
		Profiles:          course.Profiles,
		Requirements:      course.Requirements,
		Content:           course.Content,
		Evaluation:        course.Evaluation,
		Schedule:          course.Schedule,
		ProviderID:        course.Owner.ID,
		Faculty:           string(course.Faculty),
		Location:          course.Location,
		Type:              course.Type.String(),
		CreatedAt:         course.CreatedAt,
		UpdatedAt:         course.UpdatedAt,
	}
}
