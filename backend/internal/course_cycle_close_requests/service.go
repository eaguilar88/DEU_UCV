package course_cycle_close_requests

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/eaguilar88/deu/internal/email"
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
	ApproveCourseCycleCloseRequest(ctx context.Context, id, reviewerID, cycleID string, notify func() error) error
	RejectCourseCycleCloseRequest(ctx context.Context, id, reviewerID, comments, cycleID string, notify func() error) error

	// Notification lookups
	GetUser(ctx context.Context, userID string) (*entities.User, error)
	GetCoursePeriodByID(ctx context.Context, periodID string) (entities.CoursePeriod, error)
	GetCourse(ctx context.Context, courseID string) (entities.Course, error)

	// Files
	SaveFilesToDB(ctx context.Context, files []*entities.File) error
}

// StorageClient defines the file storage operations required by the course_cycle_close_requests service.
type StorageClient interface {
	UploadFile(ctx context.Context, files []*entities.File) error
}

// MailClient defines the email sending operations required by the course_cycle_close_requests service.
type MailClient interface {
	SendTemplate(ctx context.Context, to string, tmpl email.Template, data any) error
}

type service struct {
	repo        Repository
	storage     StorageClient
	emailClient MailClient
	logger      *zap.Logger
}

func NewService(repo Repository, storage StorageClient, emailClient MailClient, logger *zap.Logger) Service {
	return &service{
		repo:        repo,
		storage:     storage,
		emailClient: emailClient,
		logger:      logger,
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
	recipient, courseName, err := s.closeRequestRecipient(ctx, existing)
	if err != nil {
		return err
	}

	notify := s.notifier(ctx, id, recipient, email.TemplateCourseCycleCloseApproved, email.CourseCycleCloseApprovedData{
		CourseName: courseName,
	})
	return s.repo.ApproveCourseCycleCloseRequest(ctx, id, reviewerID, cycleID, notify)
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
	recipient, courseName, err := s.closeRequestRecipient(ctx, existing)
	if err != nil {
		return err
	}

	notify := s.notifier(ctx, id, recipient, email.TemplateCourseCycleCloseRejected, email.CourseCycleCloseRejectedData{
		CourseName: courseName,
		Reason:     comments,
	})
	return s.repo.RejectCourseCycleCloseRequest(ctx, id, reviewerID, comments, cycleID, notify)
}

// closeRequestRecipient resolves the email of the user who submitted the close request and the name
// of the course the closed cycle belongs to.
func (s *service) closeRequestRecipient(ctx context.Context, request entities.CourseCycleCloseRequest) (string, string, error) {
	requestID := strconv.FormatInt(request.ID, 10)

	submitter, err := s.repo.GetUser(ctx, request.SubmittedByID)
	if err != nil {
		s.logger.Error("failed to get close request submitter", zap.Error(err), zap.String("close_request_id", requestID))
		return "", "", err
	}

	period, err := s.repo.GetCoursePeriodByID(ctx, strconv.FormatInt(request.CourseCycleID, 10))
	if err != nil {
		s.logger.Error("failed to get close request course cycle", zap.Error(err), zap.String("close_request_id", requestID))
		return "", "", err
	}

	course, err := s.repo.GetCourse(ctx, period.Course.ID)
	if err != nil {
		s.logger.Error("failed to get close request course", zap.Error(err), zap.String("close_request_id", requestID))
		return "", "", err
	}

	return submitter.Email, course.Name, nil
}

// notifier returns the callback the repository runs inside its transaction, so a failed email rolls
// back the admin action.
func (s *service) notifier(ctx context.Context, requestID, to string, tmpl email.Template, data any) func() error {
	return func() error {
		if err := s.emailClient.SendTemplate(ctx, to, tmpl, data); err != nil {
			s.logger.Error("failed to send close request notification email", zap.Error(err), zap.String("close_request_id", requestID))
			return fmt.Errorf("error sending close request notification email: %w", err)
		}
		return nil
	}
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
