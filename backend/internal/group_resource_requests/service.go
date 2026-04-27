package group_resource_requests

import (
	"context"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

type Repository interface {
	CreateGroupResourceRequest(ctx context.Context, req entities.GroupResourceRequest) (int64, error)
	ApproveGroupResourceRequest(ctx context.Context, reqID string) error
	RejectGroupResourceRequest(ctx context.Context, reqID string) error
	GetGroupResourceRequestsByFaculty(ctx context.Context, faculty entities.Faculty, pageScope entities.PageScope) ([]entities.GroupResourceRequest, entities.PageScope, error)
	GetGroupResourceRequestByID(ctx context.Context, reqID string) (entities.GroupResourceRequest, error)
}

type service struct {
	repo   Repository
	logger *zap.Logger
}

func NewService(repo Repository, logger *zap.Logger) Service {
	return &service{repo: repo, logger: logger}
}

func (s *service) CreateGroupResourceRequest(ctx context.Context, req entities.GroupResourceRequest) (int64, error) {
	id, err := s.repo.CreateGroupResourceRequest(ctx, req)
	if err != nil {
		s.logger.Error("failed to create group resource request", zap.Error(err))
		return 0, err
	}
	return id, nil
}

func (s *service) ApproveGroupResourceRequest(ctx context.Context, reqID string) error {
	return s.repo.ApproveGroupResourceRequest(ctx, reqID)
}

func (s *service) RejectGroupResourceRequest(ctx context.Context, reqID string) error {
	return s.repo.RejectGroupResourceRequest(ctx, reqID)
}

func (s *service) GetGroupResourceRequestsByFaculty(ctx context.Context, faculty entities.Faculty, pageScope entities.PageScope) ([]entities.GroupResourceRequest, entities.PageScope, error) {
	reqs, ps, err := s.repo.GetGroupResourceRequestsByFaculty(ctx, faculty, pageScope)
	if err != nil {
		s.logger.Error("failed to get group resource requests by faculty", zap.Error(err))
		return nil, entities.PageScope{}, err
	}
	return reqs, ps, nil
}

func (s *service) GetGroupResourceRequestByID(ctx context.Context, reqID string) (entities.GroupResourceRequest, error) {
	return s.repo.GetGroupResourceRequestByID(ctx, reqID)
}
