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
			s := NewService(repoMock, mocks.NewMockMailClient(t), zap.NewNop())
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
		Course: &entities.Course{ID: "3", Name: "Curso de prueba", Owner: entities.User{ID: "provider-1"}},
	}
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
	run         func(s Service) error
}

func courseReviewActions() []reviewAction {
	return []reviewAction{
		{
			name:     "approve",
			template: email.TemplateCourseRequestApproved,
			data:     email.CourseRequestApprovedData{CourseName: "Curso de prueba", Comments: "buen curso"},
			expectWrite: func(repoMock *mocks.MockRepository) *mock.Call {
				return repoMock.EXPECT().ApproveCourseRequest(mock.Anything, "1", "reviewer-1", entities.CourseType_SkillDevelopment.String(), "buen curso", mock.Anything).
					RunAndReturn(func(_ context.Context, _, _, _, _ string, notify func() error) error { return notify() }).Call
			},
			run: func(s Service) error {
				req := entities.CourseRequest{ID: "1", Reviewer: entities.User{ID: "reviewer-1"}, Comments: "buen curso"}
				return s.ApproveCourseRequest(context.Background(), req, entities.CourseType_SkillDevelopment)
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
			run: func(s Service) error {
				return s.RejectCourseRequest(context.Background(), "1", "reviewer-1", "no cumple")
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
			run: func(s Service) error {
				return s.RedirectCourseRequest(context.Background(), "1", "reviewer-1", entities.FacultyCiencias, "otra facultad")
			},
		},
	}
}

func TestCourseRequestService_ReviewNotifications(t *testing.T) {
	type testCase struct {
		name    string
		prepare func(action reviewAction, repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient)
		wantErr error
	}

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

				s := NewService(repoMock, mailMock, zap.NewNop())
				err := action.run(s)
				if tt.wantErr != nil {
					assert.EqualError(t, err, tt.wantErr.Error())
				} else {
					assert.NoError(t, err)
				}
			})
		}
	}
}
