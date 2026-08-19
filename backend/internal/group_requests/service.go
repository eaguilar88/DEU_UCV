package group_requests

import (
	"context"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

type Repository interface {
	ApproveGroupRequest(ctx context.Context, reqID string) error
	RejectGroupRequest(ctx context.Context, reqID string) error
	GetGroupRequestsByFaculty(ctx context.Context, faculty entities.Faculty, status string, pageScope entities.PageScope) ([]entities.GroupRequest, entities.PageScope, int, error)
	GetGroupRequestByID(ctx context.Context, reqID string) (entities.GroupRequest, error)
	GetGroupRequestsByGroupID(ctx context.Context, groupID string) ([]entities.GroupRequest, error)
	GetPendingGroupRequestsCounts(ctx context.Context, faculty entities.Faculty) ([]entities.FacultyPendingCount, error)
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

func (s *service) GetGroupRequestsByFaculty(ctx context.Context, faculty entities.Faculty, status string, pageScope entities.PageScope) ([]entities.GroupRequest, entities.PageScope, int, error) {
	requests, ps, pendingCount, err := s.repo.GetGroupRequestsByFaculty(ctx, faculty, status, pageScope)
	if err != nil {
		s.logger.Error("failed to get group requests by faculty", zap.Error(err))
		return nil, entities.PageScope{}, 0, err
	}

	for i, req := range requests {
		approvals, err := s.repo.GetGroupRequestsByGroupID(ctx, req.GroupID)
		if err != nil {
			s.logger.Error("failed to get approvals for group request", zap.Error(err), zap.String("group_id", req.GroupID))
			return nil, entities.PageScope{}, 0, err
		}
		requests[i].Approvals = approvals
	}

	return requests, ps, pendingCount, nil
}

func (s *service) GetGroupRequestByID(ctx context.Context, reqID string) (entities.GroupRequest, error) {
	req, err := s.repo.GetGroupRequestByID(ctx, reqID)
	if err != nil {
		return entities.GroupRequest{}, err
	}

	req.Approvals, err = s.repo.GetGroupRequestsByGroupID(ctx, req.GroupID)
	if err != nil {
		s.logger.Error("failed to get approvals for group request", zap.Error(err), zap.String("group_id", req.GroupID))
		return entities.GroupRequest{}, err
	}

	return req, nil
}

func (s *service) GetPendingGroupRequestsCounts(ctx context.Context, faculty entities.Faculty) ([]entities.FacultyPendingCount, error) {
	return s.repo.GetPendingGroupRequestsCounts(ctx, faculty)
}
