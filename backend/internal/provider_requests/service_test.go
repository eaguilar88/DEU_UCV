package provider_requests

import (
	"context"
	"errors"
	"testing"

	"github.com/eaguilar88/deu/internal/email"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/provider_requests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestService_RejectProviderRequest(t *testing.T) {
	pending := entities.ProviderRequest{ID: 1, ProviderID: 7, Status: entities.RequestStatus_UNDER_REVIEW}
	provider := entities.Provider{
		ID:   "7",
		Name: "ACME",
		User: entities.User{Email: "user@test.com"},
	}

	tests := []struct {
		name    string
		prepare func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient)
		wantErr error
	}{
		{
			name: "rejects and notifies the provider",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetProviderRequestByID(mock.Anything, "1").Return(pending, nil)
				repoMock.EXPECT().GetProvider(mock.Anything, "7").Return(provider, nil)
				repoMock.EXPECT().RejectProviderRequest(mock.Anything, "1", "reviewer-1", "faltan documentos").Return(nil)
				mailMock.EXPECT().SendTemplate(mock.Anything, "user@test.com", email.TemplateProviderRejected,
					email.ProviderRejectedData{ProviderName: "ACME", Reason: "faltan documentos"}).Return(nil)
			},
		},
		{
			name: "uses the user's name when the provider has none",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				natural := entities.Provider{ID: "7", User: entities.User{Email: "user@test.com", FirstName: "Ana", LastName: "Pérez"}}
				repoMock.EXPECT().GetProviderRequestByID(mock.Anything, "1").Return(pending, nil)
				repoMock.EXPECT().GetProvider(mock.Anything, "7").Return(natural, nil)
				repoMock.EXPECT().RejectProviderRequest(mock.Anything, "1", "reviewer-1", "faltan documentos").Return(nil)
				mailMock.EXPECT().SendTemplate(mock.Anything, "user@test.com", email.TemplateProviderRejected,
					email.ProviderRejectedData{ProviderName: "Ana Pérez", Reason: "faltan documentos"}).Return(nil)
			},
		},
		{
			name: "a failing email does not fail the rejection",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetProviderRequestByID(mock.Anything, "1").Return(pending, nil)
				repoMock.EXPECT().GetProvider(mock.Anything, "7").Return(provider, nil)
				repoMock.EXPECT().RejectProviderRequest(mock.Anything, "1", "reviewer-1", "faltan documentos").Return(nil)
				mailMock.EXPECT().SendTemplate(mock.Anything, "user@test.com", email.TemplateProviderRejected, mock.Anything).
					Return(errors.New("smtp down"))
			},
		},
		{
			name: "an already processed request is not rejected nor notified",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				processed := pending
				processed.Status = entities.RequestStatus_APPROVED
				repoMock.EXPECT().GetProviderRequestByID(mock.Anything, "1").Return(processed, nil)
			},
			wantErr: ErrRequestIsProcessed,
		},
		{
			name: "a repository error is returned and nothing is sent",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetProviderRequestByID(mock.Anything, "1").Return(pending, nil)
				repoMock.EXPECT().GetProvider(mock.Anything, "7").Return(provider, nil)
				repoMock.EXPECT().RejectProviderRequest(mock.Anything, "1", "reviewer-1", "faltan documentos").Return(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			mailMock := mocks.NewMockMailClient(t)
			tt.prepare(repoMock, mailMock)

			s := NewService(repoMock, mailMock, zap.NewNop())
			err := s.RejectProviderRequest(context.Background(), "1", "reviewer-1", "faltan documentos")

			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}
