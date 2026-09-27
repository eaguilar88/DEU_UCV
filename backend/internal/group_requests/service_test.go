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

	// runNotify mimics the repository: it runs the notification before "committing"
	// and propagates its error, as the real transaction would roll back.
	runNotify := func(_ context.Context, _, _ string, _ entities.User, notify func() error) (int64, error) {
		if err := notify(); err != nil {
			return -1, err
		}
		return 42, nil
	}

	tests := []testCase{
		{
			name: "last approval activates group and emails original owner",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetGroupRequestByID(mock.Anything, "req-1").Return(baseReq, nil)
				repoMock.EXPECT().GetGroupRequestsByGroupID(mock.Anything, "1").Return([]entities.GroupRequest{baseReq, otherApproved}, nil)
				repoMock.EXPECT().GetGroupByID(mock.Anything, "1").Return(entities.ExtensionGroup{ID: "1", Owner: &entities.User{ID: "owner-1"}}, nil)
				repoMock.EXPECT().GetUser(mock.Anything, "owner-1").Return(&entities.User{ID: "owner-1", Email: "owner@test.com"}, nil)
				repoMock.EXPECT().ApproveGroupRequestAndActivate(mock.Anything, "req-1", "1", mock.AnythingOfType("entities.User"), mock.Anything).
					RunAndReturn(func(ctx context.Context, reqID, groupID string, u entities.User, notify func() error) (int64, error) {
						assert.NotEqual(t, "nolodire", u.Password)
						return runNotify(ctx, reqID, groupID, u, notify)
					})
				mailMock.EXPECT().Send(mock.Anything, "owner@test.com", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "other request still pending only approves this request",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockMailClient) {
				repoMock.EXPECT().GetGroupRequestByID(mock.Anything, "req-1").Return(baseReq, nil)
				repoMock.EXPECT().GetGroupRequestsByGroupID(mock.Anything, "1").Return([]entities.GroupRequest{baseReq, otherPending}, nil)
				repoMock.EXPECT().ApproveGroupRequest(mock.Anything, "req-1").Return(nil)
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
			name: "approve error is propagated when other requests are pending",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockMailClient) {
				repoMock.EXPECT().GetGroupRequestByID(mock.Anything, "req-1").Return(baseReq, nil)
				repoMock.EXPECT().GetGroupRequestsByGroupID(mock.Anything, "1").Return([]entities.GroupRequest{baseReq, otherPending}, nil)
				repoMock.EXPECT().ApproveGroupRequest(mock.Anything, "req-1").Return(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "get group error short-circuits before user creation",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockMailClient) {
				repoMock.EXPECT().GetGroupRequestByID(mock.Anything, "req-1").Return(baseReq, nil)
				repoMock.EXPECT().GetGroupRequestsByGroupID(mock.Anything, "1").Return([]entities.GroupRequest{baseReq, otherApproved}, nil)
				repoMock.EXPECT().GetGroupByID(mock.Anything, "1").Return(entities.ExtensionGroup{}, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "repo transaction error prevents email from being sent",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockMailClient) {
				repoMock.EXPECT().GetGroupRequestByID(mock.Anything, "req-1").Return(baseReq, nil)
				repoMock.EXPECT().GetGroupRequestsByGroupID(mock.Anything, "1").Return([]entities.GroupRequest{baseReq, otherApproved}, nil)
				repoMock.EXPECT().GetGroupByID(mock.Anything, "1").Return(entities.ExtensionGroup{ID: "1", Owner: &entities.User{ID: "owner-1"}}, nil)
				repoMock.EXPECT().GetUser(mock.Anything, "owner-1").Return(&entities.User{ID: "owner-1", Email: "owner@test.com"}, nil)
				repoMock.EXPECT().ApproveGroupRequestAndActivate(mock.Anything, "req-1", "1", mock.AnythingOfType("entities.User"), mock.Anything).
					Return(int64(-1), errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "email send error rolls back and is propagated",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetGroupRequestByID(mock.Anything, "req-1").Return(baseReq, nil)
				repoMock.EXPECT().GetGroupRequestsByGroupID(mock.Anything, "1").Return([]entities.GroupRequest{baseReq, otherApproved}, nil)
				repoMock.EXPECT().GetGroupByID(mock.Anything, "1").Return(entities.ExtensionGroup{ID: "1", Owner: &entities.User{ID: "owner-1"}}, nil)
				repoMock.EXPECT().GetUser(mock.Anything, "owner-1").Return(&entities.User{ID: "owner-1", Email: "owner@test.com"}, nil)
				repoMock.EXPECT().ApproveGroupRequestAndActivate(mock.Anything, "req-1", "1", mock.AnythingOfType("entities.User"), mock.Anything).
					RunAndReturn(runNotify)
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
