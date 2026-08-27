package course_cycle_close_requests

import (
	"context"
	"errors"
	"testing"

	"github.com/eaguilar88/deu/internal/course_cycle_close_requests/mocks"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestNewService(t *testing.T) {
	repoMock := mocks.NewMockRepository(t)
	storageMock := mocks.NewMockStorageClient(t)
	got := NewService(repoMock, storageMock, zap.NewNop())
	assert.NotNil(t, got)
}

func newValidCloseRequest() entities.CourseCycleCloseRequest {
	return entities.CourseCycleCloseRequest{
		CourseCycleID:    1,
		Comments:         "cierre de cohorte",
		ParticipantsFile: &entities.File{Name: "archivo_participantes.pdf"},
		VouchersFile:     &entities.File{Name: "archivo_vouchers.pdf"},
		SurveyFile:       &entities.File{Name: "archivo_encuesta.pdf"},
	}
}

func TestCourseCycleCloseRequestService_SubmitCloseRequest(t *testing.T) {
	type testCase struct {
		name          string
		request       entities.CourseCycleCloseRequest
		submittedByID string
		prepare       func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient)
		want          int64
		wantErr       error
	}

	tests := []testCase{
		{
			name:          "success",
			request:       newValidCloseRequest(),
			submittedByID: "1",
			prepare: func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient) {
				repoMock.EXPECT().HasPendingCloseRequestForCycle(mock.Anything, int64(1)).Return(false, nil)
				repoMock.EXPECT().CreateCourseCycleCloseRequest(mock.Anything, int64(1), int64(1)).Return(int64(10), nil)
				storageMock.EXPECT().UploadFile(mock.Anything, mock.MatchedBy(func(files []*entities.File) bool {
					return len(files) == 3
				})).Return(nil)
				repoMock.EXPECT().SaveFilesToDB(mock.Anything, mock.MatchedBy(func(files []*entities.File) bool {
					return len(files) == 3
				})).Return(nil)
			},
			want:    int64(10),
			wantErr: nil,
		},
		{
			name:          "error invalid submitter ID",
			request:       newValidCloseRequest(),
			submittedByID: "not-a-number",
			want:          -1,
			wantErr:       errors.New("invalid submitter ID: strconv.ParseInt: parsing \"not-a-number\": invalid syntax"),
		},
		{
			name:          "error checking pending close requests",
			request:       newValidCloseRequest(),
			submittedByID: "1",
			prepare: func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient) {
				repoMock.EXPECT().HasPendingCloseRequestForCycle(mock.Anything, int64(1)).
					Return(false, errors.New("db error"))
			},
			want:    -1,
			wantErr: errors.New("db error"),
		},
		{
			name:          "error a close request is already pending",
			request:       newValidCloseRequest(),
			submittedByID: "1",
			prepare: func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient) {
				repoMock.EXPECT().HasPendingCloseRequestForCycle(mock.Anything, int64(1)).Return(true, nil)
			},
			want:    -1,
			wantErr: ErrCloseRequestAlreadyPending,
		},
		{
			name:          "error creating close request",
			request:       newValidCloseRequest(),
			submittedByID: "1",
			prepare: func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient) {
				repoMock.EXPECT().HasPendingCloseRequestForCycle(mock.Anything, int64(1)).Return(false, nil)
				repoMock.EXPECT().CreateCourseCycleCloseRequest(mock.Anything, int64(1), int64(1)).
					Return(int64(-1), errors.New("insert error"))
			},
			want:    -1,
			wantErr: errors.New("insert error"),
		},
		{
			name:          "error uploading evidence files",
			request:       newValidCloseRequest(),
			submittedByID: "1",
			prepare: func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient) {
				repoMock.EXPECT().HasPendingCloseRequestForCycle(mock.Anything, int64(1)).Return(false, nil)
				repoMock.EXPECT().CreateCourseCycleCloseRequest(mock.Anything, int64(1), int64(1)).Return(int64(10), nil)
				storageMock.EXPECT().UploadFile(mock.Anything, mock.Anything).Return(errors.New("upload error"))
			},
			want:    -1,
			wantErr: errors.New("upload error"),
		},
		{
			name:          "error saving evidence file metadata",
			request:       newValidCloseRequest(),
			submittedByID: "1",
			prepare: func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient) {
				repoMock.EXPECT().HasPendingCloseRequestForCycle(mock.Anything, int64(1)).Return(false, nil)
				repoMock.EXPECT().CreateCourseCycleCloseRequest(mock.Anything, int64(1), int64(1)).Return(int64(10), nil)
				storageMock.EXPECT().UploadFile(mock.Anything, mock.Anything).Return(nil)
				repoMock.EXPECT().SaveFilesToDB(mock.Anything, mock.Anything).Return(errors.New("save error"))
			},
			want:    -1,
			wantErr: errors.New("save error"),
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			storageMock := mocks.NewMockStorageClient(t)
			if tt.prepare != nil {
				tt.prepare(repoMock, storageMock)
			}
			s := NewService(repoMock, storageMock, zap.NewNop())
			got, err := s.SubmitCloseRequest(ctx, tt.request, tt.submittedByID)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCourseCycleCloseRequestService_ApproveCloseRequest(t *testing.T) {
	type testCase struct {
		name    string
		prepare func(repoMock *mocks.MockRepository)
		wantErr error
	}

	tests := []testCase{
		{
			name: "success",
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetCourseCycleCloseRequestByID(mock.Anything, "1").
					Return(entities.CourseCycleCloseRequest{ID: 1, CourseCycleID: 2, Status: entities.RequestStatus_UNDER_REVIEW}, nil)
				repoMock.EXPECT().ApproveCourseCycleCloseRequest(mock.Anything, "1", "reviewer-1", "2").Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "error getting close request",
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetCourseCycleCloseRequestByID(mock.Anything, "1").
					Return(entities.CourseCycleCloseRequest{}, ErrCycleCloseRequestNotFound)
			},
			wantErr: ErrCycleCloseRequestNotFound,
		},
		{
			name: "error request already processed",
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetCourseCycleCloseRequestByID(mock.Anything, "1").
					Return(entities.CourseCycleCloseRequest{ID: 1, Status: entities.RequestStatus_APPROVED}, nil)
			},
			wantErr: ErrRequestIsProcessed,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			if tt.prepare != nil {
				tt.prepare(repoMock)
			}
			s := NewService(repoMock, mocks.NewMockStorageClient(t), zap.NewNop())
			err := s.ApproveCloseRequest(ctx, "1", "reviewer-1")
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCourseCycleCloseRequestService_RejectCloseRequest(t *testing.T) {
	type testCase struct {
		name    string
		prepare func(repoMock *mocks.MockRepository)
		wantErr error
	}

	tests := []testCase{
		{
			name: "success",
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetCourseCycleCloseRequestByID(mock.Anything, "1").
					Return(entities.CourseCycleCloseRequest{ID: 1, CourseCycleID: 2, Status: entities.RequestStatus_UNDER_REVIEW}, nil)
				repoMock.EXPECT().RejectCourseCycleCloseRequest(mock.Anything, "1", "reviewer-1", "no cumple", "2").Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "error request already processed",
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetCourseCycleCloseRequestByID(mock.Anything, "1").
					Return(entities.CourseCycleCloseRequest{ID: 1, Status: entities.RequestStatus_REJECTED}, nil)
			},
			wantErr: ErrRequestIsProcessed,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			if tt.prepare != nil {
				tt.prepare(repoMock)
			}
			s := NewService(repoMock, mocks.NewMockStorageClient(t), zap.NewNop())
			err := s.RejectCloseRequest(ctx, "1", "reviewer-1", "no cumple")
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCourseCycleCloseRequestService_GetCloseRequests(t *testing.T) {
	ctx := context.Background()
	repoMock := mocks.NewMockRepository(t)
	repoMock.EXPECT().GetCourseCycleCloseRequests(mock.Anything, entities.Faculty(""), entities.PageScope{}).
		Return([]entities.CourseCycleCloseRequest{{ID: 1}}, entities.PageScope{}, nil)
	s := NewService(repoMock, mocks.NewMockStorageClient(t), zap.NewNop())
	got, _, err := s.GetCloseRequests(ctx, "", entities.PageScope{})
	assert.NoError(t, err)
	assert.Equal(t, []entities.CourseCycleCloseRequest{{ID: 1}}, got)
}

func TestCourseCycleCloseRequestService_GetCloseRequestByID(t *testing.T) {
	ctx := context.Background()
	repoMock := mocks.NewMockRepository(t)
	repoMock.EXPECT().GetCourseCycleCloseRequestByID(mock.Anything, "1").
		Return(entities.CourseCycleCloseRequest{ID: 1}, nil)
	s := NewService(repoMock, mocks.NewMockStorageClient(t), zap.NewNop())
	got, err := s.GetCloseRequestByID(ctx, "1")
	assert.NoError(t, err)
	assert.Equal(t, entities.CourseCycleCloseRequest{ID: 1}, got)
}
