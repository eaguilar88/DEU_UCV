package courses

import (
	"fmt"
	"strings"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/utils"
	"github.com/labstack/echo/v4"
)

func formValue(c echo.Context, key string) string {
	return strings.TrimSpace(c.FormValue(key))
}

// toCourseEntity converts a multipart form request to a Course entity.
func toCourseEntity(c echo.Context) (entities.Course, error) {
	requiredFields := []struct{ key, msg string }{
		{"nombre", "nombre es requerido"},
		{"descripcion", "descripcion es requerido"},
		{"objetivos", "objetivos es requerido"},
		{"fundamentacion", "fundamentacion es requerido"},
		{"duracion", "duracion es requerido"},
		{"estructura_costos", "estructura_costos es requerido"},
		{"perfil_docente", "perfil_docente es requerido"},
		{"exigencias", "exigencias es requerido"},
		{"estructura_curricular", "estructura_curricular es requerido"},
		{"evaluacion", "evaluacion es requerido"},
		{"cronograma", "cronograma es requerido"},
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
