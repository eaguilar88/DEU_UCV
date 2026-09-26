package files

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/eaguilar88/deu/internal/files/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestService_GetPublicFile(t *testing.T) {
	const key = "files/courses/1/portada.jpg"

	tests := []struct {
		name            string
		prepare         func(repo *mocks.MockRepository, storage *mocks.MockStorageClient)
		wantBody        string
		wantContentType string
		wantErr         bool
		errTarget       error
	}{
		{
			name: "public file is returned",
			prepare: func(repo *mocks.MockRepository, storage *mocks.MockStorageClient) {
				repo.EXPECT().IsPublicFile(mock.Anything, key).Return(true, nil)
				storage.EXPECT().GetObject(mock.Anything, key).
					Return(io.NopCloser(strings.NewReader("img")), "image/jpeg", nil)
			},
			wantBody:        "img",
			wantContentType: "image/jpeg",
		},
		{
			name: "private or unknown file is not found and never fetched",
			prepare: func(repo *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repo.EXPECT().IsPublicFile(mock.Anything, key).Return(false, nil)
			},
			wantErr:   true,
			errTarget: ErrFileNotFound,
		},
		{
			name: "repository error is returned",
			prepare: func(repo *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repo.EXPECT().IsPublicFile(mock.Anything, key).Return(false, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "storage error maps to not found",
			prepare: func(repo *mocks.MockRepository, storage *mocks.MockStorageClient) {
				repo.EXPECT().IsPublicFile(mock.Anything, key).Return(true, nil)
				storage.EXPECT().GetObject(mock.Anything, key).Return(nil, "", errors.New("b2 error"))
			},
			wantErr:   true,
			errTarget: ErrFileNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockRepository(t)
			storage := mocks.NewMockStorageClient(t)
			tt.prepare(repo, storage)

			svc := NewService(repo, storage, zap.NewNop())
			body, contentType, err := svc.GetPublicFile(context.Background(), key)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errTarget != nil {
					assert.ErrorIs(t, err, tt.errTarget)
				}
				assert.Nil(t, body)
				return
			}
			require.NoError(t, err)
			defer body.Close()
			data, err := io.ReadAll(body)
			require.NoError(t, err)
			assert.Equal(t, tt.wantBody, string(data))
			assert.Equal(t, tt.wantContentType, contentType)
		})
	}
}
