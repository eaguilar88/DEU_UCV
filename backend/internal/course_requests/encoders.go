package course_requests

import "github.com/eaguilar88/deu/internal/entities"

func courseRequestToResponse(req entities.CourseRequest) GetCourseRequestResponse {
	response := GetCourseRequestResponse{
		ID:             req.ID,
		Status:         req.Status,
		Comments:       req.Comments,
		ReviewedAt:     req.ReviewedAt,
		CreatedAt:      req.CreatedAt,
		UpdatedAt:      req.UpdatedAt,
		Score:          req.Score,
		Classification: req.Classification,
		OwnerUserID:    req.User.ID,
	}
	if req.EvaluationFile != nil {
		response.EvaluationFile = req.EvaluationFile.URL
	}
	if req.Course != nil {
		response.Course = courseToInfo(*req.Course)
	}
	return response
}

func courseToInfo(course entities.Course) *CourseInfo {
	info := &CourseInfo{
		ID:                course.ID,
		Name:              course.Name,
		Description:       course.Description,
		Objectives:        course.Objectives,
		Duration:          course.Duration,
		Content:           course.Content,
		Type:              course.Type.String(),
		Faculty:           course.Faculty.String(),
		Cost:              course.Cost,
		Location:          course.Location,
		IsActive:          course.IsActive,
		HasDocumentation:  course.HasDocumentation,
		CreatedAt:         course.CreatedAt,
		UpdatedAt:         course.UpdatedAt,
		ProviderID:        course.Owner.ID,
		OriginFaculty:     course.OriginFaculty.String(),
		Rationale:         course.Rationale,
		InstructorProfile: course.InstructorProfile,
		Profiles:          course.Profiles,
		Requirements:      course.Requirements,
		Evaluation:        course.Evaluation,
		Schedule:          course.Schedule,
		Competencies:      course.Competencies,
		Bibliography:      course.Bibliography,
	}
	if course.Cover != nil {
		info.Cover = course.Cover.URL
	}
	if course.FacilitatorCV != nil {
		info.FacilitatorCV = course.FacilitatorCV.URL
	}
	return info
}
