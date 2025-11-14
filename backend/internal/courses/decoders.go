package courses

import (
	"github.com/eaguilar88/deu/internal/entities"
)

func createCourseRequestToEntitiesCourse(req CreateCourseRequest) (entities.Course, error) {
	faculty, err := entities.FromString(req.Faculty)
	if err != nil {
		faculty = entities.FacultyDEU
	}

	return entities.Course{
		Name:        req.Name,
		Description: req.Description,
		Objectives:  req.Objectives,
		Duration:    req.Duration,
		Content:     req.Content,
		Type:        entities.CourseType_Undefined,
		Faculty:     faculty,
		Cost:        req.Cost,
		Location:    req.Location,
	}, nil
}

func updateCourseRequestToEntitiesCourse(req UpdateCourseRequest, ID string) entities.Course {
	return entities.Course{
		ID:          ID,
		Name:        req.Name,
		Description: req.Description,
		Objectives:  req.Objectives,
		Content:     req.Content,
		Cost:        req.Cost,
		Location:    req.Location,
	}
}
