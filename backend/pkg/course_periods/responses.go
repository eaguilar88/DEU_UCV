package course_periods

import (
	"github.com/eaguilar88/deu/pkg/courses"
	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/eaguilar88/deu/pkg/users"
)

type GetCoursePeriodResponse struct {
	ID              int                       `json:"id"`
	Course          courses.GetCourseResponse `json:"course"`
	Participans     []users.GetUserResponse   `json:"participans"`
	StartDate       string                    `json:"start_date"`
	EndDate         string                    `json:"end_date"`
	InscriptionDate string                    `json:"inscription_date"`
	CreatedAt       string                    `json:"created_at"`
	UpdatedAt       string                    `json:"updated_at"`
	DeletedAt       string                    `json:"deleted_at"`
}

func EntitiesCoursePeriodToGetCoursePeriodResponse(coursePeriod entities.CoursePeriod) GetCoursePeriodResponse {

	return GetCoursePeriodResponse{
		ID:              coursePeriod.ID,
		Course:          courses.EntitiesCourseToGetCourseResponse(coursePeriod.Course),
		Participans:     users.UserEntitiesToGetUserResponse(coursePeriod.Participants),
		StartDate:       coursePeriod.StartDate,
		EndDate:         coursePeriod.EndDate,
		InscriptionDate: coursePeriod.InscriptionDate,
		CreatedAt:       coursePeriod.CreatedAt,
		UpdatedAt:       coursePeriod.UpdatedAt,
		DeletedAt:       coursePeriod.DeletedAt,
	}
}

type GetCoursePeriodsResponse struct {
	Periods []GetCoursePeriodResponse `json:"periods"`
	Pages   entities.PageScope        `json:"pages"`
}

func EntitiesCoursePeriodsToGetCoursePeriodsResponse(coursePeriods []entities.CoursePeriod) []GetCoursePeriodResponse {
	var out = make([]GetCoursePeriodResponse, 0, len(coursePeriods))
	for _, coursePeriod := range coursePeriods {
		out = append(out, EntitiesCoursePeriodToGetCoursePeriodResponse(coursePeriod))
	}
	return out
}

type CreateCoursePeriodResponse struct {
	ID string `json:"id"`
}
