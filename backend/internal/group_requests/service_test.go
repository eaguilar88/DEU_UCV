package group_requests

import (
	"context"
	"errors"
	"testing"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/group_requests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestService_ApproveGroupRequest(t *testing.T) {
	type testCase struct {
		name    string
		prepare func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient)
		wantErr bool
	}

	baseReq := entities.GroupRequest{ID: "req-1", GroupID: "1", GroupName: "Grupo Test", Status: entities.RequestStatus_UNDER_REVIEW}
	otherApproved := entities.GroupRequest{ID: "req-2", GroupID: "1", Status: entities.RequestStatus_APPROVED}
	otherPending := entities.GroupRequest{ID: "req-2", GroupID: "1", Status: entities.RequestStatus_UNDER_REVIEW}

	tests := []testCase{
		{
			name: "last approval activates group and emails original owner",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetGroupRequestByID(mock.Anything, "req-1").Return(baseReq, nil)
				repoMock.EXPECT().ApproveGroupRequest(mock.Anything, "req-1").Return(nil)
				repoMock.EXPECT().GetGroupRequestsByGroupID(mock.Anything, "1").Return([]entities.GroupRequest{baseReq, otherApproved}, nil)
				repoMock.EXPECT().GetGroupByID(mock.Anything, "1").Return(entities.ExtensionGroup{ID: "1", Owner: &entities.User{ID: "owner-1"}}, nil)
				repoMock.EXPECT().GetUser(mock.Anything, "owner-1").Return(&entities.User{ID: "owner-1", Email: "owner@test.com"}, nil)
				repoMock.EXPECT().CreateGroupAdminAndActivate(mock.Anything, "1", mock.AnythingOfType("entities.User")).
					RunAndReturn(func(_ context.Context, _ string, u entities.User) (int64, error) {
						assert.NotEqual(t, "nolodire", u.Password)
						return 42, nil
					})
				mailMock.EXPECT().Send(mock.Anything, "owner@test.com", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "other request still pending does not activate group",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockMailClient) {
				repoMock.EXPECT().GetGroupRequestByID(mock.Anything, "req-1").Return(baseReq, nil)
				repoMock.EXPECT().ApproveGroupRequest(mock.Anything, "req-1").Return(nil)
				repoMock.EXPECT().GetGroupRequestsByGroupID(mock.Anything, "1").Return([]entities.GroupRequest{baseReq, otherPending}, nil)
			},
			wantErr: false,
		},
		{
			name: "fetch request error short-circuits",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockMailClient) {
				repoMock.EXPECT().GetGroupRequestByID(mock.Anything, "req-1").Return(entities.GroupRequest{}, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "approve error short-circuits",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockMailClient) {
				repoMock.EXPECT().GetGroupRequestByID(mock.Anything, "req-1").Return(baseReq, nil)
				repoMock.EXPECT().ApproveGroupRequest(mock.Anything, "req-1").Return(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "get group error short-circuits before user creation",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockMailClient) {
				repoMock.EXPECT().GetGroupRequestByID(mock.Anything, "req-1").Return(baseReq, nil)
				repoMock.EXPECT().ApproveGroupRequest(mock.Anything, "req-1").Return(nil)
				repoMock.EXPECT().GetGroupRequestsByGroupID(mock.Anything, "1").Return([]entities.GroupRequest{baseReq, otherApproved}, nil)
				repoMock.EXPECT().GetGroupByID(mock.Anything, "1").Return(entities.ExtensionGroup{}, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "repo transaction error prevents email from being sent",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockMailClient) {
				repoMock.EXPECT().GetGroupRequestByID(mock.Anything, "req-1").Return(baseReq, nil)
				repoMock.EXPECT().ApproveGroupRequest(mock.Anything, "req-1").Return(nil)
				repoMock.EXPECT().GetGroupRequestsByGroupID(mock.Anything, "1").Return([]entities.GroupRequest{baseReq, otherApproved}, nil)
				repoMock.EXPECT().GetGroupByID(mock.Anything, "1").Return(entities.ExtensionGroup{ID: "1", Owner: &entities.User{ID: "owner-1"}}, nil)
				repoMock.EXPECT().GetUser(mock.Anything, "owner-1").Return(&entities.User{ID: "owner-1", Email: "owner@test.com"}, nil)
				repoMock.EXPECT().CreateGroupAdminAndActivate(mock.Anything, "1", mock.AnythingOfType("entities.User")).Return(int64(0), errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "email send error is propagated",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetGroupRequestByID(mock.Anything, "req-1").Return(baseReq, nil)
				repoMock.EXPECT().ApproveGroupRequest(mock.Anything, "req-1").Return(nil)
				repoMock.EXPECT().GetGroupRequestsByGroupID(mock.Anything, "1").Return([]entities.GroupRequest{baseReq, otherApproved}, nil)
				repoMock.EXPECT().GetGroupByID(mock.Anything, "1").Return(entities.ExtensionGroup{ID: "1", Owner: &entities.User{ID: "owner-1"}}, nil)
				repoMock.EXPECT().GetUser(mock.Anything, "owner-1").Return(&entities.User{ID: "owner-1", Email: "owner@test.com"}, nil)
				repoMock.EXPECT().CreateGroupAdminAndActivate(mock.Anything, "1", mock.AnythingOfType("entities.User")).Return(int64(42), nil)
				mailMock.EXPECT().Send(mock.Anything, "owner@test.com", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(errors.New("smtp error"))
			},
			wantErr: true,
		},
	}

	loggerMock := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			mailMock := mocks.NewMockMailClient(t)
			if tt.prepare != nil {
				tt.prepare(repoMock, mailMock)
			}
			s := NewService(repoMock, mailMock, loggerMock)
			err := s.ApproveGroupRequest(context.Background(), "req-1")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			repoMock.AssertExpectations(t)
			mailMock.AssertExpectations(t)
		})
	}
}
