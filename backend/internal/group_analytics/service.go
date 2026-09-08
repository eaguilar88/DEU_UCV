package group_analytics

import (
	"context"

	"go.uber.org/zap"
)

type Repository interface {
	GetActivityParticipantsMetrics(ctx context.Context, year int, groupID string) ([]ActivityParticipantsMetric, error)
	GetActivitiesByStateMetrics(ctx context.Context, year int, groupID string) ([]ActivitiesByStateMetric, error)
	GetGroupParticipantsMetrics(ctx context.Context, year int, groupID string) ([]GroupParticipantsMetric, error)
	GetActivitiesByCityMetrics(ctx context.Context, year int, groupID string) ([]ActivitiesByCityMetric, error)
	GetYearlyActivitiesMetrics(ctx context.Context, startYear, endYear int, groupID string) ([]YearlyActivitiesMetric, error)
	GetKnowledgeAreasByYearMetrics(ctx context.Context, startYear, endYear int, masterAreas []string, groupID string) ([]map[string]interface{}, error)
}

type Service interface {
	GetAnalytics(ctx context.Context, req AnalyticsFilterRequest) (*GroupAnalyticsResponse, error)
}

type service struct {
	repo   Repository
	logger *zap.Logger
}

func NewService(repo Repository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

func (s *service) GetAnalytics(ctx context.Context, req AnalyticsFilterRequest) (*GroupAnalyticsResponse, error) {
	startYear := req.SinceYear
	if startYear == 0 {
		startYear = req.CurrentYear - 3
	}

	endYear := req.UntilYear
	if endYear == 0 {
		endYear = req.CurrentYear
	}

	actPart, err := s.repo.GetActivityParticipantsMetrics(ctx, req.CurrentYear, req.GroupID)
	if err != nil {
		s.logger.Error("error al obtener participantes por actividad", zap.Error(err))
		return nil, err
	}

	actState, err := s.repo.GetActivitiesByStateMetrics(ctx, req.CurrentYear, req.GroupID)
	if err != nil {
		s.logger.Error("error al obtener actividades por estado", zap.Error(err))
		return nil, err
	}

	grpPart, err := s.repo.GetGroupParticipantsMetrics(ctx, req.CurrentYear, req.GroupID)
	if err != nil {
		s.logger.Error("error al obtener participantes por grupo", zap.Error(err))
		return nil, err
	}

	actCity, err := s.repo.GetActivitiesByCityMetrics(ctx, req.CurrentYear, req.GroupID)
	if err != nil {
		s.logger.Error("error al obtener actividades por ciudad", zap.Error(err))
		return nil, err
	}

	yearlyAct, err := s.repo.GetYearlyActivitiesMetrics(ctx, startYear, endYear, req.GroupID)
	if err != nil {
		s.logger.Error("error al obtener histórico anual", zap.Error(err))
		return nil, err
	}

	areasByYear, err := s.repo.GetKnowledgeAreasByYearMetrics(ctx, startYear, endYear, req.MasterAreas, req.GroupID)
	if err != nil {
		s.logger.Error("error al obtener áreas por año", zap.Error(err))
		return nil, err
	}

	return &GroupAnalyticsResponse{
		ActivityParticipants: actPart,
		ActivitiesByState:    actState,
		GroupParticipants:    grpPart,
		ActivitiesByCity:     actCity,
		YearlyActivities:     yearlyAct,
		KnowledgeAreasByYear: areasByYear,
	}, nil
}