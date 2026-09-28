package course_requests

import (
	"context"
	"errors"
	"fmt"

	"github.com/eaguilar88/deu/internal/email"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/providers"
	"go.uber.org/zap"
)

var (
	ErrRequestIsProcessed = errors.New("esta solicitud ya ha sido procesada por un administrador")
)

type Repository interface {
	ApproveCourseRequest(ctx context.Context, reqID, reviewerID, courseType, comments string, notify func() error) error
	RejectCourseRequest(ctx context.Context, reqID, reviewerID, comments string, notify func() error) error
	RedirectCourseRequest(ctx context.Context, reqID, reviewerID string, faculty entities.Faculty, reason string, notify func() error) error
	GetCourseRequestsByFaculty(ctx context.Context, faculty entities.Faculty, pageScope entities.PageScope) ([]entities.CourseRequest, entities.PageScope, error)
	GetCourseRequestByID(ctx context.Context, reqID string) (entities.CourseRequest, error)
	GetProvider(ctx context.Context, providerID string) (entities.Provider, error)
	GetProviderByUserID(ctx context.Context, userID string) (entities.Provider, error)
	GetCourseRequestsByProvider(ctx context.Context, providerID string, pageScope entities.PageScope) ([]entities.CourseRequest, entities.PageScope, error)
}

// MailClient defines the email sending operations required by the course_requests service.
type MailClient interface {
	SendTemplate(ctx context.Context, to string, tmpl email.Template, data any) error
}

type service struct {
	repo        Repository
	emailClient MailClient
	logger      *zap.Logger
}

func NewService(repo Repository, emailClient MailClient, logger *zap.Logger) Service {
	return &service{
		repo:        repo,
		emailClient: emailClient,
		logger:      logger,
	}
}

func (s *service) ApproveCourseRequest(ctx context.Context, request entities.CourseRequest, courseType entities.CourseType) error {
	existing, recipient, err := s.getPendingRequestAndRecipient(ctx, request.ID)
	if err != nil {
		return err
	}

	notify := s.notifier(ctx, request.ID, recipient, email.TemplateCourseRequestApproved, email.CourseRequestApprovedData{
		CourseName: courseName(existing),
		Comments:   request.Comments,
	})
	return s.repo.ApproveCourseRequest(ctx, request.ID, request.Reviewer.ID, courseType.String(), request.Comments, notify)
}

func (s *service) RejectCourseRequest(ctx context.Context, reqID, reviewerID, comments string) error {
	existing, recipient, err := s.getPendingRequestAndRecipient(ctx, reqID)
	if err != nil {
		return err
	}

	notify := s.notifier(ctx, reqID, recipient, email.TemplateCourseRequestRejected, email.CourseRequestRejectedData{
		CourseName: courseName(existing),
		Reason:     comments,
	})
	return s.repo.RejectCourseRequest(ctx, reqID, reviewerID, comments, notify)
}

func (s *service) RedirectCourseRequest(ctx context.Context, reqID, reviewerID string, faculty entities.Faculty, reason string) error {
	existing, recipient, err := s.getPendingRequestAndRecipient(ctx, reqID)
	if err != nil {
		return err
	}

	notify := s.notifier(ctx, reqID, recipient, email.TemplateCourseRequestRedirected, email.CourseRequestRedirectedData{
		CourseName: courseName(existing),
		Faculty:    faculty.String(),
		Reason:     reason,
	})
	return s.repo.RedirectCourseRequest(ctx, reqID, reviewerID, faculty, reason, notify)
}

// getPendingRequestAndRecipient loads a course request, checks it is still under review, and resolves
// the email of the provider that owns the course, so the admin action can notify them.
func (s *service) getPendingRequestAndRecipient(ctx context.Context, reqID string) (entities.CourseRequest, string, error) {
	existing, err := s.repo.GetCourseRequestByID(ctx, reqID)
	if err != nil {
		return entities.CourseRequest{}, "", err
	}
	if existing.Status != entities.RequestStatus_UNDER_REVIEW {
		return entities.CourseRequest{}, "", ErrRequestIsProcessed
	}

	var providerID string
	if existing.Course != nil {
		providerID = existing.Course.Owner.ID
	}
	provider, err := s.repo.GetProvider(ctx, providerID)
	if err != nil {
		s.logger.Error("failed to get course request provider", zap.Error(err), zap.String("course_request_id", reqID))
		return entities.CourseRequest{}, "", err
	}

	return existing, provider.User.Email, nil
}

// notifier returns the callback the repository runs inside its transaction, so a failed email rolls
// back the admin action.
func (s *service) notifier(ctx context.Context, reqID, to string, tmpl email.Template, data any) func() error {
	return func() error {
		if err := s.emailClient.SendTemplate(ctx, to, tmpl, data); err != nil {
			s.logger.Error("failed to send course request notification email", zap.Error(err), zap.String("course_request_id", reqID))
			return fmt.Errorf("error sending course request notification email: %w", err)
		}
		return nil
	}
}

func courseName(request entities.CourseRequest) string {
	if request.Course == nil {
		return ""
	}
	return request.Course.Name
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

func (s *service) GetMyCourseRequests(ctx context.Context, userID string, pageScope entities.PageScope) ([]entities.CourseRequest, entities.PageScope, error) {
	provider, err := s.repo.GetProviderByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, providers.ErrProviderNotFound) {
			return nil, pageScope, nil
		}
		s.logger.Error("failed to get provider by user ID", zap.Error(err), zap.String("user_id", userID))
		return nil, entities.PageScope{}, err
	}

	return s.repo.GetCourseRequestsByProvider(ctx, provider.ID, pageScope)
}
