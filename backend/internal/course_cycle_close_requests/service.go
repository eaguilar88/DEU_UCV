package course_cycle_close_requests

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

var (
	ErrRequestIsProcessed = errors.New("esta solicitud ya ha sido procesada por un administrador")
)

type Repository interface {
	CreateCourseCycleCloseRequest(ctx context.Context, cycleID, submittedByID int64) (int64, error)
	HasPendingCloseRequestForCycle(ctx context.Context, cycleID int64) (bool, error)
	GetCourseCycleCloseRequests(ctx context.Context, faculty entities.Faculty, pageScope entities.PageScope) ([]entities.CourseCycleCloseRequest, entities.PageScope, error)
	GetCourseCycleCloseRequestByID(ctx context.Context, id string) (entities.CourseCycleCloseRequest, error)
	ApproveCourseCycleCloseRequest(ctx context.Context, id, reviewerID, cycleID string) error
	RejectCourseCycleCloseRequest(ctx context.Context, id, reviewerID, comments, cycleID string) error

	// Files
	SaveFilesToDB(ctx context.Context, files []*entities.File) error
}

// StorageClient defines the file storage operations required by the course_cycle_close_requests service.
type StorageClient interface {
	UploadFile(ctx context.Context, files []*entities.File) error
}

type service struct {
	repo    Repository
	storage StorageClient
	logger  *zap.Logger
}

func NewService(repo Repository, storage StorageClient, logger *zap.Logger) Service {
	return &service{
		repo:    repo,
		storage: storage,
		logger:  logger,
	}
}

func (s *service) SubmitCloseRequest(ctx context.Context, request entities.CourseCycleCloseRequest, submittedByID string) (int64, error) {
	submitterID, err := strconv.ParseInt(submittedByID, 10, 64)
	if err != nil {
		return -1, fmt.Errorf("invalid submitter ID: %w", err)
	}

	pending, err := s.repo.HasPendingCloseRequestForCycle(ctx, request.CourseCycleID)
	if err != nil {
		s.logger.Error("failed to check for pending close requests", zap.Error(err))
		return -1, err
	}
	if pending {
		return -1, ErrCloseRequestAlreadyPending
	}

	id, err := s.repo.CreateCourseCycleCloseRequest(ctx, request.CourseCycleID, submitterID)
	if err != nil {
		s.logger.Error("failed to create cycle close request", zap.Error(err))
		return -1, err
	}

	files := []*entities.File{request.ParticipantsFile, request.VouchersFile, request.SurveyFile}
	idStr := strconv.FormatInt(id, 10)
	for _, f := range files {
		f.OwnerID = idStr
		f.OwnerType = entities.OwnerTypeCourseCycleCloseRequest
		f.Key = fmt.Sprintf("files/course-cycle-close-requests/%s/%s", idStr, f.Name)
		f.Public = false
		f.UploadedBy = submittedByID
		f.CreatedAt = time.Now().Format(time.RFC3339)
	}

	if err := s.storage.UploadFile(ctx, files); err != nil {
		s.logger.Error("failed to upload close request evidence files", zap.Error(err), zap.String("close_request_id", idStr))
		return -1, err
	}

	if err := s.repo.SaveFilesToDB(ctx, files); err != nil {
		s.logger.Error("failed to save close request evidence file metadata", zap.Error(err), zap.String("close_request_id", idStr))
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

	cycleID := strconv.FormatInt(existing.CourseCycleID, 10)
	return s.repo.RejectCourseCycleCloseRequest(ctx, id, reviewerID, comments, cycleID)
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
