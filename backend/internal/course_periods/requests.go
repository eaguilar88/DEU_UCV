package course_periods

import (
	"github.com/eaguilar88/deu/internal/entities"
)

type GetCoursePeriodRequest struct {
	CourseID string `param:"course_id"`
	ID       string `param:"id"`
}

type GetCoursePeriodsRequest struct {
	CourseID string `param:"course_id"`
	Page     int    `query:"page"`
	PerPage  int    `query:"per_page"`
}

type CreateCoursePeriodRequest struct {
	CourseID        string `param:"course_id"`
	StartDate       string `json:"fecha_inicio"`
	EndDate         string `json:"fecha_fin"`
	InscriptionDate string `json:"fecha_inscripcion"`
}

// toPeriodEntity converts CreateCoursePeriodRequest to a CoursePeriod entity.
func toPeriodEntity(req CreateCoursePeriodRequest, userID string) entities.CoursePeriod {
	return entities.CoursePeriod{
		Course: entities.Course{
			ID: req.CourseID,
			Owner: entities.User{
				ID: userID,
			},
		},
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		InscriptionDate: req.InscriptionDate,
	}
}

type UpdateCoursePeriodRequest struct {
	ID              string `param:"id"`
	CourseID        string `json:"curso_id"`
	StartDate       string `json:"fecha_inicio"`
	EndDate         string `json:"fecha_fin"`
	InscriptionDate string `json:"fecha_inscripcion"`
}

// toPeriodUpdateEntity converts UpdateCoursePeriodRequest to a CoursePeriod entity.
func toPeriodUpdateEntity(req UpdateCoursePeriodRequest, userID string) entities.CoursePeriod {
	return entities.CoursePeriod{
		ID: req.ID,
		Course: entities.Course{
			ID: req.CourseID,
			Owner: entities.User{
				ID: userID,
			},
		},
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		InscriptionDate: req.InscriptionDate,
	}
}

type DeleteCoursePeriodRequest struct {
	ID       string `param:"id"`
	CourseID string `param:"course_id"`
}

// Announcement requests
type GetAnnouncementRequest struct {
	PeriodID string `param:"period_id"`
	ID       string `param:"id"`
}

type CreateAnnouncementRequest struct {
	PeriodID string `param:"period_id" validate:"required"`
	Title    string `json:"titulo" validate:"required"`
	Content  string `json:"contenido" validate:"required"`
}

// toAnnouncementEntity converts CreateAnnouncementRequest to an Announcement entity.
func toAnnouncementEntity(req CreateAnnouncementRequest) entities.Announcement {
	return entities.Announcement{
		Title:   req.Title,
		Content: req.Content,
	}
}

type UpdateAnnouncementRequest struct {
	PeriodID string `param:"period_id"`
	ID       string `param:"id"`
	Title    string `json:"titulo" validate:"required"`
	Content  string `json:"contenido" validate:"required"`
}

// toAnnouncementUpdateEntity converts UpdateAnnouncementRequest to an Announcement entity.
func toAnnouncementUpdateEntity(req UpdateAnnouncementRequest) entities.Announcement {
	return entities.Announcement{
		ID:      req.ID,
		Title:   req.Title,
		Content: req.Content,
	}
}

type DeleteAnnouncementRequest struct {
	PeriodID string `param:"period_id"`
	ID       string `param:"id"`
}
