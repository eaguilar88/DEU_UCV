package course_periods

import (
	"testing"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/stretchr/testify/assert"
)

func Test_createCoursePeriodRequestToEntitiesCoursePeriod(t *testing.T) {
	type testCase struct {
		name   string
		userID string
		req    CreateCoursePeriodRequest
		want   entities.CoursePeriod
	}
	tc := testCase{
		name:   "test",
		userID: "user-id",
		req: CreateCoursePeriodRequest{
			CourseID:        "course-id",
			StartDate:       "2023-01-01",
			EndDate:         "2023-12-31",
			InscriptionDate: "2023-01-01",
		},
		want: entities.CoursePeriod{
			Course: entities.Course{
				ID: "course-id",
				Owner: entities.User{
					ID: "user-id",
				},
			},
			StartDate:       "2023-01-01",
			EndDate:         "2023-12-31",
			InscriptionDate: "2023-01-01",
		},
	}
	t.Run(tc.name, func(t *testing.T) {
		got := createCoursePeriodRequestToEntitiesCoursePeriod(tc.req, tc.userID)
		assert.Equal(t, tc.want, got)
	})
}

func Test_updateCoursePeriodRequestToEntitiesCoursePeriod(t *testing.T) {
	type testCase struct {
		name   string
		userID string
		req    UpdateCoursePeriodRequest
		want   entities.CoursePeriod
	}
	tc := testCase{
		name:   "test",
		userID: "user-id",
		req: UpdateCoursePeriodRequest{
			ID:              "period-id",
			CourseID:        "course-id",
			StartDate:       "2023-01-01",
			EndDate:         "2023-12-31",
			InscriptionDate: "2023-01-01",
		},
		want: entities.CoursePeriod{
			ID: "period-id",
			Course: entities.Course{
				ID: "course-id",
				Owner: entities.User{
					ID: "user-id",
				},
			},
			StartDate:       "2023-01-01",
			EndDate:         "2023-12-31",
			InscriptionDate: "2023-01-01",
		},
	}
	t.Run(tc.name, func(t *testing.T) {
		got := updateCoursePeriodRequestToEntitiesCoursePeriod(tc.req, tc.userID)
		assert.Equal(t, tc.want, got)
	})
}
