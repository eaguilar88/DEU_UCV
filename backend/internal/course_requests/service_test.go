package course_requests

import (
	"context"
	"errors"
	"testing"

	"github.com/eaguilar88/deu/internal/course_requests/mocks"
	"github.com/eaguilar88/deu/internal/email"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestCourseRequestService_GetMyCourseRequests(t *testing.T) {
	type testCase struct {
		name    string
		prepare func(repoMock *mocks.MockRepository)
		want    []entities.CourseRequest
		wantErr error
	}

	tests := []testCase{
		{
			name: "success",
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetProviderByUserID(mock.Anything, "user-1").
					Return(entities.Provider{ID: "provider-1"}, nil)
				repoMock.EXPECT().GetCourseRequestsByProvider(mock.Anything, "provider-1", mock.AnythingOfType("entities.PageScope")).
					Return([]entities.CourseRequest{{ID: "1"}}, entities.PageScope{Page: 1, PerPage: 10}, nil)
			},
			want: []entities.CourseRequest{{ID: "1"}},
		},
		{
			name: "no provider record returns empty list, not an error",
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetProviderByUserID(mock.Anything, "user-1").
					Return(entities.Provider{}, providers.ErrProviderNotFound)
			},
			want: nil,
		},
		{
			name: "error resolving provider",
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetProviderByUserID(mock.Anything, "user-1").
					Return(entities.Provider{}, errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
		{
			name: "error listing requests by provider",
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetProviderByUserID(mock.Anything, "user-1").
					Return(entities.Provider{ID: "provider-1"}, nil)
				repoMock.EXPECT().GetCourseRequestsByProvider(mock.Anything, "provider-1", mock.AnythingOfType("entities.PageScope")).
					Return(nil, entities.PageScope{}, errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			if tt.prepare != nil {
				tt.prepare(repoMock)
			}
			s := NewService(repoMock, mocks.NewMockStorageClient(t), mocks.NewMockMailClient(t), zap.NewNop())
			got, _, err := s.GetMyCourseRequests(ctx, "user-1", entities.PageScope{})
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

// pendingCourseRequest is an under-review request for course "Curso de prueba", owned by provider "provider-1".
func pendingCourseRequest() entities.CourseRequest {
	return entities.CourseRequest{
		ID:     "1",
		Status: entities.RequestStatus_UNDER_REVIEW,
		Course: &entities.Course{
			ID:            "3",
			Name:          "Curso de prueba",
			Owner:         entities.User{ID: "provider-1"},
			Faculty:       entities.FacultyCiencias,
			OriginFaculty: entities.FacultyCiencias,
		},
	}
}

// facultyAdmin is a faculty_admin of the given faculty.
func facultyAdmin(faculty entities.Faculty) entities.Viewer {
	return entities.Viewer{UserID: "reviewer-1", Roles: []string{"faculty_admin"}, Faculty: faculty}
}

func expectProviderLookup(repoMock *mocks.MockRepository) {
	repoMock.EXPECT().GetProvider(mock.Anything, "provider-1").
		Return(entities.Provider{ID: "provider-1", User: entities.User{Email: "provider@test.com"}}, nil)
}

// reviewAction describes one of the three admin actions under test, so the shared cases (already
// processed, provider lookup failure, email failure) run against each of them.
type reviewAction struct {
	name        string
	template    email.Template
	data        any
	expectWrite func(repoMock *mocks.MockRepository) *mock.Call
	run         func(s Service, viewer entities.Viewer) error
}

func courseReviewActions() []reviewAction {
	return []reviewAction{
		{
			name:     "approve",
			template: email.TemplateCourseRequestApproved,
			data:     email.CourseRequestApprovedData{CourseName: "Curso de prueba", Comments: "buen curso"},
			expectWrite: func(repoMock *mocks.MockRepository) *mock.Call {
				req := entities.CourseRequest{ID: "1", Reviewer: entities.User{ID: "reviewer-1"}, Comments: "buen curso"}
				return repoMock.EXPECT().ApproveCourseRequest(mock.Anything, req, entities.CourseType_SkillDevelopment.String(), mock.Anything).
					RunAndReturn(func(_ context.Context, _ entities.CourseRequest, _ string, notify func() error) error {
						return notify()
					}).Call
			},
			run: func(s Service, viewer entities.Viewer) error {
				req := entities.CourseRequest{ID: "1", Reviewer: entities.User{ID: "reviewer-1"}, Comments: "buen curso"}
				return s.ApproveCourseRequest(context.Background(), req, entities.CourseType_SkillDevelopment, viewer)
			},
		},
		{
			name:     "reject",
			template: email.TemplateCourseRequestRejected,
			data:     email.CourseRequestRejectedData{CourseName: "Curso de prueba", Reason: "no cumple"},
			expectWrite: func(repoMock *mocks.MockRepository) *mock.Call {
				return repoMock.EXPECT().RejectCourseRequest(mock.Anything, "1", "reviewer-1", "no cumple", mock.Anything).
					RunAndReturn(func(_ context.Context, _, _, _ string, notify func() error) error { return notify() }).Call
			},
			run: func(s Service, viewer entities.Viewer) error {
				return s.RejectCourseRequest(context.Background(), "1", "reviewer-1", "no cumple", viewer)
			},
		},
		{
			name:     "redirect",
			template: email.TemplateCourseRequestRedirected,
			data: email.CourseRequestRedirectedData{
				CourseName: "Curso de prueba",
				Faculty:    entities.FacultyCiencias.String(),
				Reason:     "otra facultad",
			},
			expectWrite: func(repoMock *mocks.MockRepository) *mock.Call {
				return repoMock.EXPECT().RedirectCourseRequest(mock.Anything, "1", "reviewer-1", entities.FacultyCiencias, "otra facultad", mock.Anything).
					RunAndReturn(func(_ context.Context, _, _ string, _ entities.Faculty, _ string, notify func() error) error {
						return notify()
					}).Call
			},
			run: func(s Service, viewer entities.Viewer) error {
				return s.RedirectCourseRequest(context.Background(), "1", "reviewer-1", entities.FacultyCiencias, "otra facultad", viewer)
			},
		},
	}
}

func TestCourseRequestService_ReviewNotifications(t *testing.T) {
	type testCase struct {
		name    string
		viewer  *entities.Viewer
		prepare func(action reviewAction, repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient)
		wantErr error
	}

	deuAdmin := entities.Viewer{UserID: "reviewer-1", Roles: []string{"deu_admin"}}
	otherFacultyAdmin := facultyAdmin(entities.FacultyIngenieria)

	tests := []testCase{
		{
			name: "success notifies the provider inside the transaction",
			prepare: func(action reviewAction, repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetCourseRequestByID(mock.Anything, "1").Return(pendingCourseRequest(), nil)
				expectProviderLookup(repoMock)
				action.expectWrite(repoMock)
				mailMock.EXPECT().SendTemplate(mock.Anything, "provider@test.com", action.template, action.data).
					Return(nil)
			},
		},
		{
			name:   "a global admin may review any faculty's request",
			viewer: &deuAdmin,
			prepare: func(action reviewAction, repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetCourseRequestByID(mock.Anything, "1").Return(pendingCourseRequest(), nil)
				expectProviderLookup(repoMock)
				action.expectWrite(repoMock)
				mailMock.EXPECT().SendTemplate(mock.Anything, "provider@test.com", action.template, action.data).
					Return(nil)
			},
		},
		{
			name:   "an admin of another faculty is forbidden",
			viewer: &otherFacultyAdmin,
			prepare: func(action reviewAction, repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetCourseRequestByID(mock.Anything, "1").Return(pendingCourseRequest(), nil)
			},
			wantErr: ErrCourseRequestForbidden,
		},
		{
			name:   "the faculty a request was redirected from is forbidden",
			viewer: &otherFacultyAdmin,
			prepare: func(action reviewAction, repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				req := pendingCourseRequest()
				req.Course.OriginFaculty = entities.FacultyIngenieria
				repoMock.EXPECT().GetCourseRequestByID(mock.Anything, "1").Return(req, nil)
			},
			wantErr: ErrCourseRequestForbidden,
		},
		{
			name: "error request already processed",
			prepare: func(action reviewAction, repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				req := pendingCourseRequest()
				req.Status = entities.RequestStatus_APPROVED
				repoMock.EXPECT().GetCourseRequestByID(mock.Anything, "1").Return(req, nil)
			},
			wantErr: ErrRequestIsProcessed,
		},
		{
			name: "error getting course request",
			prepare: func(action reviewAction, repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetCourseRequestByID(mock.Anything, "1").
					Return(entities.CourseRequest{}, ErrCourseRequestNotFound)
			},
			wantErr: ErrCourseRequestNotFound,
		},
		{
			name: "error getting provider does not touch the request",
			prepare: func(action reviewAction, repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetCourseRequestByID(mock.Anything, "1").Return(pendingCourseRequest(), nil)
				repoMock.EXPECT().GetProvider(mock.Anything, "provider-1").
					Return(entities.Provider{}, errors.New("provider not found"))
			},
			wantErr: errors.New("provider not found"),
		},
		{
			name: "error sending email rolls back the action",
			prepare: func(action reviewAction, repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetCourseRequestByID(mock.Anything, "1").Return(pendingCourseRequest(), nil)
				expectProviderLookup(repoMock)
				action.expectWrite(repoMock)
				mailMock.EXPECT().SendTemplate(mock.Anything, "provider@test.com", action.template, action.data).
					Return(errors.New("smtp down"))
			},
			wantErr: errors.New("error sending course request notification email: smtp down"),
		},
	}

	for _, action := range courseReviewActions() {
		for _, tt := range tests {
			t.Run(action.name+"/"+tt.name, func(t *testing.T) {
				repoMock := mocks.NewMockRepository(t)
				mailMock := mocks.NewMockMailClient(t)
				tt.prepare(action, repoMock, mailMock)

				s := NewService(repoMock, mocks.NewMockStorageClient(t), mailMock, zap.NewNop())
				viewer := facultyAdmin(entities.FacultyCiencias)
				if tt.viewer != nil {
					viewer = *tt.viewer
				}
				err := action.run(s, viewer)
				if tt.wantErr != nil {
					assert.EqualError(t, err, tt.wantErr.Error())
				} else {
					assert.NoError(t, err)
				}
			})
		}
	}
}

func TestCourseRequestService_ApproveWithEvaluation(t *testing.T) {
	score := 18.5
	newRequest := func() entities.CourseRequest {
		return entities.CourseRequest{
			ID:             "1",
			Reviewer:       entities.User{ID: "reviewer-1"},
			Score:          &score,
			Classification: "Formación para el trabajo",
			EvaluationFile: &entities.File{Name: "archivo_evaluacion.pdf"},
		}
	}
	evaluationKey := "files/course-requests/1/archivo_evaluacion.pdf"

	t.Run("uploads the private evaluation file before approving", func(t *testing.T) {
		repoMock := mocks.NewMockRepository(t)
		storageMock := mocks.NewMockStorageClient(t)
		mailMock := mocks.NewMockMailClient(t)
		repoMock.EXPECT().GetCourseRequestByID(mock.Anything, "1").Return(pendingCourseRequest(), nil)
		expectProviderLookup(repoMock)
		storageMock.EXPECT().UploadFile(mock.Anything, mock.MatchedBy(func(files []*entities.File) bool {
			f := files[0]
			return len(files) == 1 && f.Key == evaluationKey && !f.Public &&
				f.OwnerType == entities.OwnerTypeCourseRequest && f.OwnerID == "1" &&
				f.Purpose == entities.CourseRequestFileTypeEvaluation && f.UploadedBy == "reviewer-1"
		})).Return(nil)
		repoMock.EXPECT().ApproveCourseRequest(mock.Anything, mock.MatchedBy(func(r entities.CourseRequest) bool {
			return *r.Score == score && r.Classification == "Formación para el trabajo" && r.EvaluationFile.Key == evaluationKey
		}), entities.CourseType_LifeSkills.String(), mock.Anything).Return(nil)

		s := NewService(repoMock, storageMock, mailMock, zap.NewNop())
		err := s.ApproveCourseRequest(context.Background(), newRequest(), entities.CourseType_LifeSkills, facultyAdmin(entities.FacultyCiencias))

		assert.NoError(t, err)
	})

	t.Run("failed approval removes the uploaded file", func(t *testing.T) {
		repoMock := mocks.NewMockRepository(t)
		storageMock := mocks.NewMockStorageClient(t)
		repoMock.EXPECT().GetCourseRequestByID(mock.Anything, "1").Return(pendingCourseRequest(), nil)
		expectProviderLookup(repoMock)
		storageMock.EXPECT().UploadFile(mock.Anything, mock.Anything).Return(nil)
		repoMock.EXPECT().ApproveCourseRequest(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error"))
		storageMock.EXPECT().DeleteFile(mock.Anything, evaluationKey).Return(nil)

		s := NewService(repoMock, storageMock, mocks.NewMockMailClient(t), zap.NewNop())
		err := s.ApproveCourseRequest(context.Background(), newRequest(), entities.CourseType_LifeSkills, facultyAdmin(entities.FacultyCiencias))

		assert.EqualError(t, err, "db error")
	})

	t.Run("upload failure does not approve", func(t *testing.T) {
		repoMock := mocks.NewMockRepository(t)
		storageMock := mocks.NewMockStorageClient(t)
		repoMock.EXPECT().GetCourseRequestByID(mock.Anything, "1").Return(pendingCourseRequest(), nil)
		expectProviderLookup(repoMock)
		storageMock.EXPECT().UploadFile(mock.Anything, mock.Anything).Return(errors.New("b2 down"))

		s := NewService(repoMock, storageMock, mocks.NewMockMailClient(t), zap.NewNop())
		err := s.ApproveCourseRequest(context.Background(), newRequest(), entities.CourseType_LifeSkills, facultyAdmin(entities.FacultyCiencias))

		assert.EqualError(t, err, "b2 down")
	})
}

func TestCourseRequestService_GetCourseRequestByID(t *testing.T) {
	repoMock := mocks.NewMockRepository(t)
	storageMock := mocks.NewMockStorageClient(t)
	repoMock.EXPECT().GetCourseRequestByID(mock.Anything, "1").Return(pendingCourseRequest(), nil)
	// The course under review is inactive, so it is loaded including inactive ones.
	repoMock.EXPECT().GetCourseIncludingInactive(mock.Anything, "3").
		Return(entities.Course{ID: "3", Name: "Curso de prueba", Bibliography: "Libro"}, nil)
	repoMock.EXPECT().GetFilesByOwner(mock.Anything, "3", entities.OwnerTypeCourse).
		Return(entities.GroupedFiles{
			entities.CourseFileTypeCover:         {{Key: "cover-key", Public: true}},
			entities.CourseFileTypeFacilitatorCV: {{Key: "cv-key"}},
		}, nil)
	storageMock.EXPECT().GetFileURL(mock.Anything, "cover-key").Return("https://extension.ucv.ve/files/cover-key", nil)
	storageMock.EXPECT().GetPresignedFileURL(mock.Anything, "cv-key").Return("https://b2/cv", nil)
	repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1", entities.OwnerTypeCourseRequest).
		Return(entities.GroupedFiles{entities.CourseRequestFileTypeEvaluation: {{Key: "eval-key"}}}, nil)
	storageMock.EXPECT().GetPresignedFileURL(mock.Anything, "eval-key").Return("https://b2/eval", nil)

	s := NewService(repoMock, storageMock, mocks.NewMockMailClient(t), zap.NewNop())
	got, err := s.GetCourseRequestByID(context.Background(), "1")

	assert.NoError(t, err)
	assert.Equal(t, "Libro", got.Course.Bibliography)
	assert.Equal(t, "https://extension.ucv.ve/files/cover-key", got.Course.Cover.URL)
	assert.Equal(t, "https://b2/cv", got.Course.FacilitatorCV.URL)
	assert.Equal(t, "https://b2/eval", got.EvaluationFile.URL)
}

func TestCourseRequestService_GetMyCourseRequests_AttachesCovers(t *testing.T) {
	repoMock := mocks.NewMockRepository(t)
	storageMock := mocks.NewMockStorageClient(t)
	repoMock.EXPECT().GetProviderByUserID(mock.Anything, "user-1").Return(entities.Provider{ID: "provider-1"}, nil)
	repoMock.EXPECT().GetCourseRequestsByProvider(mock.Anything, "provider-1", mock.Anything).
		Return([]entities.CourseRequest{
			{ID: "1", Course: &entities.Course{ID: "3"}},
			{ID: "2", Course: &entities.Course{ID: "4"}},
		}, entities.PageScope{}, nil)
	repoMock.EXPECT().GetFilesByOwner(mock.Anything, "3", entities.OwnerTypeCourse).
		Return(entities.GroupedFiles{entities.CourseFileTypeCover: {{Key: "cover-3", Public: true}}}, nil)
	// A failing lookup leaves that course without a cover instead of failing the list.
	repoMock.EXPECT().GetFilesByOwner(mock.Anything, "4", entities.OwnerTypeCourse).Return(nil, errors.New("db error"))
	storageMock.EXPECT().GetFileURL(mock.Anything, "cover-3").Return("https://extension.ucv.ve/files/cover-3", nil)

	s := NewService(repoMock, storageMock, mocks.NewMockMailClient(t), zap.NewNop())
	got, _, err := s.GetMyCourseRequests(context.Background(), "user-1", entities.PageScope{})

	assert.NoError(t, err)
	assert.Equal(t, "https://extension.ucv.ve/files/cover-3", got[0].Course.Cover.URL)
	assert.Nil(t, got[1].Course.Cover)
}
