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
	var coverURL, cvURL string
	if course.Cover != nil {
		coverURL = course.Cover.URL
	}
	if course.FacilitatorCV != nil {
		cvURL = course.FacilitatorCV.URL
	}

	response := GetCourseResponse{
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
		OriginFaculty:     string(course.OriginFaculty),
		Location:          course.Location,
		Type:              course.Type.String(),
		CreatedAt:         course.CreatedAt,
		UpdatedAt:         course.UpdatedAt,
		IsActive:          course.IsActive,
		HasDocumentation:  course.HasDocumentation,
		ManagementStatus:  string(course.ManagementStatus),
		Competencies:      course.Competencies,
		Bibliography:      course.Bibliography,
		FacilitatorCV:     cvURL,
	}
	if course.Provider != nil {
		response.Provider = &CourseProviderInfo{Name: course.Provider.Name, LogoURL: course.Provider.LogoURL}
		response.OwnerUserID = course.Provider.UserID
	}
	return response
}
