package course_periods

import (
	"testing"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/users"
	"github.com/stretchr/testify/assert"
)

func TestEntitiesCoursePeriodToGetCoursePeriodResponse(t *testing.T) {
	type testCase struct {
		name string
		cp   entities.CoursePeriod
		want GetCoursePeriodResponse
	}
	tc := testCase{
		name: "test",
		cp: entities.CoursePeriod{
			ID: "period-id",
			Participants: []entities.User{
				{
					ID:        "user-id",
					FirstName: "John",
					LastName:  "Doe",
				},
			},
			StartDate:       "2023-01-01",
			EndDate:         "2023-12-31",
			InscriptionDate: "2023-01-01",
		},
		want: GetCoursePeriodResponse{
			ID: "period-id",
			Participans: []users.GetUserResponse{
				{
					ID:        "user-id",
					FirstName: "John",
					LastName:  "Doe",
				},
			},
			StartDate:       "2023-01-01",
			EndDate:         "2023-12-31",
			InscriptionDate: "2023-01-01",
		},
	}
	t.Run(tc.name, func(t *testing.T) {
		got := EntitiesCoursePeriodToGetCoursePeriodResponse(tc.cp)
		assert.Equal(t, tc.want, got)
	})
}

func TestEntitiesCoursePeriodsToGetCoursePeriodsResponse(t *testing.T) {
	type testCase struct {
		name string
		cps  []entities.CoursePeriod
		want []GetCoursePeriodResponse
	}
	tc := testCase{
		name: "test multiple course periods",
		cps: []entities.CoursePeriod{
			{
				ID: "period-id-1",
				Participants: []entities.User{
					{
						ID:        "user-id-1",
						FirstName: "John",
						LastName:  "Doe",
					},
				},
				StartDate:       "2023-01-01",
				EndDate:         "2023-06-30",
				InscriptionDate: "2023-01-01",
			},
			{
				ID: "period-id-2",
				Participants: []entities.User{
					{
						ID:        "user-id-2",
						FirstName: "Jane",
						LastName:  "Smith",
					},
				},
				StartDate:       "2023-07-01",
				EndDate:         "2023-12-31",
				InscriptionDate: "2023-07-01",
			},
		},
		want: []GetCoursePeriodResponse{
			{
				ID: "period-id-1",
				Participans: []users.GetUserResponse{
					{
						ID:        "user-id-1",
						FirstName: "John",
						LastName:  "Doe",
					},
				},
				StartDate:       "2023-01-01",
				EndDate:         "2023-06-30",
				InscriptionDate: "2023-01-01",
			},
			{
				ID: "period-id-2",
				Participans: []users.GetUserResponse{
					{
						ID:        "user-id-2",
						FirstName: "Jane",
						LastName:  "Smith",
					},
				},
				StartDate:       "2023-07-01",
				EndDate:         "2023-12-31",
				InscriptionDate: "2023-07-01",
			},
		},
	}
	t.Run(tc.name, func(t *testing.T) {
		got := EntitiesCoursePeriodsToGetCoursePeriodsResponse(tc.cps)
		assert.Equal(t, tc.want, got)
	})
}
