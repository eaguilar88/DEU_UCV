package groups

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/eaguilar88/deu/internal/email"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/groups/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestService_RenewGroup(t *testing.T) {
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	window := 30 * 24 * time.Hour
	owner := &entities.User{ID: "50"}
	current := entities.ExtensionGroup{ID: "7", Owner: owner, RenewalDueAt: now.Add(10 * 24 * time.Hour)}

	newRenewal := func() entities.ExtensionGroup {
		return entities.ExtensionGroup{
			Name:    "Grupo Coral",
			Owner:   owner,
			Faculty: []entities.Faculty{entities.FacultyCiencias},
			Email:   "coral@example.com",
			Logo:    &entities.File{Name: "logo.png"},
			Project: &entities.File{Name: "proyecto_grupo.pdf"},
			Members: []entities.GroupMember{
				{Name: "Ana", Document: &entities.File{Name: "documento_miembro_0.pdf"}},
				{Name: "Luis", Document: &entities.File{Name: "documento_miembro_1.pdf"}},
			},
		}
	}
	submittedData := email.GroupRequestSubmittedData{GroupName: "Grupo Coral", Faculty: "Ciencias"}

	tests := []struct {
		name    string
		current entities.ExtensionGroup
		prepare func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient, mailMock *mocks.MockMailClient)
		wantID  int64
		wantErr error
	}{
		{
			name:    "stores the proposal, its files and notifies the approvers",
			current: current,
			prepare: func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().CreateGroupRenewal(mock.Anything, mock.MatchedBy(func(r entities.GroupRenewal) bool {
					return r.GroupID == "7" && r.SubmittedBy == "50" && r.Group.ID == "7" && r.Group.Name == "Grupo Coral"
				}), mock.MatchedBy(func(reqs []entities.GroupRequest) bool {
					return len(reqs) == 2 && reqs[0].Faculty == entities.FacultyCiencias && reqs[1].Faculty == entities.FacultyDEU
				})).Return(int64(3), nil)
				storageMock.EXPECT().UploadFile(mock.Anything, mock.MatchedBy(func(files []*entities.File) bool {
					if len(files) != 4 {
						return false
					}
					for _, f := range files {
						if f.OwnerType != entities.OwnerTypeGroupRenewal || f.OwnerID != "3" || f.UploadedBy != "50" {
							return false
						}
					}
					logo, project, doc := files[0], files[1], files[3]
					return logo.Key == "files/groups/7/renewals/3/logo_logo.png" && logo.Public &&
						project.Purpose == entities.GroupFileTypeProject && !project.Public &&
						doc.Purpose == entities.GroupMemberFileTypeDocument &&
						doc.MetaData[entities.GroupRenewalMemberIndexKey] == "1" &&
						doc.Key == "files/groups/7/renewals/3/members/1/documento_miembro_documento_miembro_1.pdf"
				})).Return(nil)
				repoMock.EXPECT().SaveFilesToDB(mock.Anything, mock.Anything).Return(nil)
				repoMock.EXPECT().GetFacultyCoordinatorEmails(mock.Anything, entities.FacultyCiencias).Return([]string{"coord@example.com"}, nil)
				repoMock.EXPECT().GetDEUAdminEmails(mock.Anything).Return([]string{"deu@example.com"}, nil)
				mailMock.EXPECT().SendTemplate(mock.Anything, "coord@example.com", email.TemplateGroupRenewalSubmittedFaculty, submittedData).Return(nil)
				mailMock.EXPECT().SendTemplate(mock.Anything, "deu@example.com", email.TemplateGroupRenewalSubmittedDEU, submittedData).Return(nil)
			},
			wantID: 3,
		},
		{
			name:    "overdue group can still renew",
			current: entities.ExtensionGroup{ID: "7", Owner: owner, RenewalDueAt: now.Add(-60 * 24 * time.Hour)},
			prepare: func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().CreateGroupRenewal(mock.Anything, mock.Anything, mock.Anything).Return(int64(3), nil)
				storageMock.EXPECT().UploadFile(mock.Anything, mock.Anything).Return(nil)
				repoMock.EXPECT().SaveFilesToDB(mock.Anything, mock.Anything).Return(nil)
				repoMock.EXPECT().GetFacultyCoordinatorEmails(mock.Anything, mock.Anything).Return(nil, nil)
				repoMock.EXPECT().GetDEUAdminEmails(mock.Anything).Return(nil, nil)
			},
			wantID: 3,
		},
		{
			name:    "someone else than the group's user is rejected",
			current: entities.ExtensionGroup{ID: "7", Owner: &entities.User{ID: "99"}, RenewalDueAt: current.RenewalDueAt},
			wantErr: ErrNotGroupOwner,
		},
		{
			name:    "never approved group cannot renew",
			current: entities.ExtensionGroup{ID: "7", Owner: owner},
			wantErr: ErrRenewalNotOpen,
		},
		{
			name:    "renewal before the window is rejected",
			current: entities.ExtensionGroup{ID: "7", Owner: owner, RenewalDueAt: now.Add(window + time.Hour)},
			wantErr: ErrRenewalNotOpen,
		},
		{
			name:    "pending renewal is reported",
			current: current,
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockStorageClient, _ *mocks.MockMailClient) {
				repoMock.EXPECT().CreateGroupRenewal(mock.Anything, mock.Anything, mock.Anything).Return(int64(-1), ErrRenewalPending)
			},
			wantErr: ErrRenewalPending,
		},
		{
			name:    "upload failure cancels the renewal",
			current: current,
			prepare: func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient, _ *mocks.MockMailClient) {
				repoMock.EXPECT().CreateGroupRenewal(mock.Anything, mock.Anything, mock.Anything).Return(int64(3), nil)
				storageMock.EXPECT().UploadFile(mock.Anything, mock.Anything).Return(errors.New("b2 down"))
				repoMock.EXPECT().CancelGroupRenewal(mock.Anything, "3").Return(nil)
			},
			wantErr: errors.New("failed to upload files: b2 down"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			storageMock := mocks.NewMockStorageClient(t)
			mailMock := mocks.NewMockMailClient(t)
			repoMock.EXPECT().GetGroupByID(mock.Anything, "7").Return(tt.current, nil)
			if tt.prepare != nil {
				tt.prepare(repoMock, storageMock, mailMock)
			}

			svc := NewService(repoMock, storageMock, mailMock, window, zap.NewNop()).(*service)
			svc.now = func() time.Time { return now }

			id, err := svc.RenewGroup(context.Background(), "7", newRenewal())
			if tt.wantErr != nil {
				if errors.Is(err, tt.wantErr) {
					return
				}
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantID, id)
		})
	}
}

func TestMapGroupError_Renewal(t *testing.T) {
	tests := map[error]int{
		ErrGroupNotFound:  http.StatusNotFound,
		ErrNotGroupOwner:  http.StatusForbidden,
		ErrRenewalNotOpen: http.StatusConflict,
		ErrRenewalPending: http.StatusConflict,
	}
	for err, want := range tests {
		got := mapGroupError(fmt.Errorf("wrapped: %w", err)).(interface{ StatusCode() int })
		assert.Equal(t, want, got.StatusCode(), err.Error())
	}
}
