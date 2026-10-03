package groups

import (
	"context"
	"errors"
	"testing"

	"github.com/eaguilar88/deu/internal/email"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/groups/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestService_GetRandomActiveGroups(t *testing.T) {
	type testCase struct {
		name    string
		limit   int
		prepare func(repoMock *mocks.MockRepository, limit int)
		want    []entities.ExtensionGroup
		wantErr error
	}

	tests := []testCase{
		{
			name:  "success",
			limit: 3,
			prepare: func(repoMock *mocks.MockRepository, limit int) {
				repoMock.EXPECT().GetRandomActiveGroups(mock.Anything, limit).RunAndReturn(
					func(ctx context.Context, limit int) ([]entities.ExtensionGroup, error) {
						return []entities.ExtensionGroup{{ID: "1"}, {ID: "2"}, {ID: "3"}}, nil
					})
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, mock.Anything, entities.OwnerTypeExtensionGroup).
					Return(entities.GroupedFiles{}, nil)
			},
			want: []entities.ExtensionGroup{{ID: "1"}, {ID: "2"}, {ID: "3"}},
		},
		{
			name:  "repository error",
			limit: 3,
			prepare: func(repoMock *mocks.MockRepository, limit int) {
				repoMock.EXPECT().GetRandomActiveGroups(mock.Anything, limit).RunAndReturn(
					func(ctx context.Context, limit int) ([]entities.ExtensionGroup, error) {
						return nil, errors.New("repository error")
					})
			},
			wantErr: errors.New("repository error"),
		},
	}

	ctx := context.Background()
	loggerMock := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			storageMock := mocks.NewMockStorageClient(t)
			if tt.prepare != nil {
				tt.prepare(repoMock, tt.limit)
			}
			s := NewService(repoMock, storageMock, nil, 0, loggerMock)
			got, err := s.GetRandomActiveGroups(ctx, tt.limit)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_GetGroups(t *testing.T) {
	type testCase struct {
		name      string
		filter    entities.GroupFilter
		pageScope entities.PageScope
		prepare   func(repoMock *mocks.MockRepository, filter entities.GroupFilter, pageScope entities.PageScope)
		want      []entities.ExtensionGroup
		wantErr   error
	}

	tests := []testCase{
		{
			name:   "success passes filter through",
			filter: entities.GroupFilter{Faculty: entities.FacultyIngenieria},
			prepare: func(repoMock *mocks.MockRepository, filter entities.GroupFilter, pageScope entities.PageScope) {
				repoMock.EXPECT().GetGroups(mock.Anything, filter, pageScope).RunAndReturn(
					func(ctx context.Context, filter entities.GroupFilter, pageScope entities.PageScope) ([]entities.ExtensionGroup, entities.PageScope, error) {
						return []entities.ExtensionGroup{{ID: "1", Faculty: []entities.Faculty{entities.FacultyIngenieria}}}, pageScope, nil
					})
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, mock.Anything, entities.OwnerTypeExtensionGroup).
					Return(entities.GroupedFiles{}, nil)
				repoMock.EXPECT().GetContactsByOwner(mock.Anything, mock.Anything, entities.OwnerTypeExtensionGroup).
					Return(nil, nil)
			},
			want: []entities.ExtensionGroup{{ID: "1", Faculty: []entities.Faculty{entities.FacultyIngenieria}}},
		},
		{
			name:   "repository error",
			filter: entities.GroupFilter{},
			prepare: func(repoMock *mocks.MockRepository, filter entities.GroupFilter, pageScope entities.PageScope) {
				repoMock.EXPECT().GetGroups(mock.Anything, filter, pageScope).RunAndReturn(
					func(ctx context.Context, filter entities.GroupFilter, pageScope entities.PageScope) ([]entities.ExtensionGroup, entities.PageScope, error) {
						return nil, entities.PageScope{}, errors.New("repository error")
					})
			},
			wantErr: errors.New("repository error"),
		},
	}

	ctx := context.Background()
	loggerMock := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			storageMock := mocks.NewMockStorageClient(t)
			if tt.prepare != nil {
				tt.prepare(repoMock, tt.filter, tt.pageScope)
			}
			s := NewService(repoMock, storageMock, nil, 0, loggerMock)
			got, _, err := s.GetGroups(ctx, tt.filter, tt.pageScope, entities.Viewer{UserID: "1", Roles: []string{"deu_admin"}})
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_GetGroup_Visibility(t *testing.T) {
	newGroup := func() entities.ExtensionGroup {
		return entities.ExtensionGroup{
			ID:        "1",
			Name:      "Coro",
			Owner:     &entities.User{ID: "10", CI: "123"},
			Faculty:   []entities.Faculty{entities.FacultyIngenieria},
			Members:   []entities.GroupMember{{ID: "m1", Name: "Ana", CI: 999}},
			Active:    true,
			UpdatedAt: "2026-01-01",
		}
	}
	groupFiles := entities.GroupedFiles{
		entities.GroupFileTypeLogo:    {{Key: "logo-key"}},
		entities.GroupFileTypeProject: {{Key: "project-key"}},
	}

	t.Run("anonymous gets the public view without private URLs", func(t *testing.T) {
		repoMock := mocks.NewMockRepository(t)
		storageMock := mocks.NewMockStorageClient(t)
		repoMock.EXPECT().GetGroupByID(mock.Anything, "1").Return(newGroup(), nil)
		repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1", entities.OwnerTypeExtensionGroup).Return(groupFiles, nil)
		repoMock.EXPECT().GetContactsByOwner(mock.Anything, "1", entities.OwnerTypeExtensionGroup).Return(nil, nil)
		storageMock.EXPECT().GetFileURL(mock.Anything, "logo-key").Return("https://files/logo-key", nil)
		// No GetPresignedFileURL and no member-file lookups: the mocks fail on unexpected calls.

		got, err := NewService(repoMock, storageMock, nil, 0, zap.NewNop()).GetGroup(context.Background(), "1", entities.Viewer{})

		assert.NoError(t, err)
		assert.Nil(t, got.Members)
		assert.Nil(t, got.Project)
		assert.Nil(t, got.Owner)
		assert.False(t, got.Active)
		assert.Empty(t, got.UpdatedAt)
		assert.Equal(t, "Coro", got.Name)
		assert.Equal(t, "https://files/logo-key", got.Logo.URL)
	})

	t.Run("owner gets members, project and pre-signed member documents", func(t *testing.T) {
		repoMock := mocks.NewMockRepository(t)
		storageMock := mocks.NewMockStorageClient(t)
		repoMock.EXPECT().GetGroupByID(mock.Anything, "1").Return(newGroup(), nil)
		repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1", entities.OwnerTypeExtensionGroup).Return(groupFiles, nil)
		repoMock.EXPECT().GetContactsByOwner(mock.Anything, "1", entities.OwnerTypeExtensionGroup).Return(nil, nil)
		repoMock.EXPECT().GetFilesByOwner(mock.Anything, "m1", entities.OwnerTypeGroupMember).
			Return(entities.GroupedFiles{entities.GroupMemberFileTypeDocument: {{Key: "doc-key"}}}, nil)
		storageMock.EXPECT().GetFileURL(mock.Anything, "logo-key").Return("https://files/logo-key", nil)
		storageMock.EXPECT().GetPresignedFileURL(mock.Anything, "project-key").Return("https://b2/project?sig", nil)
		storageMock.EXPECT().GetPresignedFileURL(mock.Anything, "doc-key").Return("https://b2/doc?sig", nil)

		got, err := NewService(repoMock, storageMock, nil, 0, zap.NewNop()).GetGroup(context.Background(), "1", entities.Viewer{UserID: "10"})

		assert.NoError(t, err)
		assert.Len(t, got.Members, 1)
		assert.Equal(t, "https://b2/doc?sig", got.Members[0].Document.URL)
		assert.Equal(t, "https://b2/project?sig", got.Project.URL)
		assert.NotNil(t, got.Owner)
		assert.True(t, got.Active)
	})
}

func TestService_GetGroups_MixedVisibility(t *testing.T) {
	repoMock := mocks.NewMockRepository(t)
	storageMock := mocks.NewMockStorageClient(t)
	repoMock.EXPECT().GetGroups(mock.Anything, mock.Anything, entities.PageScope{}).Return(
		[]entities.ExtensionGroup{
			{ID: "1", Owner: &entities.User{ID: "10"}, Members: []entities.GroupMember{{ID: "m1"}}},
			{ID: "2", Owner: &entities.User{ID: "20"}, Members: []entities.GroupMember{{ID: "m2"}}},
		}, entities.PageScope{}, nil)
	repoMock.EXPECT().GetFilesByOwner(mock.Anything, mock.Anything, entities.OwnerTypeExtensionGroup).
		Return(entities.GroupedFiles{}, nil)
	repoMock.EXPECT().GetContactsByOwner(mock.Anything, mock.Anything, entities.OwnerTypeExtensionGroup).
		Return(nil, nil)

	got, _, err := NewService(repoMock, storageMock, nil, 0, zap.NewNop()).
		GetGroups(context.Background(), entities.GroupFilter{}, entities.PageScope{}, entities.Viewer{UserID: "10"})

	assert.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Len(t, got[0].Members, 1, "the viewer's own group keeps its members")
	assert.Nil(t, got[1].Members, "other groups are redacted")
	assert.Nil(t, got[1].Owner)
}

func TestRestrictFilter(t *testing.T) {
	active := true
	inactive := false
	requested := entities.GroupFilter{Faculty: entities.FacultyIngenieria, Active: &inactive, Deleted: true}
	restricted := entities.GroupFilter{Faculty: entities.FacultyIngenieria, Active: &active, Deleted: false}

	tests := []struct {
		name   string
		viewer entities.Viewer
		want   entities.GroupFilter
	}{
		{name: "anonymous is limited to active groups", viewer: entities.Viewer{}, want: restricted},
		{name: "group_admin is limited to active groups", viewer: entities.Viewer{UserID: "10", Roles: []string{"group_admin"}}, want: restricted},
		{name: "deu_admin keeps the requested filter", viewer: entities.Viewer{UserID: "1", Roles: []string{"deu_admin"}}, want: requested},
		{name: "faculty_admin keeps the requested filter", viewer: entities.Viewer{UserID: "2", Roles: []string{"faculty_admin"}}, want: requested},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, restrictFilter(requested, tt.viewer))
		})
	}
}

func TestService_CreateGroup_NotifiesApprovers(t *testing.T) {
	owner := &entities.User{ID: "10"}
	facultyGroup := entities.ExtensionGroup{Name: "Grupo Coral", Owner: owner, Faculty: []entities.Faculty{entities.FacultyCiencias}}
	multiGroup := entities.ExtensionGroup{Name: "Grupo Mixto", Owner: owner, IsMultidisciplinary: true, Faculty: []entities.Faculty{entities.FacultyCiencias}}

	facultyData := email.GroupRequestSubmittedData{GroupName: "Grupo Coral", Faculty: "Ciencias"}
	multiData := email.GroupRequestSubmittedData{GroupName: "Grupo Mixto", Faculty: multidisciplinaryLabel}

	tests := []struct {
		name    string
		group   entities.ExtensionGroup
		prepare func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient)
	}{
		{
			name:  "faculty group notifies the faculty coordinators and the DEU",
			group: facultyGroup,
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetFacultyCoordinatorEmails(mock.Anything, entities.FacultyCiencias).Return([]string{"coord@example.com"}, nil)
				repoMock.EXPECT().GetDEUAdminEmails(mock.Anything).Return([]string{"deu@example.com"}, nil)
				mailMock.EXPECT().SendTemplate(mock.Anything, "coord@example.com", email.TemplateGroupRequestSubmittedFaculty, facultyData).Return(nil)
				mailMock.EXPECT().SendTemplate(mock.Anything, "deu@example.com", email.TemplateGroupRequestSubmittedDEU, facultyData).Return(nil)
			},
		},
		{
			name:  "multidisciplinary group notifies only the DEU",
			group: multiGroup,
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetDEUAdminEmails(mock.Anything).Return([]string{"deu@example.com"}, nil)
				mailMock.EXPECT().SendTemplate(mock.Anything, "deu@example.com", email.TemplateGroupRequestSubmittedDEU, multiData).Return(nil)
			},
		},
		{
			name:  "notification failures do not fail the creation",
			group: facultyGroup,
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetFacultyCoordinatorEmails(mock.Anything, entities.FacultyCiencias).Return(nil, errors.New("db down"))
				repoMock.EXPECT().GetDEUAdminEmails(mock.Anything).Return([]string{"deu@example.com"}, nil)
				mailMock.EXPECT().SendTemplate(mock.Anything, "deu@example.com", email.TemplateGroupRequestSubmittedDEU, facultyData).Return(errors.New("smtp down"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			storageMock := mocks.NewMockStorageClient(t)
			mailMock := mocks.NewMockMailClient(t)

			repoMock.EXPECT().CreateGroupWithRequests(mock.Anything, tt.group, mock.Anything).Return(int64(5), nil, nil)
			tt.prepare(repoMock, mailMock)

			id, _, err := NewService(repoMock, storageMock, mailMock, 0, zap.NewNop()).CreateGroup(context.Background(), tt.group)
			assert.NoError(t, err)
			assert.Equal(t, int64(5), id)
		})
	}
}
