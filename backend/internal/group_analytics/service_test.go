package group_analytics_test

import (
	"context"
	"errors"
	"testing"

	"github.com/eaguilar88/deu/internal/group_analytics"
	"github.com/eaguilar88/deu/internal/group_analytics/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestService_GetAnalytics(t *testing.T) {
	type testCase struct {
		name    string
		req     group_analytics.AnalyticsFilterRequest
		prepare func(repoMock *mocks.MockRepository, req group_analytics.AnalyticsFilterRequest)
		want    *group_analytics.GroupAnalyticsResponse
		wantErr error
	}

	tests := []testCase{
		{
			name: "success with explicit year range",
			req: group_analytics.AnalyticsFilterRequest{
				CurrentYear: 2025,
				SinceYear:   2020,
				UntilYear:   2024,
				MasterAreas: []string{"Innovación"},
				GroupID:     "1",
			},
			prepare: func(repoMock *mocks.MockRepository, req group_analytics.AnalyticsFilterRequest) {
				repoMock.EXPECT().GetActivityParticipantsMetrics(mock.Anything, req.CurrentYear, req.GroupID).
					Return([]group_analytics.ActivityParticipantsMetric{{Lugar: "Actividad 1", CantidadReal: 5}}, nil)
				repoMock.EXPECT().GetActivitiesByStateMetrics(mock.Anything, req.CurrentYear, req.GroupID).
					Return([]group_analytics.ActivitiesByStateMetric{{Lugar: "Miranda", CantidadReal: 2}}, nil)
				repoMock.EXPECT().GetGroupParticipantsMetrics(mock.Anything, req.CurrentYear, req.GroupID).
					Return([]group_analytics.GroupParticipantsMetric{{Lugar: "Grupo 1", CantidadReal: 3}}, nil)
				repoMock.EXPECT().GetActivitiesByCityMetrics(mock.Anything, req.CurrentYear, req.GroupID).
					Return([]group_analytics.ActivitiesByCityMetric{{Lugar: "Caracas", CantidadReal: 4}}, nil)
				// explicit range passed through unchanged, no defaulting applied
				repoMock.EXPECT().GetYearlyActivitiesMetrics(mock.Anything, 2020, 2024, req.GroupID).
					Return([]group_analytics.YearlyActivitiesMetric{{Lugar: "2024", CantidadReal: 6}}, nil)
				repoMock.EXPECT().GetKnowledgeAreasByYearMetrics(mock.Anything, 2020, 2024, req.MasterAreas, req.GroupID).
					Return([]group_analytics.KnowledgeAreaYear{{Lugar: "2024", Areas: map[string]int{"Innovación": 2}}}, nil)
			},
			want: &group_analytics.GroupAnalyticsResponse{
				ActivityParticipants: []group_analytics.ActivityParticipantsMetric{{Lugar: "Actividad 1", CantidadReal: 5}},
				ActivitiesByState:    []group_analytics.ActivitiesByStateMetric{{Lugar: "Miranda", CantidadReal: 2}},
				GroupParticipants:    []group_analytics.GroupParticipantsMetric{{Lugar: "Grupo 1", CantidadReal: 3}},
				ActivitiesByCity:     []group_analytics.ActivitiesByCityMetric{{Lugar: "Caracas", CantidadReal: 4}},
				YearlyActivities:     []group_analytics.YearlyActivitiesMetric{{Lugar: "2024", CantidadReal: 6}},
				KnowledgeAreasByYear: []group_analytics.KnowledgeAreaYear{{Lugar: "2024", Areas: map[string]int{"Innovación": 2}}},
			},
		},
		{
			name: "defaults SinceYear/UntilYear when zero",
			req: group_analytics.AnalyticsFilterRequest{
				CurrentYear: 2025,
				GroupID:     "2",
			},
			prepare: func(repoMock *mocks.MockRepository, req group_analytics.AnalyticsFilterRequest) {
				repoMock.EXPECT().GetActivityParticipantsMetrics(mock.Anything, 2025, "2").Return(nil, nil)
				repoMock.EXPECT().GetActivitiesByStateMetrics(mock.Anything, 2025, "2").Return(nil, nil)
				repoMock.EXPECT().GetGroupParticipantsMetrics(mock.Anything, 2025, "2").Return(nil, nil)
				repoMock.EXPECT().GetActivitiesByCityMetrics(mock.Anything, 2025, "2").Return(nil, nil)
				// SinceYear=0 -> CurrentYear-3 = 2022; UntilYear=0 -> CurrentYear = 2025
				repoMock.EXPECT().GetYearlyActivitiesMetrics(mock.Anything, 2022, 2025, "2").Return(nil, nil)
				repoMock.EXPECT().GetKnowledgeAreasByYearMetrics(mock.Anything, 2022, 2025, req.MasterAreas, "2").Return(nil, nil)
			},
			want: &group_analytics.GroupAnalyticsResponse{},
		},
		{
			name: "activity participants error short-circuits before later calls",
			req:  group_analytics.AnalyticsFilterRequest{CurrentYear: 2025, GroupID: "3"},
			prepare: func(repoMock *mocks.MockRepository, req group_analytics.AnalyticsFilterRequest) {
				repoMock.EXPECT().GetActivityParticipantsMetrics(mock.Anything, 2025, "3").
					Return(nil, errors.New("boom"))
			},
			wantErr: errors.New("boom"),
		},
		{
			name: "activities by state error short-circuits before later calls",
			req:  group_analytics.AnalyticsFilterRequest{CurrentYear: 2025, GroupID: "4"},
			prepare: func(repoMock *mocks.MockRepository, req group_analytics.AnalyticsFilterRequest) {
				repoMock.EXPECT().GetActivityParticipantsMetrics(mock.Anything, 2025, "4").Return(nil, nil)
				repoMock.EXPECT().GetActivitiesByStateMetrics(mock.Anything, 2025, "4").
					Return(nil, errors.New("boom"))
			},
			wantErr: errors.New("boom"),
		},
		{
			name: "group participants error short-circuits before later calls",
			req:  group_analytics.AnalyticsFilterRequest{CurrentYear: 2025, GroupID: "5"},
			prepare: func(repoMock *mocks.MockRepository, req group_analytics.AnalyticsFilterRequest) {
				repoMock.EXPECT().GetActivityParticipantsMetrics(mock.Anything, 2025, "5").Return(nil, nil)
				repoMock.EXPECT().GetActivitiesByStateMetrics(mock.Anything, 2025, "5").Return(nil, nil)
				repoMock.EXPECT().GetGroupParticipantsMetrics(mock.Anything, 2025, "5").
					Return(nil, errors.New("boom"))
			},
			wantErr: errors.New("boom"),
		},
		{
			name: "activities by city error short-circuits before later calls",
			req:  group_analytics.AnalyticsFilterRequest{CurrentYear: 2025, GroupID: "6"},
			prepare: func(repoMock *mocks.MockRepository, req group_analytics.AnalyticsFilterRequest) {
				repoMock.EXPECT().GetActivityParticipantsMetrics(mock.Anything, 2025, "6").Return(nil, nil)
				repoMock.EXPECT().GetActivitiesByStateMetrics(mock.Anything, 2025, "6").Return(nil, nil)
				repoMock.EXPECT().GetGroupParticipantsMetrics(mock.Anything, 2025, "6").Return(nil, nil)
				repoMock.EXPECT().GetActivitiesByCityMetrics(mock.Anything, 2025, "6").
					Return(nil, errors.New("boom"))
			},
			wantErr: errors.New("boom"),
		},
		{
			name: "yearly activities error short-circuits before knowledge areas call",
			req:  group_analytics.AnalyticsFilterRequest{CurrentYear: 2025, GroupID: "7"},
			prepare: func(repoMock *mocks.MockRepository, req group_analytics.AnalyticsFilterRequest) {
				repoMock.EXPECT().GetActivityParticipantsMetrics(mock.Anything, 2025, "7").Return(nil, nil)
				repoMock.EXPECT().GetActivitiesByStateMetrics(mock.Anything, 2025, "7").Return(nil, nil)
				repoMock.EXPECT().GetGroupParticipantsMetrics(mock.Anything, 2025, "7").Return(nil, nil)
				repoMock.EXPECT().GetActivitiesByCityMetrics(mock.Anything, 2025, "7").Return(nil, nil)
				repoMock.EXPECT().GetYearlyActivitiesMetrics(mock.Anything, 2022, 2025, "7").
					Return(nil, errors.New("boom"))
			},
			wantErr: errors.New("boom"),
		},
		{
			name: "knowledge areas by year error",
			req:  group_analytics.AnalyticsFilterRequest{CurrentYear: 2025, GroupID: "8"},
			prepare: func(repoMock *mocks.MockRepository, req group_analytics.AnalyticsFilterRequest) {
				repoMock.EXPECT().GetActivityParticipantsMetrics(mock.Anything, 2025, "8").Return(nil, nil)
				repoMock.EXPECT().GetActivitiesByStateMetrics(mock.Anything, 2025, "8").Return(nil, nil)
				repoMock.EXPECT().GetGroupParticipantsMetrics(mock.Anything, 2025, "8").Return(nil, nil)
				repoMock.EXPECT().GetActivitiesByCityMetrics(mock.Anything, 2025, "8").Return(nil, nil)
				repoMock.EXPECT().GetYearlyActivitiesMetrics(mock.Anything, 2022, 2025, "8").Return(nil, nil)
				repoMock.EXPECT().GetKnowledgeAreasByYearMetrics(mock.Anything, 2022, 2025, req.MasterAreas, "8").
					Return(nil, errors.New("boom"))
			},
			wantErr: errors.New("boom"),
		},
	}

	ctx := context.Background()
	loggerMock := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			if tt.prepare != nil {
				tt.prepare(repoMock, tt.req)
			}
			s := group_analytics.NewService(repoMock, loggerMock)
			got, err := s.GetAnalytics(ctx, tt.req)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
