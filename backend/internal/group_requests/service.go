package group_requests

import (
	"context"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

type Repository interface {
	ApproveGroupRequest(ctx context.Context, reqID string) error
	RejectGroupRequest(ctx context.Context, reqID string) error
	GetGroupRequestsByFaculty(ctx context.Context, faculty entities.Faculty, pageScope entities.PageScope) ([]entities.GroupAuthRequest, entities.PageScope, error)
	GetGroupRequestByID(ctx context.Context, reqID string) (entities.GroupAuthRequest, error)
}
type GroupRequestService struct {
	repo   Repository
	logger *zap.Logger
}

func NewGroupRequestService(repo Repository, logger *zap.Logger) *GroupRequestService {
	return &GroupRequestService{
		repo:   repo,
		logger: logger,
	}
}

func (s *GroupRequestService) ApproveGroupRequest(ctx context.Context, reqID string) error {
	return s.repo.ApproveGroupRequest(ctx, reqID)
}

func (s *GroupRequestService) RejectGroupRequest(ctx context.Context, reqID string) error {
	return s.repo.RejectGroupRequest(ctx, reqID)
}

func (s *GroupRequestService) GetGroupRequestsByFaculty(ctx context.Context, faculty entities.Faculty, pageScope entities.PageScope) ([]entities.GroupAuthRequest, entities.PageScope, error) {
	cr, ps, err := s.repo.GetGroupRequestsByFaculty(ctx, faculty, pageScope)
	if err != nil {
		s.logger.Error("failed to get group requests by faculty", zap.Error(err))
		return nil, entities.PageScope{}, err
	}

	return cr, ps, nil
}

func (s *GroupRequestService) GetGroupRequestByID(ctx context.Context, reqID string) (entities.GroupAuthRequest, error) {
	return s.repo.GetGroupRequestByID(ctx, reqID)
}
