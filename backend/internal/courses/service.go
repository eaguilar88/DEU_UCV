package courses

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/eaguilar88/deu/internal/email"
	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

type Repository interface {
	GetCourse(ctx context.Context, courseID string) (entities.Course, error)
	GetCourseIncludingInactive(ctx context.Context, courseID string) (entities.Course, error)
	GetCourses(ctx context.Context, filter entities.CourseFilter, pageScope entities.PageScope) ([]entities.Course, entities.PageScope, error)
	GetPublicCourses(ctx context.Context, pageScope entities.PageScope) ([]entities.Course, entities.PageScope, error)
	GetProviderByUserID(ctx context.Context, userID string) (entities.Provider, error)
	GetProvider(ctx context.Context, providerID string) (entities.Provider, error)
	GetLatestCoursePeriod(ctx context.Context, courseID string) (entities.CoursePeriod, error)
	GetFacultyCoordinatorEmails(ctx context.Context, faculty entities.Faculty) ([]string, error)
	GetDEUAdminEmails(ctx context.Context) ([]string, error)

	// Unit of Work: Atomic operations
	CreateCourseWithRequest(ctx context.Context, course entities.Course) (courseID int64, requestID int64, err error)

	// Individual operations (for flexibility)
	CreateCourse(ctx context.Context, course entities.Course) (int64, error)
	UpdateCourse(ctx context.Context, courseID string, user entities.Course) error
	DeleteCourse(ctx context.Context, courseID string) error

	// Course Requests
	CreateCourseRequest(ctx context.Context, request entities.CourseRequest) (int64, error)

	// Files
	GetFilesByOwner(ctx context.Context, ownerID string, ownerType entities.OwnerType) (entities.GroupedFiles, error)
	SaveFilesToDB(ctx context.Context, file []*entities.File) error
}

// StorageClient defines the file storage operations required by the courses service.
type StorageClient interface {
	UploadFile(ctx context.Context, file []*entities.File) error
	DeleteFile(ctx context.Context, objectKey string) error
	GetFileURL(ctx context.Context, objectKey string) (string, error)
	GetPresignedFileURL(ctx context.Context, objectKey string) (string, error)
	GetObject(ctx context.Context, objectKey string) (io.ReadCloser, string, error)
	GetFileMetadata(ctx context.Context, objectKey string) (map[string]string, error)
}

// MailClient defines the email sending operations required by the courses service.
type MailClient interface {
	SendTemplate(ctx context.Context, to string, tmpl email.Template, data any) error
}

type service struct {
	repo        Repository
	storage     StorageClient
	emailClient MailClient
	log         *zap.Logger
}

func NewService(repository Repository, storage StorageClient, emailClient MailClient, logger *zap.Logger) Service {
	return &service{
		repo:        repository,
		storage:     storage,
		emailClient: emailClient,
		log:         logger,
	}
}

// GetCourse returns an approved course to anyone. A course still under review (or rejected) is
// only returned to its provider and to the admins who review it; anyone else gets
// ErrCourseNotFound.
func (s *service) GetCourse(ctx context.Context, courseID string, viewer entities.Viewer) (entities.Course, error) {
	course, err := s.repo.GetCourse(ctx, courseID)
	if errors.Is(err, ErrCourseNotFound) && !viewer.IsAnonymous() {
		course, err = s.unapprovedCourse(ctx, courseID, viewer)
	}
	if err != nil {
		return entities.Course{}, err
	}

	files, err := s.repo.GetFilesByOwner(ctx, courseID, entities.OwnerTypeCourse)
	if err != nil {
		s.log.Error("failed to get course files", zap.Error(err), zap.String("course_id", courseID))
		return entities.Course{}, err
	}
	if cover := files.GetSingleFile(entities.CourseFileTypeCover); cover != nil {
		url, err := s.storage.GetFileURL(ctx, cover.Key)
		if err != nil {
			s.log.Error("failed to get cover URL", zap.Error(err), zap.String("course_id", courseID))
			return entities.Course{}, err
		}
		cover.URL = url
		course.Cover = cover
	}
	// The CV is private, so it gets a short-lived pre-signed URL instead of the /files proxy.
	if cv := files.GetSingleFile(entities.CourseFileTypeFacilitatorCV); cv != nil {
		url, err := s.storage.GetPresignedFileURL(ctx, cv.Key)
		if err != nil {
			s.log.Error("failed to get facilitator CV URL", zap.Error(err), zap.String("course_id", courseID))
			return entities.Course{}, err
		}
		cv.URL = url
		course.FacilitatorCV = cv
	}

	provider, err := s.courseProvider(ctx, course.Owner.ID)
	if err != nil {
		// The provider summary is decorative: the course is still returned without it.
		s.log.Warn("failed to get course provider", zap.Error(err), zap.String("course_id", courseID))
	} else {
		course.Provider = provider
	}

	return course, nil
}

// unapprovedCourse returns a course that is not active yet, if viewer is its provider or an
// admin of its faculty.
func (s *service) unapprovedCourse(ctx context.Context, courseID string, viewer entities.Viewer) (entities.Course, error) {
	course, err := s.repo.GetCourseIncludingInactive(ctx, courseID)
	if err != nil {
		return entities.Course{}, err
	}
	if viewer.IsGlobalAdmin() || viewer.IsFacultyAdminOf(course.Faculty, course.OriginFaculty) {
		return course, nil
	}
	provider, err := s.repo.GetProvider(ctx, course.Owner.ID)
	if err != nil {
		s.log.Error("failed to get course provider", zap.Error(err), zap.String("course_id", courseID))
		return entities.Course{}, err
	}
	if provider.User.ID != viewer.UserID {
		return entities.Course{}, ErrCourseNotFound
	}
	return course, nil
}

// courseProvider loads the public name and logo of the provider that owns a course.
func (s *service) courseProvider(ctx context.Context, providerID string) (*entities.CourseProviderSummary, error) {
	if providerID == "" {
		return nil, nil
	}
	provider, err := s.repo.GetProvider(ctx, providerID)
	if err != nil {
		return nil, err
	}
	summary := &entities.CourseProviderSummary{UserID: provider.User.ID, Name: provider.Name}
	if summary.Name == "" {
		summary.Name = strings.TrimSpace(provider.User.FirstName + " " + provider.User.LastName)
	}

	files, err := s.repo.GetFilesByOwner(ctx, providerID, entities.OwnerTypeProvider)
	if err != nil {
		return nil, err
	}
	// Only the logo is public; the provider's documents are never exposed here.
	if logo := files.GetSingleFile(entities.ProviderFileTypeLogo); logo != nil && logo.Public {
		url, err := s.storage.GetFileURL(ctx, logo.Key)
		if err != nil {
			return nil, err
		}
		summary.LogoURL = url
	}
	return summary, nil
}

func (s *service) GetLatestCoursePeriod(ctx context.Context, courseID string) (entities.CoursePeriod, error) {
	period, err := s.repo.GetLatestCoursePeriod(ctx, courseID)
	if err != nil {
		return entities.CoursePeriod{}, err
	}
	return period, nil
}

// GetCourses lists approved courses. Admins (root, deu_admin, faculty_admin) and a provider
// listing their own courses (usuario_id = the caller) see every management status; anyone
// else only sees open or closed courses.
func (s *service) GetCourses(ctx context.Context, filter entities.CourseFilter, viewer entities.Viewer, pageScope entities.PageScope) ([]entities.Course, entities.PageScope, error) {
	filter.VisibleStatuses = nil
	if !canSeeAllCourseStatuses(filter, viewer) {
		filter.VisibleStatuses = []entities.CourseManagementStatus{
			entities.CourseManagementStatusOpen,
			entities.CourseManagementStatusClosed,
		}
	}

	courses, page, err := s.repo.GetCourses(ctx, filter, pageScope)
	if err != nil {
		return nil, entities.PageScope{}, err
	}

	s.attachCoverURLs(ctx, courses)

	return courses, page, nil
}

func canSeeAllCourseStatuses(filter entities.CourseFilter, viewer entities.Viewer) bool {
	if viewer.IsGlobalAdmin() || viewer.IsFacultyAdmin() {
		return true
	}
	return filter.OwnerUserID != "" && filter.OwnerUserID == viewer.UserID
}

func (s *service) GetPublicCourses(ctx context.Context, pageScope entities.PageScope) ([]entities.Course, entities.PageScope, error) {
	courses, page, err := s.repo.GetPublicCourses(ctx, pageScope)
	if err != nil {
		return nil, entities.PageScope{}, err
	}

	s.attachCoverURLs(ctx, courses)

	return courses, page, nil
}

// attachCoverURLs enriches each course in place with its cover image URL, when one exists.
func (s *service) attachCoverURLs(ctx context.Context, courses []entities.Course) {
	for i := range courses {
		files, err := s.repo.GetFilesByOwner(ctx, courses[i].ID, entities.OwnerTypeCourse)
		if err != nil {
			s.log.Error("failed to get course files", zap.Error(err), zap.String("course_id", courses[i].ID))
			continue
		}
		if cover := files.GetSingleFile(entities.CourseFileTypeCover); cover != nil {
			url, err := s.storage.GetFileURL(ctx, cover.Key)
			if err != nil {
				s.log.Error("failed to get cover URL", zap.Error(err), zap.String("course_id", courses[i].ID))
				continue
			}
			cover.URL = url
			courses[i].Cover = cover
		}
	}
}

func (s *service) CreateCourse(ctx context.Context, userID string, course entities.Course) (int64, error) {
	provider, err := s.repo.GetProviderByUserID(ctx, userID)
	if err != nil {
		s.log.Error("failed to get provider by user ID", zap.Error(err), zap.String("user_id", userID))
		return -1, err
	}

	course.Owner.ID = provider.ID
	// The faculty is inherited from the provider, never taken from the client.
	// OriginFaculty is immutable: redirects only change Faculty.
	course.Faculty = provider.Faculty
	if course.Faculty == "" {
		course.Faculty = entities.FacultyDEU
	}
	course.OriginFaculty = course.Faculty
	courseID, requestID, err := s.repo.CreateCourseWithRequest(ctx, course)
	if err != nil {
		s.log.Error("failed to create course with request", zap.Error(err))
		return -1, err
	}

	s.log.Info("course and course request created successfully",
		zap.Int64("course_id", courseID),
		zap.Int64("request_id", requestID))

	// The cover is shown publicly; the facilitator CV is only for reviewers.
	if err := s.uploadCourseFile(ctx, courseID, userID, course.Cover, entities.CourseFileTypeCover, true); err != nil {
		return -1, err
	}
	if err := s.uploadCourseFile(ctx, courseID, userID, course.FacilitatorCV, entities.CourseFileTypeFacilitatorCV, false); err != nil {
		return -1, err
	}

	s.notifyCourseRequest(ctx, course, provider)

	return courseID, nil
}

// notifyCourseRequest tells the faculty's coordinators and the DEU admins that a provider submitted
// a course for approval. It is best-effort: the request is already created, so failures are only
// logged.
func (s *service) notifyCourseRequest(ctx context.Context, course entities.Course, provider entities.Provider) {
	providerName := provider.Name
	if providerName == "" {
		providerName = strings.TrimSpace(provider.User.FirstName + " " + provider.User.LastName)
	}
	data := email.CourseRequestSubmittedData{
		CourseName:   course.Name,
		ProviderName: providerName,
		Faculty:      string(course.Faculty),
	}

	coordinators, err := s.repo.GetFacultyCoordinatorEmails(ctx, course.Faculty)
	if err != nil {
		s.log.Warn("failed to get faculty coordinators", zap.Error(err), zap.String("faculty", string(course.Faculty)))
	}
	s.sendToAll(ctx, coordinators, email.TemplateCourseRequestSubmittedFaculty, data)

	admins, err := s.repo.GetDEUAdminEmails(ctx)
	if err != nil {
		s.log.Warn("failed to get DEU admins", zap.Error(err))
	}
	s.sendToAll(ctx, admins, email.TemplateCourseRequestSubmittedDEU, data)
}

func (s *service) sendToAll(ctx context.Context, recipients []string, tmpl email.Template, data any) {
	if len(recipients) == 0 {
		s.log.Warn("no recipients for notification", zap.String("template", string(tmpl)))
		return
	}
	for _, to := range recipients {
		if err := s.emailClient.SendTemplate(ctx, to, tmpl, data); err != nil {
			s.log.Warn("failed to send course request email", zap.Error(err), zap.String("template", string(tmpl)))
		}
	}
}

// uploadCourseFile stores file under the course and records its metadata. A nil file is skipped.
func (s *service) uploadCourseFile(ctx context.Context, courseID int64, userID string, file *entities.File, purpose string, public bool) error {
	if file == nil {
		return nil
	}
	courseIDStr := fmt.Sprintf("%d", courseID)
	file.OwnerID = courseIDStr
	file.OwnerType = entities.OwnerTypeCourse
	file.Key = fmt.Sprintf("files/courses/%d/%s", courseID, file.Name)
	file.Purpose = purpose
	file.Public = public
	file.UploadedBy = userID
	file.CreatedAt = time.Now().Format(time.RFC3339)

	if err := s.storage.UploadFile(ctx, []*entities.File{file}); err != nil {
		s.log.Error("failed to upload course file", zap.Error(err), zap.String("course_id", courseIDStr), zap.String("purpose", purpose))
		return err
	}

	if err := s.repo.SaveFilesToDB(ctx, []*entities.File{file}); err != nil {
		s.log.Error("failed to save course file metadata", zap.Error(err), zap.String("course_id", courseIDStr), zap.String("purpose", purpose))
		return err
	}

	s.log.Info("course file uploaded successfully", zap.String("course_id", courseIDStr), zap.String("purpose", purpose))
	return nil
}

func (s *service) UpdateCourse(
	ctx context.Context,
	courseID string,
	course entities.Course,
) error {
	return s.repo.UpdateCourse(ctx, courseID, course)
}

func (s *service) DeleteCourse(ctx context.Context, courseID string) error {
	return s.repo.DeleteCourse(ctx, courseID)
}
