package courses

import (
	"context"
	"errors"
	"testing"

	"github.com/eaguilar88/deu/internal/courses/mocks"
	"github.com/eaguilar88/deu/internal/email"
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
		prepare func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient)
		want    int64
		wantErr bool
	}

	tests := []testCase{
		{
			name: "faculty and origin faculty are inherited from the provider",
			// A faculty set by the caller is overwritten by the provider's.
			course: entities.Course{Name: "Test Course", Faculty: entities.FacultyIngenieria},
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetProviderByUserID(mock.Anything, userID).
					Return(entities.Provider{ID: "p-1", Faculty: entities.FacultyCiencias}, nil)
				repoMock.EXPECT().CreateCourseWithRequest(mock.Anything, mock.MatchedBy(func(c entities.Course) bool {
					return c.Owner.ID == "p-1" &&
						c.Faculty == entities.FacultyCiencias &&
						c.OriginFaculty == entities.FacultyCiencias
				})).Return(int64(10), int64(20), nil)
				repoMock.EXPECT().GetFacultyCoordinatorEmails(mock.Anything, mock.Anything).Return(nil, nil)
				repoMock.EXPECT().GetDEUAdminEmails(mock.Anything).Return(nil, nil)
			},
			want: 10,
		},
		{
			name:   "provider without faculty falls back to DEU",
			course: entities.Course{Name: "Test Course"},
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetProviderByUserID(mock.Anything, userID).
					Return(entities.Provider{ID: "p-1"}, nil)
				repoMock.EXPECT().CreateCourseWithRequest(mock.Anything, mock.MatchedBy(func(c entities.Course) bool {
					return c.Faculty == entities.FacultyDEU && c.OriginFaculty == entities.FacultyDEU
				})).Return(int64(10), int64(20), nil)
				repoMock.EXPECT().GetFacultyCoordinatorEmails(mock.Anything, mock.Anything).Return(nil, nil)
				repoMock.EXPECT().GetDEUAdminEmails(mock.Anything).Return(nil, nil)
			},
			want: 10,
		},
		{
			name:   "notifies the faculty coordinators and the DEU admins",
			course: entities.Course{Name: "Test Course"},
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetProviderByUserID(mock.Anything, userID).
					Return(entities.Provider{ID: "p-1", Name: "ACME", Faculty: entities.FacultyCiencias}, nil)
				repoMock.EXPECT().CreateCourseWithRequest(mock.Anything, mock.Anything).Return(int64(10), int64(20), nil)
				repoMock.EXPECT().GetFacultyCoordinatorEmails(mock.Anything, entities.FacultyCiencias).
					Return([]string{"coord1@example.com", "coord2@example.com"}, nil)
				repoMock.EXPECT().GetDEUAdminEmails(mock.Anything).Return([]string{"deu@example.com"}, nil)
				data := email.CourseRequestSubmittedData{CourseName: "Test Course", ProviderName: "ACME", Faculty: string(entities.FacultyCiencias)}
				mailMock.EXPECT().SendTemplate(mock.Anything, "coord1@example.com", email.TemplateCourseRequestSubmittedFaculty, data).Return(nil)
				mailMock.EXPECT().SendTemplate(mock.Anything, "coord2@example.com", email.TemplateCourseRequestSubmittedFaculty, data).Return(nil)
				mailMock.EXPECT().SendTemplate(mock.Anything, "deu@example.com", email.TemplateCourseRequestSubmittedDEU, data).Return(nil)
			},
			want: 10,
		},
		{
			name:   "email and recipient lookup failures do not fail the creation",
			course: entities.Course{Name: "Test Course"},
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetProviderByUserID(mock.Anything, userID).
					Return(entities.Provider{ID: "p-1", Faculty: entities.FacultyCiencias}, nil)
				repoMock.EXPECT().CreateCourseWithRequest(mock.Anything, mock.Anything).Return(int64(10), int64(20), nil)
				repoMock.EXPECT().GetFacultyCoordinatorEmails(mock.Anything, mock.Anything).Return([]string{"coord@example.com"}, nil)
				repoMock.EXPECT().GetDEUAdminEmails(mock.Anything).Return(nil, errors.New("db error"))
				mailMock.EXPECT().SendTemplate(mock.Anything, "coord@example.com", email.TemplateCourseRequestSubmittedFaculty, mock.Anything).
					Return(errors.New("smtp down"))
			},
			want: 10,
		},
		{
			name:   "error getting provider",
			course: entities.Course{Name: "Test Course"},
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
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
			mailMock := mocks.NewMockMailClient(t)
			if tt.prepare != nil {
				tt.prepare(repoMock, mailMock)
			}
			s := NewService(repoMock, storageMock, mailMock, zap.NewNop())
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

			s := NewService(repoMock, storageMock, mocks.NewMockMailClient(t), zap.NewNop())
			_, _, err := s.GetCourses(context.Background(), tt.filter, tt.viewer, entities.PageScope{})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_CreateCourse_UploadsCoverAndFacilitatorCV(t *testing.T) {
	repoMock := mocks.NewMockRepository(t)
	storageMock := mocks.NewMockStorageClient(t)
	repoMock.EXPECT().GetProviderByUserID(mock.Anything, "user-1").Return(entities.Provider{ID: "p-1"}, nil)
	repoMock.EXPECT().CreateCourseWithRequest(mock.Anything, mock.Anything).Return(int64(10), int64(20), nil)
	repoMock.EXPECT().GetFacultyCoordinatorEmails(mock.Anything, mock.Anything).Return(nil, nil)
	repoMock.EXPECT().GetDEUAdminEmails(mock.Anything).Return(nil, nil)

	var uploaded []*entities.File
	storageMock.EXPECT().UploadFile(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, files []*entities.File) error {
			uploaded = append(uploaded, files...)
			return nil
		}).Times(2)
	repoMock.EXPECT().SaveFilesToDB(mock.Anything, mock.Anything).Return(nil).Times(2)

	course := entities.Course{
		Name:          "Fotografía",
		Cover:         &entities.File{Name: "portada.png"},
		FacilitatorCV: &entities.File{Name: "cv_facilitador.pdf"},
	}
	s := NewService(repoMock, storageMock, mocks.NewMockMailClient(t), zap.NewNop())
	_, err := s.CreateCourse(context.Background(), "user-1", course)

	assert.NoError(t, err)
	if assert.Len(t, uploaded, 2) {
		assert.Equal(t, "files/courses/10/portada.png", uploaded[0].Key)
		assert.True(t, uploaded[0].Public, "the cover is public")
		assert.Equal(t, "files/courses/10/cv_facilitador.pdf", uploaded[1].Key)
		assert.Equal(t, entities.CourseFileTypeFacilitatorCV, uploaded[1].Purpose)
		assert.False(t, uploaded[1].Public, "the CV is private")
	}
}

func TestService_GetCourse(t *testing.T) {
	courseFiles := entities.GroupedFiles{
		entities.CourseFileTypeCover:         {{Key: "cover-key", Public: true}},
		entities.CourseFileTypeFacilitatorCV: {{Key: "cv-key"}},
	}

	t.Run("attaches cover, private CV and provider summary", func(t *testing.T) {
		repoMock := mocks.NewMockRepository(t)
		storageMock := mocks.NewMockStorageClient(t)
		repoMock.EXPECT().GetCourse(mock.Anything, "10").Return(entities.Course{ID: "10", Owner: entities.User{ID: "p-1"}}, nil)
		repoMock.EXPECT().GetFilesByOwner(mock.Anything, "10", entities.OwnerTypeCourse).Return(courseFiles, nil)
		storageMock.EXPECT().GetFileURL(mock.Anything, "cover-key").Return("https://extension.ucv.ve/files/cover-key", nil)
		storageMock.EXPECT().GetPresignedFileURL(mock.Anything, "cv-key").Return("https://b2/cv?signed", nil)
		repoMock.EXPECT().GetProvider(mock.Anything, "p-1").
			Return(entities.Provider{ID: "p-1", User: entities.User{FirstName: "Ana", LastName: "Pérez"}}, nil)
		// The private CI must never be exposed: only a public logo is.
		repoMock.EXPECT().GetFilesByOwner(mock.Anything, "p-1", entities.OwnerTypeProvider).
			Return(entities.GroupedFiles{
				entities.ProviderFileTypeLogo: {{Key: "logo-key", Public: true}},
				entities.ProviderFileTypeCI:   {{Key: "ci-key"}},
			}, nil)
		storageMock.EXPECT().GetFileURL(mock.Anything, "logo-key").Return("https://extension.ucv.ve/files/logo-key", nil)

		s := NewService(repoMock, storageMock, mocks.NewMockMailClient(t), zap.NewNop())
		got, err := s.GetCourse(context.Background(), "10", entities.Viewer{})

		assert.NoError(t, err)
		assert.Equal(t, "https://extension.ucv.ve/files/cover-key", got.Cover.URL)
		assert.Equal(t, "https://b2/cv?signed", got.FacilitatorCV.URL)
		assert.Equal(t, &entities.CourseProviderSummary{Name: "Ana Pérez", LogoURL: "https://extension.ucv.ve/files/logo-key"}, got.Provider)
	})

	t.Run("provider lookup failure still returns the course", func(t *testing.T) {
		repoMock := mocks.NewMockRepository(t)
		storageMock := mocks.NewMockStorageClient(t)
		repoMock.EXPECT().GetCourse(mock.Anything, "10").Return(entities.Course{ID: "10", Owner: entities.User{ID: "p-1"}}, nil)
		repoMock.EXPECT().GetFilesByOwner(mock.Anything, "10", entities.OwnerTypeCourse).Return(entities.GroupedFiles{}, nil)
		repoMock.EXPECT().GetProvider(mock.Anything, "p-1").Return(entities.Provider{}, errors.New("db error"))

		s := NewService(repoMock, storageMock, mocks.NewMockMailClient(t), zap.NewNop())
		got, err := s.GetCourse(context.Background(), "10", entities.Viewer{})

		assert.NoError(t, err)
		assert.Equal(t, "10", got.ID)
		assert.Nil(t, got.Provider)
	})

	t.Run("CV URL error fails", func(t *testing.T) {
		repoMock := mocks.NewMockRepository(t)
		storageMock := mocks.NewMockStorageClient(t)
		repoMock.EXPECT().GetCourse(mock.Anything, "10").Return(entities.Course{ID: "10"}, nil)
		repoMock.EXPECT().GetFilesByOwner(mock.Anything, "10", entities.OwnerTypeCourse).
			Return(entities.GroupedFiles{entities.CourseFileTypeFacilitatorCV: {{Key: "cv-key"}}}, nil)
		storageMock.EXPECT().GetPresignedFileURL(mock.Anything, "cv-key").Return("", errors.New("b2 down"))

		s := NewService(repoMock, storageMock, mocks.NewMockMailClient(t), zap.NewNop())
		_, err := s.GetCourse(context.Background(), "10", entities.Viewer{})

		assert.EqualError(t, err, "b2 down")
	})
}

func TestService_GetCourse_Unapproved(t *testing.T) {
	pending := entities.Course{ID: "10", Owner: entities.User{ID: "p-1"}, Faculty: entities.FacultyFarmacia, OriginFaculty: entities.FacultyMedicina}
	provider := entities.Provider{ID: "p-1", User: entities.User{ID: "owner-1"}}
	deuAdmin := entities.Viewer{UserID: "1", Roles: []string{entities.RoleNameFromID(entities.RoleDeuAdmin)}, Faculty: entities.FacultyDEU}
	facultyAdmin := func(f entities.Faculty) entities.Viewer {
		return entities.Viewer{UserID: "2", Roles: []string{entities.RoleNameFromID(entities.RoleFacultyAdmin)}, Faculty: f}
	}

	tests := []struct {
		name          string
		viewer        entities.Viewer
		checksOwner   bool
		wantNotFound  bool
		skipsFallback bool
	}{
		{name: "anonymous never sees it", viewer: entities.Viewer{}, wantNotFound: true, skipsFallback: true},
		{name: "owner sees it", viewer: entities.Viewer{UserID: "owner-1"}, checksOwner: true},
		{name: "another user does not", viewer: entities.Viewer{UserID: "intruder"}, checksOwner: true, wantNotFound: true},
		{name: "deu_admin sees it", viewer: deuAdmin},
		{name: "admin of the reviewing faculty sees it", viewer: facultyAdmin(entities.FacultyFarmacia)},
		{name: "admin of the origin faculty sees it", viewer: facultyAdmin(entities.FacultyMedicina)},
		{name: "admin of another faculty does not", viewer: facultyAdmin(entities.FacultyIngenieria), checksOwner: true, wantNotFound: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			storageMock := mocks.NewMockStorageClient(t)
			repoMock.EXPECT().GetCourse(mock.Anything, "10").Return(entities.Course{}, ErrCourseNotFound)
			if !tt.skipsFallback {
				repoMock.EXPECT().GetCourseIncludingInactive(mock.Anything, "10").Return(pending, nil)
			}
			if tt.checksOwner {
				repoMock.EXPECT().GetProvider(mock.Anything, "p-1").Return(provider, nil).Once()
			}
			if !tt.wantNotFound {
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, mock.Anything, mock.Anything).Return(entities.GroupedFiles{}, nil)
				repoMock.EXPECT().GetProvider(mock.Anything, "p-1").Return(provider, nil)
			}

			s := NewService(repoMock, storageMock, mocks.NewMockMailClient(t), zap.NewNop())
			got, err := s.GetCourse(context.Background(), "10", tt.viewer)

			if tt.wantNotFound {
				assert.ErrorIs(t, err, ErrCourseNotFound)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, "10", got.ID)
			assert.Equal(t, "owner-1", got.Provider.UserID)
		})
	}
}
