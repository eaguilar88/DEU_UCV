package course_cycle_close_requests

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

var (
	ErrRequestIsProcessed = errors.New("esta solicitud ya ha sido procesada por un administrador")
)

type Repository interface {
	CreateCourseCycleCloseRequest(ctx context.Context, cycleID, submittedByID int64) (int64, error)
	GetCourseCycleCloseRequests(ctx context.Context, faculty entities.Faculty, pageScope entities.PageScope) ([]entities.CourseCycleCloseRequest, entities.PageScope, error)
	GetCourseCycleCloseRequestByID(ctx context.Context, id string) (entities.CourseCycleCloseRequest, error)
	ApproveCourseCycleCloseRequest(ctx context.Context, id, reviewerID, cycleID string) error
	RejectCourseCycleCloseRequest(ctx context.Context, id, reviewerID, comments string) error
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

func (s *service) SubmitCloseRequest(ctx context.Context, cycleID int64, submittedByID string) (int64, error) {
	submitterID, err := strconv.ParseInt(submittedByID, 10, 64)
	if err != nil {
		return -1, fmt.Errorf("invalid submitter ID: %w", err)
	}

	id, err := s.repo.CreateCourseCycleCloseRequest(ctx, cycleID, submitterID)
	if err != nil {
		s.logger.Error("failed to create cycle close request", zap.Error(err))
		return -1, err
	}
	return id, nil
}

func (s *service) ApproveCloseRequest(ctx context.Context, id, reviewerID string) error {
	existing, err := s.repo.GetCourseCycleCloseRequestByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get cycle close request", zap.Error(err))
		return err
	}
	if existing.Status != entities.RequestStatus_UNDER_REVIEW {
		return ErrRequestIsProcessed
	}

	cycleID := strconv.FormatInt(existing.CourseCycleID, 10)
	return s.repo.ApproveCourseCycleCloseRequest(ctx, id, reviewerID, cycleID)
}

func (s *service) RejectCloseRequest(ctx context.Context, id, reviewerID, comments string) error {
	existing, err := s.repo.GetCourseCycleCloseRequestByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get cycle close request", zap.Error(err))
		return err
	}
	if existing.Status != entities.RequestStatus_UNDER_REVIEW {
		return ErrRequestIsProcessed
	}

	return s.repo.RejectCourseCycleCloseRequest(ctx, id, reviewerID, comments)
}

func (s *service) GetCloseRequests(ctx context.Context, faculty entities.Faculty, pageScope entities.PageScope) ([]entities.CourseCycleCloseRequest, entities.PageScope, error) {
	requests, ps, err := s.repo.GetCourseCycleCloseRequests(ctx, faculty, pageScope)
	if err != nil {
		s.logger.Error("failed to get cycle close requests", zap.Error(err))
		return nil, entities.PageScope{}, err
	}
	return requests, ps, nil
}

func (s *service) GetCloseRequestByID(ctx context.Context, id string) (entities.CourseCycleCloseRequest, error) {
	return s.repo.GetCourseCycleCloseRequestByID(ctx, id)
}
