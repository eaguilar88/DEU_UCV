package groups

import (
	"context"
	"errors"
	"testing"

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
			s := NewService(repoMock, storageMock, loggerMock)
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
			s := NewService(repoMock, storageMock, loggerMock)
			got, _, err := s.GetGroups(ctx, tt.filter, tt.pageScope)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
