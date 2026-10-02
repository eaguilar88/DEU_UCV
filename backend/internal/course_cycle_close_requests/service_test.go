package course_cycle_close_requests

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/eaguilar88/deu/internal/course_cycle_close_requests/mocks"
	"github.com/eaguilar88/deu/internal/email"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestNewService(t *testing.T) {
	repoMock := mocks.NewMockRepository(t)
	storageMock := mocks.NewMockStorageClient(t)
	got := NewService(repoMock, storageMock, mocks.NewMockMailClient(t), zap.NewNop())
	assert.NotNil(t, got)
}

func newValidCloseRequest() entities.CourseCycleCloseRequest {
	participants := buildParticipantsXLSX(participantsHeaders, []string{"Ana María", "Pérez Gómez", "12.345.678", "ana@example.com"})
	return entities.CourseCycleCloseRequest{
		CourseCycleID:    1,
		Comments:         "cierre de cohorte",
		ParticipantsFile: &entities.File{Name: "archivo_participantes.xlsx", Body: bytes.NewReader(participants)},
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
				repoMock.EXPECT().CreateCourseCycleCloseRequest(mock.Anything, int64(1), int64(1), mock.MatchedBy(isParsedCertificate)).
					Return(int64(10), nil)
				storageMock.EXPECT().UploadFile(mock.Anything, mock.MatchedBy(func(files []*entities.File) bool {
					// The participants file was read to parse it; its body must still be uploadable.
					data, err := io.ReadAll(files[0].Body)
					return len(files) == 3 && err == nil && len(data) > 0
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
			name: "error invalid participants file does not create the request",
			request: func() entities.CourseCycleCloseRequest {
				r := newValidCloseRequest()
				r.ParticipantsFile = &entities.File{Name: "archivo_participantes.pdf", Body: bytes.NewReader([]byte("%PDF"))}
				return r
			}(),
			submittedByID: "1",
			want:          -1,
			wantErr:       errors.New("archivo de participantes: debe ser un archivo .xlsx generado a partir de la plantilla"),
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
				repoMock.EXPECT().CreateCourseCycleCloseRequest(mock.Anything, int64(1), int64(1), mock.Anything).
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
				repoMock.EXPECT().CreateCourseCycleCloseRequest(mock.Anything, int64(1), int64(1), mock.Anything).Return(int64(10), nil)
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
				repoMock.EXPECT().CreateCourseCycleCloseRequest(mock.Anything, int64(1), int64(1), mock.Anything).Return(int64(10), nil)
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
			s := NewService(repoMock, storageMock, mocks.NewMockMailClient(t), zap.NewNop())
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

// isParsedCertificate matches the certificates created from newValidCloseRequest's participants file.
func isParsedCertificate(certs []entities.Certificate) bool {
	return len(certs) == 1 &&
		certs[0].FirstName == "Ana María" &&
		certs[0].LastName == "Pérez Gómez" &&
		certs[0].Document == "V-12345678" &&
		certs[0].Email == "ana@example.com" &&
		len(certs[0].VerificationCode) == 26
}

// expectRecipientLookups stubs the lookups used to build the notification email: the submitter
// (user "5"), the closed cycle ("2") and its course ("3").
func expectRecipientLookups(repoMock *mocks.MockRepository) {
	repoMock.EXPECT().GetUser(mock.Anything, "5").Return(&entities.User{ID: "5", Email: "submitter@test.com"}, nil)
	repoMock.EXPECT().GetCoursePeriodByID(mock.Anything, "2").
		Return(entities.CoursePeriod{ID: "2", Course: entities.Course{ID: "3"}}, nil)
	repoMock.EXPECT().GetCourse(mock.Anything, "3").Return(entities.Course{ID: "3", Name: "Curso de prueba"}, nil)
}

func pendingCloseRequest() entities.CourseCycleCloseRequest {
	return entities.CourseCycleCloseRequest{ID: 1, CourseCycleID: 2, SubmittedByID: "5", Status: entities.RequestStatus_UNDER_REVIEW}
}

func TestCourseCycleCloseRequestService_ApproveCloseRequest(t *testing.T) {
	type testCase struct {
		name    string
		prepare func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient)
		wantErr error
	}

	// runNotify mimics the repository: it runs the notify callback inside the "transaction" and
	// propagates its error, as the real implementation does before committing.
	runNotify := func(_ context.Context, _, _, _, _ string, _ entities.Job, notify func() error) error {
		return notify()
	}

	// isCertificatesJob matches the certificates job enqueued for close request 1.
	isCertificatesJob := mock.MatchedBy(func(job entities.Job) bool {
		return job.Kind == entities.JobKindCourseCycleCertificates && string(job.Payload) == `{"close_request_id":1}`
	})
	isToken := mock.MatchedBy(func(token string) bool { return len(token) == 26 })

	tests := []testCase{
		{
			name: "success sends approval email to submitter",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetCourseCycleCloseRequestByID(mock.Anything, "1").Return(pendingCloseRequest(), nil)
				expectRecipientLookups(repoMock)
				repoMock.EXPECT().ApproveCourseCycleCloseRequest(mock.Anything, "1", "reviewer-1", "2", isToken, isCertificatesJob, mock.Anything).
					RunAndReturn(runNotify)
				mailMock.EXPECT().SendTemplate(mock.Anything, "submitter@test.com", email.TemplateCourseCycleCloseApproved,
					email.CourseCycleCloseApprovedData{CourseName: "Curso de prueba"}).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "error getting close request",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetCourseCycleCloseRequestByID(mock.Anything, "1").
					Return(entities.CourseCycleCloseRequest{}, ErrCycleCloseRequestNotFound)
			},
			wantErr: ErrCycleCloseRequestNotFound,
		},
		{
			name: "error request already processed",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetCourseCycleCloseRequestByID(mock.Anything, "1").
					Return(entities.CourseCycleCloseRequest{ID: 1, Status: entities.RequestStatus_APPROVED}, nil)
			},
			wantErr: ErrRequestIsProcessed,
		},
		{
			name: "error getting submitter does not touch the request",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetCourseCycleCloseRequestByID(mock.Anything, "1").Return(pendingCloseRequest(), nil)
				repoMock.EXPECT().GetUser(mock.Anything, "5").Return(nil, errors.New("user not found"))
			},
			wantErr: errors.New("user not found"),
		},
		{
			name: "error getting course cycle does not touch the request",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetCourseCycleCloseRequestByID(mock.Anything, "1").Return(pendingCloseRequest(), nil)
				repoMock.EXPECT().GetUser(mock.Anything, "5").Return(&entities.User{ID: "5", Email: "submitter@test.com"}, nil)
				repoMock.EXPECT().GetCoursePeriodByID(mock.Anything, "2").
					Return(entities.CoursePeriod{}, errors.New("period not found"))
			},
			wantErr: errors.New("period not found"),
		},
		{
			name: "error sending email rolls back the approval",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetCourseCycleCloseRequestByID(mock.Anything, "1").Return(pendingCloseRequest(), nil)
				expectRecipientLookups(repoMock)
				repoMock.EXPECT().ApproveCourseCycleCloseRequest(mock.Anything, "1", "reviewer-1", "2", isToken, isCertificatesJob, mock.Anything).
					RunAndReturn(runNotify)
				mailMock.EXPECT().SendTemplate(mock.Anything, "submitter@test.com", email.TemplateCourseCycleCloseApproved,
					email.CourseCycleCloseApprovedData{CourseName: "Curso de prueba"}).
					Return(errors.New("smtp down"))
			},
			wantErr: errors.New("error sending close request notification email: smtp down"),
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			mailMock := mocks.NewMockMailClient(t)
			if tt.prepare != nil {
				tt.prepare(repoMock, mailMock)
			}
			s := NewService(repoMock, mocks.NewMockStorageClient(t), mailMock, zap.NewNop())
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
		prepare func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient)
		wantErr error
	}

	runNotify := func(_ context.Context, _, _, _, _ string, notify func() error) error {
		return notify()
	}

	tests := []testCase{
		{
			name: "success sends rejection email with reason to submitter",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetCourseCycleCloseRequestByID(mock.Anything, "1").Return(pendingCloseRequest(), nil)
				expectRecipientLookups(repoMock)
				repoMock.EXPECT().RejectCourseCycleCloseRequest(mock.Anything, "1", "reviewer-1", "no cumple", "2", mock.Anything).
					RunAndReturn(runNotify)
				mailMock.EXPECT().SendTemplate(mock.Anything, "submitter@test.com", email.TemplateCourseCycleCloseRejected,
					email.CourseCycleCloseRejectedData{CourseName: "Curso de prueba", Reason: "no cumple"}).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "error request already processed",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetCourseCycleCloseRequestByID(mock.Anything, "1").
					Return(entities.CourseCycleCloseRequest{ID: 1, Status: entities.RequestStatus_REJECTED}, nil)
			},
			wantErr: ErrRequestIsProcessed,
		},
		{
			name: "error getting course does not touch the request",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetCourseCycleCloseRequestByID(mock.Anything, "1").Return(pendingCloseRequest(), nil)
				repoMock.EXPECT().GetUser(mock.Anything, "5").Return(&entities.User{ID: "5", Email: "submitter@test.com"}, nil)
				repoMock.EXPECT().GetCoursePeriodByID(mock.Anything, "2").
					Return(entities.CoursePeriod{ID: "2", Course: entities.Course{ID: "3"}}, nil)
				repoMock.EXPECT().GetCourse(mock.Anything, "3").Return(entities.Course{}, errors.New("course not found"))
			},
			wantErr: errors.New("course not found"),
		},
		{
			name: "error sending email rolls back the rejection",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetCourseCycleCloseRequestByID(mock.Anything, "1").Return(pendingCloseRequest(), nil)
				expectRecipientLookups(repoMock)
				repoMock.EXPECT().RejectCourseCycleCloseRequest(mock.Anything, "1", "reviewer-1", "no cumple", "2", mock.Anything).
					RunAndReturn(runNotify)
				mailMock.EXPECT().SendTemplate(mock.Anything, "submitter@test.com", email.TemplateCourseCycleCloseRejected,
					email.CourseCycleCloseRejectedData{CourseName: "Curso de prueba", Reason: "no cumple"}).
					Return(errors.New("smtp down"))
			},
			wantErr: errors.New("error sending close request notification email: smtp down"),
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			mailMock := mocks.NewMockMailClient(t)
			if tt.prepare != nil {
				tt.prepare(repoMock, mailMock)
			}
			s := NewService(repoMock, mocks.NewMockStorageClient(t), mailMock, zap.NewNop())
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
	s := NewService(repoMock, mocks.NewMockStorageClient(t), mocks.NewMockMailClient(t), zap.NewNop())
	got, _, err := s.GetCloseRequests(ctx, "", entities.PageScope{})
	assert.NoError(t, err)
	assert.Equal(t, []entities.CourseCycleCloseRequest{{ID: 1}}, got)
}

func TestCourseCycleCloseRequestService_GetCloseRequestByID(t *testing.T) {
	ctx := context.Background()
	repoMock := mocks.NewMockRepository(t)
	storageMock := mocks.NewMockStorageClient(t)
	repoMock.EXPECT().GetCourseCycleCloseRequestByID(mock.Anything, "1").
		Return(entities.CourseCycleCloseRequest{ID: 1, CourseName: "Fotografía", CohortName: "Cohorte 1"}, nil)
	// The vouchers file is missing: it is left nil instead of failing the request.
	repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1", entities.OwnerTypeCourseCycleCloseRequest).
		Return(entities.GroupedFiles{
			entities.CloseRequestFileTypeParticipants: {{Key: "p-key"}},
			entities.CloseRequestFileTypeSurvey:       {{Key: "s-key"}},
		}, nil)
	storageMock.EXPECT().GetPresignedFileURL(mock.Anything, "p-key").Return("https://b2/p", nil)
	storageMock.EXPECT().GetPresignedFileURL(mock.Anything, "s-key").Return("https://b2/s", nil)

	s := NewService(repoMock, storageMock, mocks.NewMockMailClient(t), zap.NewNop())
	got, err := s.GetCloseRequestByID(ctx, "1")

	assert.NoError(t, err)
	assert.Equal(t, "Fotografía", got.CourseName)
	assert.Equal(t, "Cohorte 1", got.CohortName)
	if assert.NotNil(t, got.ParticipantsFile) && assert.NotNil(t, got.SurveyFile) {
		assert.Equal(t, "https://b2/p", got.ParticipantsFile.URL)
		assert.Equal(t, "https://b2/s", got.SurveyFile.URL)
	}
	assert.Nil(t, got.VouchersFile)
}

func TestCourseCycleCloseRequestService_GetCloseRequestByID_FileURLError(t *testing.T) {
	repoMock := mocks.NewMockRepository(t)
	storageMock := mocks.NewMockStorageClient(t)
	repoMock.EXPECT().GetCourseCycleCloseRequestByID(mock.Anything, "1").Return(entities.CourseCycleCloseRequest{ID: 1}, nil)
	repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1", entities.OwnerTypeCourseCycleCloseRequest).
		Return(entities.GroupedFiles{entities.CloseRequestFileTypeSurvey: {{Key: "s-key"}}}, nil)
	storageMock.EXPECT().GetPresignedFileURL(mock.Anything, "s-key").Return("", errors.New("b2 down"))

	s := NewService(repoMock, storageMock, mocks.NewMockMailClient(t), zap.NewNop())
	_, err := s.GetCloseRequestByID(context.Background(), "1")

	assert.EqualError(t, err, "b2 down")
}
