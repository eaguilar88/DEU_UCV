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
	Page     int    `                  query:"page"`
	PerPage  int    `                  query:"per_page"`
}

type CreateCoursePeriodRequest struct {
	CourseID        string `param:"course_id"`
	StartDate       string `                  json:"start_date"`
	EndDate         string `                  json:"end_date"`
	InscriptionDate string `                  json:"inscription_date"`
}

func createCoursePeriodRequestToEntitiesCoursePeriod(
	req CreateCoursePeriodRequest,
	userID string,
) entities.CoursePeriod {
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
	CourseID        string `           json:"course_id"`
	StartDate       string `           json:"start_date"`
	EndDate         string `           json:"end_date"`
	InscriptionDate string `           json:"inscription_date"`
}

func updateCoursePeriodRequestToEntitiesCoursePeriod(
	req UpdateCoursePeriodRequest,
	userID string,
) entities.CoursePeriod {
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
