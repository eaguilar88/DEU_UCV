package course_periods

import (
	"github.com/eaguilar88/deu/pkg/entities"
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
	CourseID        int    `param:"course_id"`
	StartDate       string `json:"start_date"`
	EndDate         string `json:"end_date"`
	InscriptionDate string `json:"inscription_date"`
}

func createCoursePeriodRequestToEntitiesCoursePeriod(req CreateCoursePeriodRequest, userID int) entities.CoursePeriod {
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
	ID              int    `param:"id"`
	CourseID        int    `json:"course_id"`
	StartDate       string `json:"start_date"`
	EndDate         string `json:"end_date"`
	InscriptionDate string `json:"inscription_date"`
}

func updateCoursePeriodRequestToEntitiesCoursePeriod(req UpdateCoursePeriodRequest, userID int) entities.CoursePeriod {
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
	CourseID string `param:"course_id"`
	ID       string `param:"id"`
}
