package group_resource_requests_test

import (
	"context"
	"errors"
	"testing"

	"github.com/eaguilar88/deu/internal/email"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/group_resource_requests"
	"github.com/eaguilar88/deu/internal/group_resource_requests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestService_ReviewGroupResourceRequest(t *testing.T) {
	baseReq := entities.GroupResourceRequest{ID: "req-1", GroupID: "1", GroupName: "Grupo Test", Type: "Auditorio", Status: "under_review"}
	emailContact := []entities.Contact{
		{Type: entities.ContactTypePhone, Value: "0212-5555555"},
		{Type: entities.ContactTypeEmail, Value: "grupo@example.com"},
	}
	dbErr := errors.New("db error")
	reason := "Documentación incompleta"

	type action struct {
		name     string
		template email.Template
		data     any
		mockOp   func(repoMock *mocks.MockRepository, err error)
		call     func(s group_resource_requests.Service) error
	}
	ops := []action{
		{
			name:     "approve",
			template: email.TemplateGroupResourceRequestApproved,
			data:     email.GroupResourceRequestApprovedData{GroupName: "Grupo Test", ResourceType: "Auditorio"},
			mockOp: func(repoMock *mocks.MockRepository, err error) {
				repoMock.EXPECT().ApproveGroupResourceRequest(mock.Anything, "req-1").Return(err)
			},
			call: func(s group_resource_requests.Service) error {
				return s.ApproveGroupResourceRequest(context.Background(), "req-1")
			},
		},
		{
			name:     "reject",
			template: email.TemplateGroupResourceRequestRejected,
			data:     email.GroupResourceRequestRejectedData{GroupName: "Grupo Test", ResourceType: "Auditorio", Reason: reason},
			mockOp: func(repoMock *mocks.MockRepository, err error) {
				repoMock.EXPECT().RejectGroupResourceRequest(mock.Anything, "req-1", reason).Return(err)
			},
			call: func(s group_resource_requests.Service) error {
				return s.RejectGroupResourceRequest(context.Background(), "req-1", reason)
			},
		},
	}

	for _, op := range ops {
		tests := []struct {
			name    string
			prepare func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient)
			wantErr error
		}{
			{
				name: "emails the group contact",
				prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
					repoMock.EXPECT().GetGroupResourceRequestByID(mock.Anything, "req-1").Return(baseReq, nil)
					op.mockOp(repoMock, nil)
					repoMock.EXPECT().GetContactsByOwner(mock.Anything, "1", entities.OwnerTypeExtensionGroup).Return(emailContact, nil)
					mailMock.EXPECT().SendTemplate(mock.Anything, "grupo@example.com", op.template, op.data).Return(nil)
				},
			},
			{
				name: "request not found is propagated",
				prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockMailClient) {
					repoMock.EXPECT().GetGroupResourceRequestByID(mock.Anything, "req-1").Return(entities.GroupResourceRequest{}, group_resource_requests.ErrNotFound)
				},
				wantErr: group_resource_requests.ErrNotFound,
			},
			{
				name: "status update error is returned and no email is sent",
				prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockMailClient) {
					repoMock.EXPECT().GetGroupResourceRequestByID(mock.Anything, "req-1").Return(baseReq, nil)
					op.mockOp(repoMock, dbErr)
				},
				wantErr: dbErr,
			},
			{
				name: "email send error does not fail the request",
				prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
					repoMock.EXPECT().GetGroupResourceRequestByID(mock.Anything, "req-1").Return(baseReq, nil)
					op.mockOp(repoMock, nil)
					repoMock.EXPECT().GetContactsByOwner(mock.Anything, "1", entities.OwnerTypeExtensionGroup).Return(emailContact, nil)
					mailMock.EXPECT().SendTemplate(mock.Anything, "grupo@example.com", mock.Anything, mock.Anything).Return(errors.New("smtp error"))
				},
			},
			{
				name: "group without email contact skips the email",
				prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockMailClient) {
					repoMock.EXPECT().GetGroupResourceRequestByID(mock.Anything, "req-1").Return(baseReq, nil)
					op.mockOp(repoMock, nil)
					repoMock.EXPECT().GetContactsByOwner(mock.Anything, "1", entities.OwnerTypeExtensionGroup).Return(nil, nil)
				},
			},
			{
				name: "contacts lookup error does not fail the request",
				prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockMailClient) {
					repoMock.EXPECT().GetGroupResourceRequestByID(mock.Anything, "req-1").Return(baseReq, nil)
					op.mockOp(repoMock, nil)
					repoMock.EXPECT().GetContactsByOwner(mock.Anything, "1", entities.OwnerTypeExtensionGroup).Return(nil, dbErr)
				},
			},
		}

		for _, tt := range tests {
			t.Run(op.name+"/"+tt.name, func(t *testing.T) {
				repoMock := mocks.NewMockRepository(t)
				mailMock := mocks.NewMockMailClient(t)
				tt.prepare(repoMock, mailMock)

				err := op.call(group_resource_requests.NewService(repoMock, mailMock, zap.NewNop()))
				if tt.wantErr != nil {
					assert.ErrorIs(t, err, tt.wantErr)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	}
}
