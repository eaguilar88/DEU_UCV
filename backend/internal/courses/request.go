package courses

import "github.com/eaguilar88/deu/internal/entities"

type GetCoursesRequest struct {
	PageScope entities.PageScope
}

type GetCourseRequest struct {
	ID string `path:"id"`
}

type CreateCourseRequest struct {
	Name              string `json:"nombre"`
	Description       string `json:"descripcion"`
	Objectives        string `json:"objetivos"`
	Rationale         string `json:"fundamentacion"`
	Duration          string `json:"duracion"`
	Cost              string `json:"estructura_costos"`
	InstructorProfile string `json:"perfil_docente"`
	Profiles          string `json:"perfiles"`
	Requirements      string `json:"exigencias"`
	Content           string `json:"estructura_curricular"`
	Evaluation        string `json:"evaluacion"`
	Schedule          string `json:"cronograma"`
	Type              string `json:"tipo"`
	Faculty           string `json:"facultad"`
	Location          string `json:"ubicacion"`
}

type DeleteCourseRequest struct {
	ID string `path:"id"`
}

type UpdateCourseRequest struct {
	ID                string `path:"id"`
	Name              string `json:"nombre"`
	Description       string `json:"descripcion"`
	Objectives        string `json:"objetivos"`
	Rationale         string `json:"fundamentacion"`
	Duration          string `json:"duracion"`
	Cost              string `json:"estructura_costos"`
	InstructorProfile string `json:"perfil_docente"`
	Profiles          string `json:"perfiles"`
	Requirements      string `json:"exigencias"`
	Content           string `json:"estructura_curricular"`
	Evaluation        string `json:"evaluacion"`
	Schedule          string `json:"cronograma"`
	Type              string `json:"tipo"`
	Faculty           string `json:"facultad"`
	Location          string `json:"ubicacion"`
}
