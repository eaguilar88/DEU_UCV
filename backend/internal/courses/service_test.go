package courses

import (
	"context"
	"errors"
	"testing"

	"github.com/eaguilar88/deu/internal/courses/mocks"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestService_CreateCourse(t *testing.T) {
	userID := "user-1"

	type testCase struct {
		name    string
		course  entities.Course
		prepare func(repoMock *mocks.MockRepository)
		want    int64
		wantErr bool
	}

	tests := []testCase{
		{
			name: "faculty and origin faculty are inherited from the provider",
			// A faculty set by the caller is overwritten by the provider's.
			course: entities.Course{Name: "Test Course", Faculty: entities.FacultyIngenieria},
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetProviderByUserID(mock.Anything, userID).
					Return(entities.Provider{ID: "p-1", Faculty: entities.FacultyCiencias}, nil)
				repoMock.EXPECT().CreateCourseWithRequest(mock.Anything, mock.MatchedBy(func(c entities.Course) bool {
					return c.Owner.ID == "p-1" &&
						c.Faculty == entities.FacultyCiencias &&
						c.OriginFaculty == entities.FacultyCiencias
				})).Return(int64(10), int64(20), nil)
			},
			want: 10,
		},
		{
			name:   "provider without faculty falls back to DEU",
			course: entities.Course{Name: "Test Course"},
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetProviderByUserID(mock.Anything, userID).
					Return(entities.Provider{ID: "p-1"}, nil)
				repoMock.EXPECT().CreateCourseWithRequest(mock.Anything, mock.MatchedBy(func(c entities.Course) bool {
					return c.Faculty == entities.FacultyDEU && c.OriginFaculty == entities.FacultyDEU
				})).Return(int64(10), int64(20), nil)
			},
			want: 10,
		},
		{
			name:   "error getting provider",
			course: entities.Course{Name: "Test Course"},
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetProviderByUserID(mock.Anything, userID).
					Return(entities.Provider{}, errors.New("db error"))
			},
			want:    -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			storageMock := mocks.NewMockStorageClient(t)
			if tt.prepare != nil {
				tt.prepare(repoMock)
			}
			s := NewService(repoMock, storageMock, zap.NewNop())
			got, err := s.CreateCourse(context.Background(), userID, tt.course)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_GetCourses(t *testing.T) {
	openOrClosed := []entities.CourseManagementStatus{
		entities.CourseManagementStatusOpen,
		entities.CourseManagementStatusClosed,
	}

	type testCase struct {
		name        string
		filter      entities.CourseFilter
		viewer      entities.Viewer
		wantVisible []entities.CourseManagementStatus
		wantRepoErr error
		wantErr     bool
	}

	tests := []testCase{
		{
			name:        "anonymous only sees open or closed courses",
			viewer:      entities.Viewer{},
			wantVisible: openOrClosed,
		},
		{
			name:        "anonymous asking for a user's courses is still restricted",
			filter:      entities.CourseFilter{OwnerUserID: "7"},
			viewer:      entities.Viewer{},
			wantVisible: openOrClosed,
		},
		{
			name:        "logged-in user asking for someone else's courses is restricted",
			filter:      entities.CourseFilter{OwnerUserID: "7"},
			viewer:      entities.Viewer{UserID: "8", Roles: []string{"provider"}},
			wantVisible: openOrClosed,
		},
		{
			name:        "provider filtering by code without usuario_id is restricted",
			filter:      entities.CourseFilter{ProviderCode: "ECP-abc123"},
			viewer:      entities.Viewer{UserID: "7", Roles: []string{"provider"}},
			wantVisible: openOrClosed,
		},
		{
			name:   "owner listing their own courses sees every status",
			filter: entities.CourseFilter{OwnerUserID: "7"},
			viewer: entities.Viewer{UserID: "7", Roles: []string{"provider"}},
		},
		{
			name:   "deu_admin sees every status",
			viewer: entities.Viewer{UserID: "1", Roles: []string{entities.RoleNameFromID(entities.RoleDeuAdmin)}},
		},
		{
			name:   "root sees every status",
			viewer: entities.Viewer{UserID: "1", Roles: []string{entities.RoleNameFromID(entities.RoleRoot)}},
		},
		{
			name:   "faculty_admin sees every status",
			viewer: entities.Viewer{UserID: "2", Roles: []string{entities.RoleNameFromID(entities.RoleFacultyAdmin)}, Faculty: entities.FacultyCiencias},
		},
		{
			name:        "client-provided visible statuses are ignored",
			filter:      entities.CourseFilter{VisibleStatuses: []entities.CourseManagementStatus{entities.CourseManagementStatusClosureRequested}},
			viewer:      entities.Viewer{},
			wantVisible: openOrClosed,
		},
		{
			name:        "repository error",
			viewer:      entities.Viewer{},
			wantVisible: openOrClosed,
			wantRepoErr: errors.New("db error"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			storageMock := mocks.NewMockStorageClient(t)

			wantFilter := tt.filter
			wantFilter.VisibleStatuses = tt.wantVisible
			repoMock.EXPECT().GetCourses(mock.Anything, wantFilter, entities.PageScope{}).
				Return([]entities.Course{}, entities.PageScope{}, tt.wantRepoErr)

			s := NewService(repoMock, storageMock, zap.NewNop())
			_, _, err := s.GetCourses(context.Background(), tt.filter, tt.viewer, entities.PageScope{})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
