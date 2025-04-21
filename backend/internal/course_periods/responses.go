package course_periods

import (
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/users"
)

type GetCoursePeriodResponse struct {
	ID              string                  `json:"id"`
	Participans     []users.GetUserResponse `json:"participans,omitempty"`
	StartDate       string                  `json:"start_date"`
	EndDate         string                  `json:"end_date"`
	InscriptionDate string                  `json:"inscription_date"`
	CreatedAt       string                  `json:"created_at"`
	UpdatedAt       string                  `json:"updated_at"`
	DeletedAt       string                  `json:"deleted_at,omitempty"`
}

func EntitiesCoursePeriodToGetCoursePeriodResponse(
	coursePeriod entities.CoursePeriod,
) GetCoursePeriodResponse {
	return GetCoursePeriodResponse{
		ID:              coursePeriod.ID,
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

func EntitiesCoursePeriodsToGetCoursePeriodsResponse(
	coursePeriods []entities.CoursePeriod,
) []GetCoursePeriodResponse {
	out := make([]GetCoursePeriodResponse, 0, len(coursePeriods))
	for _, coursePeriod := range coursePeriods {
		out = append(out, EntitiesCoursePeriodToGetCoursePeriodResponse(coursePeriod))
	}
	return out
}

type CreateCoursePeriodResponse struct {
	ID string `json:"id"`
}

type UpdateCoursePeriodResponse struct{}

type DeleteCoursePeriodResponse struct{}
