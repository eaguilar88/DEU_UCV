package course_requests

import (
	"context"
	"errors"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

var (
	ErrRequestIsProcessed = errors.New("esta solicitud ya ha sido procesada por un administrador")
)

type Repository interface {
	ApproveCourseRequest(ctx context.Context, reqID, reviewerID, courseType, comments string) error
	RejectCourseRequest(ctx context.Context, reqID, reviewerID, comments string) error
	RedirectCourseRequest(ctx context.Context, reqID, reviewerID string, faculty entities.Faculty, reason string) error
	GetCourseRequestsByFaculty(ctx context.Context, faculty entities.Faculty, pageScope entities.PageScope) ([]entities.CourseRequest, entities.PageScope, error)
	GetCourseRequestByID(ctx context.Context, reqID string) (entities.CourseRequest, error)
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

func (s *service) ApproveCourseRequest(ctx context.Context, request entities.CourseRequest, courseType entities.CourseType) error {
	if existingRequest, err := s.repo.GetCourseRequestByID(ctx, request.ID); err != nil {
		return err
	} else if existingRequest.Status != entities.RequestStatus_UNDER_REVIEW {
		return ErrRequestIsProcessed
	}

	return s.repo.ApproveCourseRequest(ctx, request.ID, request.Reviewer.ID, courseType.String(), request.Comments)
}

func (s *service) RejectCourseRequest(ctx context.Context, reqID, reviewerID, comments string) error {
	if existingRequest, err := s.repo.GetCourseRequestByID(ctx, reqID); err != nil {
		return err
	} else if existingRequest.Status != entities.RequestStatus_UNDER_REVIEW {
		return ErrRequestIsProcessed
	}

	return s.repo.RejectCourseRequest(ctx, reqID, reviewerID, comments)
}

func (s *service) RedirectCourseRequest(ctx context.Context, reqID, reviewerID string, faculty entities.Faculty, reason string) error {
	if existingRequest, err := s.repo.GetCourseRequestByID(ctx, reqID); err != nil {
		return err
	} else if existingRequest.Status != entities.RequestStatus_UNDER_REVIEW {
		return ErrRequestIsProcessed
	}

	return s.repo.RedirectCourseRequest(ctx, reqID, reviewerID, faculty, reason)
}

func (s *service) GetCourseRequestsByFaculty(ctx context.Context, faculty entities.Faculty, pageScope entities.PageScope) ([]entities.CourseRequest, entities.PageScope, error) {
	cr, ps, err := s.repo.GetCourseRequestsByFaculty(ctx, faculty, pageScope)
	if err != nil {
		s.logger.Error("failed to get course requests by faculty", zap.Error(err))
		return nil, entities.PageScope{}, err
	}

	return cr, ps, nil
}

func (s *service) GetCourseRequestByID(ctx context.Context, reqID string) (entities.CourseRequest, error) {
	return s.repo.GetCourseRequestByID(ctx, reqID)
}
