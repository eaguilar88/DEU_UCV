package courses

import (
	"fmt"
	"strings"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/utils"
	"github.com/labstack/echo/v4"
)

const (
	NameMissingError              = "nombre es requerido"
	DescriptionMissingError       = "descripcion es requerido"
	ObjectivesMissingError        = "objetivos es requerido"
	RationaleMissingError         = "fundamentacion es requerido"
	DurationMissingError          = "duracion es requerido"
	CostMissingError              = "estructura_costos es requerido"
	InstructorProfileMissingError = "perfil_docente es requerido"
	RequirementsMissingError      = "exigencias es requerido"
	ContentMissingError           = "estructura_curricular es requerido"
	EvaluationMissingError        = "evaluacion es requerido"
	ScheduleMissingError          = "cronograma es requerido"
	CoverMissingError             = "portada is required"
)

func formValue(c echo.Context, key string) string {
	return strings.TrimSpace(c.FormValue(key))
}

// toCourseEntity converts a multipart form request to a Course entity.
func toCourseEntity(c echo.Context) (entities.Course, error) {
	requiredFields := []struct{ key, msg string }{
		{"nombre", NameMissingError},
		{"descripcion", DescriptionMissingError},
		{"objetivos", ObjectivesMissingError},
		{"fundamentacion", RationaleMissingError},
		{"duracion", DurationMissingError},
		{"estructura_costos", CostMissingError},
		{"perfil_docente", InstructorProfileMissingError},
		{"exigencias", RequirementsMissingError},
		{"estructura_curricular", ContentMissingError},
		{"evaluacion", EvaluationMissingError},
		{"cronograma", ScheduleMissingError},
	}
	for _, f := range requiredFields {
		if formValue(c, f.key) == "" {
			return entities.Course{}, fmt.Errorf("%s", f.msg)
		}
	}

	faculty, err := entities.FromString(formValue(c, "facultad"))
	if err != nil {
		return entities.Course{}, fmt.Errorf("facultad is required")
	}

	course := entities.Course{
		Name:              formValue(c, "nombre"),
		Description:       formValue(c, "descripcion"),
		Objectives:        formValue(c, "objetivos"),
		Rationale:         formValue(c, "fundamentacion"),
		Duration:          formValue(c, "duracion"),
		Cost:              formValue(c, "estructura_costos"),
		InstructorProfile: formValue(c, "perfil_docente"),
		Profiles:          formValue(c, "perfiles"),
		Requirements:      formValue(c, "exigencias"),
		Content:           formValue(c, "estructura_curricular"),
		Evaluation:        formValue(c, "evaluacion"),
		Schedule:          formValue(c, "cronograma"),
		Type:              entities.FromStringCourseType(formValue(c, "tipo")),
		Faculty:           faculty,
		Location:          formValue(c, "ubicacion"),
	}

	cover, err := utils.GetFileFrom(c, entities.CourseFileTypeCover)
	if err != nil {
		return entities.Course{}, fmt.Errorf("portada is required")
	}
	course.Cover = cover
	return course, nil
}

// toCourseUpdateEntity converts UpdateCourseRequest to a Course entity.
func toCourseUpdateEntity(req UpdateCourseRequest, ID string) entities.Course {
	return entities.Course{
		ID:                ID,
		Name:              req.Name,
		Description:       req.Description,
		Objectives:        req.Objectives,
		Rationale:         req.Rationale,
		Duration:          req.Duration,
		Cost:              req.Cost,
		InstructorProfile: req.InstructorProfile,
		Profiles:          req.Profiles,
		Requirements:      req.Requirements,
		Content:           req.Content,
		Evaluation:        req.Evaluation,
		Schedule:          req.Schedule,
		Type:              entities.FromStringCourseType(req.Type),
		Location:          req.Location,
	}
}
