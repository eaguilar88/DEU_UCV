package group_requests

import (
	"context"
	"errors"
	"testing"

	"github.com/eaguilar88/deu/internal/email"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/group_requests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestService_ApproveGroupRequest_Renewal(t *testing.T) {
	renewalReq := entities.GroupRequest{ID: "req-5", GroupID: "1", GroupName: "Grupo Test", RenewalID: "9", Status: entities.RequestStatus_UNDER_REVIEW}
	renewalSiblingApproved := entities.GroupRequest{ID: "req-6", GroupID: "1", RenewalID: "9", Status: entities.RequestStatus_APPROVED}
	renewalSiblingPending := entities.GroupRequest{ID: "req-6", GroupID: "1", RenewalID: "9", Status: entities.RequestStatus_UNDER_REVIEW}
	// Earlier rounds of the same group must not affect this one.
	creationApproved := entities.GroupRequest{ID: "req-1", GroupID: "1", Status: entities.RequestStatus_APPROVED}
	oldRenewalRejected := entities.GroupRequest{ID: "req-3", GroupID: "1", RenewalID: "4", Status: entities.RequestStatus_REJECTED}

	renewal := entities.GroupRenewal{ID: "9", GroupID: "1", Group: entities.ExtensionGroup{Name: "Grupo Renovado", Email: "nuevo@example.com"}}
	approvedData := email.GroupRequestApprovedData{GroupName: "Grupo Renovado"}

	tests := []struct {
		name    string
		prepare func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient)
		wantErr bool
	}{
		{
			name: "last renewal approval applies the renewal and emails the group",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetGroupRequestByID(mock.Anything, "req-5").Return(renewalReq, nil)
				repoMock.EXPECT().GetGroupRequestsByGroupID(mock.Anything, "1").
					Return([]entities.GroupRequest{renewalReq, renewalSiblingApproved, creationApproved, oldRenewalRejected}, nil)
				repoMock.EXPECT().GetGroupRenewal(mock.Anything, "9").Return(renewal, nil)
				repoMock.EXPECT().ApplyGroupRenewal(mock.Anything, "req-5", renewal).Return(nil)
				mailMock.EXPECT().SendTemplate(mock.Anything, "nuevo@example.com", email.TemplateGroupRenewalApproved, approvedData).Return(nil)
			},
		},
		{
			name: "pending renewal sibling only approves this request",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockMailClient) {
				repoMock.EXPECT().GetGroupRequestByID(mock.Anything, "req-5").Return(renewalReq, nil)
				repoMock.EXPECT().GetGroupRequestsByGroupID(mock.Anything, "1").
					Return([]entities.GroupRequest{renewalReq, renewalSiblingPending, creationApproved}, nil)
				repoMock.EXPECT().ApproveGroupRequest(mock.Anything, "req-5").Return(nil)
			},
		},
		{
			name: "apply error is propagated and no email is sent",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockMailClient) {
				repoMock.EXPECT().GetGroupRequestByID(mock.Anything, "req-5").Return(renewalReq, nil)
				repoMock.EXPECT().GetGroupRequestsByGroupID(mock.Anything, "1").Return([]entities.GroupRequest{renewalReq}, nil)
				repoMock.EXPECT().GetGroupRenewal(mock.Anything, "9").Return(renewal, nil)
				repoMock.EXPECT().ApplyGroupRenewal(mock.Anything, "req-5", renewal).Return(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "email failure does not fail an applied renewal",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetGroupRequestByID(mock.Anything, "req-5").Return(renewalReq, nil)
				repoMock.EXPECT().GetGroupRequestsByGroupID(mock.Anything, "1").Return([]entities.GroupRequest{renewalReq}, nil)
				repoMock.EXPECT().GetGroupRenewal(mock.Anything, "9").Return(renewal, nil)
				repoMock.EXPECT().ApplyGroupRenewal(mock.Anything, "req-5", renewal).Return(nil)
				mailMock.EXPECT().SendTemplate(mock.Anything, "nuevo@example.com", email.TemplateGroupRenewalApproved, approvedData).Return(errors.New("smtp error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			mailMock := mocks.NewMockMailClient(t)
			tt.prepare(repoMock, mailMock)

			err := NewService(repoMock, mailMock, nil, zap.NewNop()).ApproveGroupRequest(context.Background(), "req-5")
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestService_ApproveGroupRequest_CreationIgnoresRenewals(t *testing.T) {
	creationReq := entities.GroupRequest{ID: "req-1", GroupID: "1", GroupName: "Grupo Test", Status: entities.RequestStatus_UNDER_REVIEW}
	creationPending := entities.GroupRequest{ID: "req-2", GroupID: "1", Status: entities.RequestStatus_UNDER_REVIEW}
	renewalApproved := entities.GroupRequest{ID: "req-5", GroupID: "1", RenewalID: "9", Status: entities.RequestStatus_APPROVED}

	repoMock := mocks.NewMockRepository(t)
	repoMock.EXPECT().GetGroupRequestByID(mock.Anything, "req-1").Return(creationReq, nil)
	repoMock.EXPECT().GetGroupRequestsByGroupID(mock.Anything, "1").Return([]entities.GroupRequest{creationReq, creationPending, renewalApproved}, nil)
	repoMock.EXPECT().ApproveGroupRequest(mock.Anything, "req-1").Return(nil)

	err := NewService(repoMock, mocks.NewMockMailClient(t), nil, zap.NewNop()).ApproveGroupRequest(context.Background(), "req-1")
	assert.NoError(t, err)
}

func TestService_RejectGroupRequest_Renewal(t *testing.T) {
	renewalReq := entities.GroupRequest{ID: "req-5", GroupID: "1", GroupName: "Grupo Test", RenewalID: "9", Status: entities.RequestStatus_UNDER_REVIEW}
	reason := "Faltan documentos"

	repoMock := mocks.NewMockRepository(t)
	mailMock := mocks.NewMockMailClient(t)
	repoMock.EXPECT().GetGroupRequestByID(mock.Anything, "req-5").Return(renewalReq, nil)
	repoMock.EXPECT().RejectGroupRenewal(mock.Anything, "req-5", "9", reason).Return(nil)
	repoMock.EXPECT().GetContactsByOwner(mock.Anything, "1", entities.OwnerTypeExtensionGroup).
		Return([]entities.Contact{{Type: entities.ContactTypeEmail, Value: "grupo@example.com"}}, nil)
	mailMock.EXPECT().SendTemplate(mock.Anything, "grupo@example.com", email.TemplateGroupRequestRejected,
		email.GroupRequestRejectedData{GroupName: "Grupo Test", Reason: reason}).Return(nil)

	err := NewService(repoMock, mailMock, nil, zap.NewNop()).RejectGroupRequest(context.Background(), "req-5", reason)
	assert.NoError(t, err)
}

func TestService_GetGroupRequestByID_Renewal(t *testing.T) {
	renewalReq := entities.GroupRequest{ID: "req-5", GroupID: "1", RenewalID: "9"}
	creationApproved := entities.GroupRequest{ID: "req-1", GroupID: "1"}
	renewal := entities.GroupRenewal{ID: "9", GroupID: "1", Group: entities.ExtensionGroup{
		Name:    "Grupo Renovado",
		Members: []entities.GroupMember{{Name: "Ana"}, {Name: "Luis"}},
	}}
	files := entities.GroupedFiles{
		entities.GroupFileTypeLogo: {{Key: "logo-key"}},
		entities.GroupMemberFileTypeDocument: {
			{Key: "doc-1", MetaData: map[string]string{entities.GroupRenewalMemberIndexKey: "1"}},
			{Key: "doc-bad", MetaData: map[string]string{entities.GroupRenewalMemberIndexKey: "7"}},
		},
	}

	repoMock := mocks.NewMockRepository(t)
	storageMock := mocks.NewMockStorageClient(t)
	repoMock.EXPECT().GetGroupRequestByID(mock.Anything, "req-5").Return(renewalReq, nil)
	repoMock.EXPECT().GetGroupRequestsByGroupID(mock.Anything, "1").Return([]entities.GroupRequest{renewalReq, creationApproved}, nil)
	repoMock.EXPECT().GetGroupRenewal(mock.Anything, "9").Return(renewal, nil)
	repoMock.EXPECT().GetFilesByOwner(mock.Anything, "9", entities.OwnerTypeGroupRenewal).Return(files, nil)
	storageMock.EXPECT().GetPresignedFileURL(mock.Anything, "logo-key").Return("https://files/logo", nil)
	storageMock.EXPECT().GetPresignedFileURL(mock.Anything, "doc-1").Return("https://files/doc-1", nil)

	got, err := NewService(repoMock, mocks.NewMockMailClient(t), storageMock, zap.NewNop()).GetGroupRequestByID(context.Background(), "req-5")
	assert.NoError(t, err)
	assert.Equal(t, []entities.GroupRequest{renewalReq}, got.Approvals, "approvals only include this renewal's requests")
	if assert.NotNil(t, got.Renewal) {
		assert.Equal(t, "https://files/logo", got.Renewal.Group.Logo.URL)
		assert.Nil(t, got.Renewal.Group.Project)
		assert.Nil(t, got.Renewal.Group.Members[0].Document)
		assert.Equal(t, "https://files/doc-1", got.Renewal.Group.Members[1].Document.URL)
	}
}
