package group_requests

import (
	"context"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

type Repository interface {
	ApproveGroupRequest(ctx context.Context, reqID string) error
	RejectGroupRequest(ctx context.Context, reqID string) error
	GetGroupRequestsByFaculty(ctx context.Context, faculty entities.Faculty, pageScope entities.PageScope) ([]entities.GroupRequest, entities.PageScope, error)
	GetGroupRequestByID(ctx context.Context, reqID string) (entities.GroupRequest, error)
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

func (s *service) ApproveGroupRequest(ctx context.Context, reqID string) error {
	return s.repo.ApproveGroupRequest(ctx, reqID)
}

func (s *service) RejectGroupRequest(ctx context.Context, reqID string) error {
	return s.repo.RejectGroupRequest(ctx, reqID)
}

func (s *service) GetGroupRequestsByFaculty(ctx context.Context, faculty entities.Faculty, pageScope entities.PageScope) ([]entities.GroupRequest, entities.PageScope, error) {
	cr, ps, err := s.repo.GetGroupRequestsByFaculty(ctx, faculty, pageScope)
	if err != nil {
		s.logger.Error("failed to get group requests by faculty", zap.Error(err))
		return nil, entities.PageScope{}, err
	}

	return cr, ps, nil
}

func (s *service) GetGroupRequestByID(ctx context.Context, reqID string) (entities.GroupRequest, error) {
	return s.repo.GetGroupRequestByID(ctx, reqID)
}
