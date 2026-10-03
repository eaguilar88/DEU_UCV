package course_requests

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/eaguilar88/deu/internal/email"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/providers"
	"go.uber.org/zap"
)

var (
	ErrRequestIsProcessed = errors.New("esta solicitud ya ha sido procesada por un administrador")
)

type Repository interface {
	ApproveCourseRequest(ctx context.Context, request entities.CourseRequest, courseType string, notify func() error) error
	RejectCourseRequest(ctx context.Context, reqID, reviewerID, comments string, notify func() error) error
	RedirectCourseRequest(ctx context.Context, reqID, reviewerID string, faculty entities.Faculty, reason string, notify func() error) error
	GetCourseRequestsByFaculty(ctx context.Context, faculty entities.Faculty, pageScope entities.PageScope) ([]entities.CourseRequest, entities.PageScope, error)
	GetCourseRequestByID(ctx context.Context, reqID string) (entities.CourseRequest, error)
	GetProvider(ctx context.Context, providerID string) (entities.Provider, error)
	GetProviderByUserID(ctx context.Context, userID string) (entities.Provider, error)
	GetCourseRequestsByProvider(ctx context.Context, providerID string, pageScope entities.PageScope) ([]entities.CourseRequest, entities.PageScope, error)
	// A request under review has an inactive course, so the detail looks it up including inactive ones.
	GetCourseIncludingInactive(ctx context.Context, courseID string) (entities.Course, error)
	GetFilesByOwner(ctx context.Context, ownerID string, ownerType entities.OwnerType) (entities.GroupedFiles, error)
}

// StorageClient defines the file storage operations required by the course_requests service.
type StorageClient interface {
	UploadFile(ctx context.Context, file []*entities.File) error
	DeleteFile(ctx context.Context, objectKey string) error
	GetFileURL(ctx context.Context, objectKey string) (string, error)
	GetPresignedFileURL(ctx context.Context, objectKey string) (string, error)
}

// MailClient defines the email sending operations required by the course_requests service.
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

func (s *service) ApproveCourseRequest(ctx context.Context, request entities.CourseRequest, courseType entities.CourseType, viewer entities.Viewer) error {
	existing, recipient, err := s.getPendingRequestAndRecipient(ctx, request.ID, viewer)
	if err != nil {
		return err
	}

	// The evaluation document is uploaded first; its metadata is saved in the approval transaction.
	if f := request.EvaluationFile; f != nil {
		f.OwnerID = request.ID
		f.OwnerType = entities.OwnerTypeCourseRequest
		f.Key = fmt.Sprintf("files/course-requests/%s/%s", request.ID, f.Name)
		f.Purpose = entities.CourseRequestFileTypeEvaluation
		f.Public = false
		f.UploadedBy = request.Reviewer.ID
		f.CreatedAt = time.Now().Format(time.RFC3339)
		if err := s.storage.UploadFile(ctx, []*entities.File{f}); err != nil {
			s.logger.Error("failed to upload evaluation file", zap.Error(err), zap.String("course_request_id", request.ID))
			return err
		}
	}

	notify := s.notifier(ctx, request.ID, recipient, email.TemplateCourseRequestApproved, email.CourseRequestApprovedData{
		CourseName: courseName(existing),
		Comments:   request.Comments,
	})
	if err := s.repo.ApproveCourseRequest(ctx, request, courseType.String(), notify); err != nil {
		if f := request.EvaluationFile; f != nil {
			if delErr := s.storage.DeleteFile(ctx, f.Key); delErr != nil {
				s.logger.Warn("failed to remove evaluation file of failed approval", zap.Error(delErr), zap.String("key", f.Key))
			}
		}
		return err
	}
	return nil
}

func (s *service) RejectCourseRequest(ctx context.Context, reqID, reviewerID, comments string, viewer entities.Viewer) error {
	existing, recipient, err := s.getPendingRequestAndRecipient(ctx, reqID, viewer)
	if err != nil {
		return err
	}

	notify := s.notifier(ctx, reqID, recipient, email.TemplateCourseRequestRejected, email.CourseRequestRejectedData{
		CourseName: courseName(existing),
		Reason:     comments,
	})
	return s.repo.RejectCourseRequest(ctx, reqID, reviewerID, comments, notify)
}

func (s *service) RedirectCourseRequest(ctx context.Context, reqID, reviewerID string, faculty entities.Faculty, reason string, viewer entities.Viewer) error {
	existing, recipient, err := s.getPendingRequestAndRecipient(ctx, reqID, viewer)
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

// getPendingRequestAndRecipient loads a course request, checks the viewer may review it and that it is
// still under review, and resolves the email of the provider that owns the course, so the admin action
// can notify them.
func (s *service) getPendingRequestAndRecipient(ctx context.Context, reqID string, viewer entities.Viewer) (entities.CourseRequest, string, error) {
	existing, err := s.repo.GetCourseRequestByID(ctx, reqID)
	if err != nil {
		return entities.CourseRequest{}, "", err
	}
	if !canReview(viewer, existing) {
		return entities.CourseRequest{}, "", ErrCourseRequestForbidden
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

// canReview reports whether viewer may act on the request: only the course's current faculty, not the
// one it was redirected from, which still lists it.
func canReview(viewer entities.Viewer, request entities.CourseRequest) bool {
	if viewer.IsGlobalAdmin() {
		return true
	}
	return request.Course != nil && viewer.IsFacultyAdminOf(request.Course.Faculty)
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

// GetCourseRequestByID returns the request with its full course (cover and facilitator CV
// included) and its evaluation file, so reviewers see the whole proposal.
func (s *service) GetCourseRequestByID(ctx context.Context, reqID string) (entities.CourseRequest, error) {
	request, err := s.repo.GetCourseRequestByID(ctx, reqID)
	if err != nil {
		return entities.CourseRequest{}, err
	}
	logger := s.logger.With(zap.String("course_request_id", reqID))

	if request.Course != nil {
		course, err := s.repo.GetCourseIncludingInactive(ctx, request.Course.ID)
		if err != nil {
			logger.Error("failed to get course request course", zap.Error(err))
			return entities.CourseRequest{}, err
		}
		files, err := s.repo.GetFilesByOwner(ctx, course.ID, entities.OwnerTypeCourse)
		if err != nil {
			logger.Error("failed to get course files", zap.Error(err))
			return entities.CourseRequest{}, err
		}
		if course.Cover, err = s.fileWithURL(ctx, files.GetSingleFile(entities.CourseFileTypeCover)); err != nil {
			logger.Error("failed to get course cover URL", zap.Error(err))
			return entities.CourseRequest{}, err
		}
		if course.FacilitatorCV, err = s.fileWithURL(ctx, files.GetSingleFile(entities.CourseFileTypeFacilitatorCV)); err != nil {
			logger.Error("failed to get facilitator CV URL", zap.Error(err))
			return entities.CourseRequest{}, err
		}
		request.Course = &course
	}

	files, err := s.repo.GetFilesByOwner(ctx, reqID, entities.OwnerTypeCourseRequest)
	if err != nil {
		logger.Error("failed to get course request files", zap.Error(err))
		return entities.CourseRequest{}, err
	}
	if request.EvaluationFile, err = s.fileWithURL(ctx, files.GetSingleFile(entities.CourseRequestFileTypeEvaluation)); err != nil {
		logger.Error("failed to get evaluation file URL", zap.Error(err))
		return entities.CourseRequest{}, err
	}

	return request, nil
}

// fileWithURL fills in the URL of f: public files go through the /files proxy, private ones get a
// short-lived pre-signed URL. A nil file stays nil.
func (s *service) fileWithURL(ctx context.Context, f *entities.File) (*entities.File, error) {
	if f == nil {
		return nil, nil
	}
	getURL := s.storage.GetPresignedFileURL
	if f.Public {
		getURL = s.storage.GetFileURL
	}
	url, err := getURL(ctx, f.Key)
	if err != nil {
		return nil, err
	}
	f.URL = url
	return f, nil
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

	requests, scope, err := s.repo.GetCourseRequestsByProvider(ctx, provider.ID, pageScope)
	if err != nil {
		return nil, entities.PageScope{}, err
	}

	// The provider's list shows each proposal with its cover, approved or not.
	for i := range requests {
		course := requests[i].Course
		if course == nil {
			continue
		}
		files, err := s.repo.GetFilesByOwner(ctx, course.ID, entities.OwnerTypeCourse)
		if err != nil {
			s.logger.Warn("failed to get course files", zap.Error(err), zap.String("course_id", course.ID))
			continue
		}
		if course.Cover, err = s.fileWithURL(ctx, files.GetSingleFile(entities.CourseFileTypeCover)); err != nil {
			s.logger.Warn("failed to get course cover URL", zap.Error(err), zap.String("course_id", course.ID))
		}
	}

	return requests, scope, nil
}
