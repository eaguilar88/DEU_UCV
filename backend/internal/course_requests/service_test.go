package course_requests

import (
	"context"
	"errors"
	"testing"

	"github.com/eaguilar88/deu/internal/course_requests/mocks"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestCourseRequestService_GetMyCourseRequests(t *testing.T) {
	type testCase struct {
		name    string
		prepare func(repoMock *mocks.MockRepository)
		want    []entities.CourseRequest
		wantErr error
	}

	tests := []testCase{
		{
			name: "success",
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetProviderByUserID(mock.Anything, "user-1").
					Return(entities.Provider{ID: "provider-1"}, nil)
				repoMock.EXPECT().GetCourseRequestsByProvider(mock.Anything, "provider-1", mock.AnythingOfType("entities.PageScope")).
					Return([]entities.CourseRequest{{ID: "1"}}, entities.PageScope{Page: 1, PerPage: 10}, nil)
			},
			want: []entities.CourseRequest{{ID: "1"}},
		},
		{
			name: "no provider record returns empty list, not an error",
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetProviderByUserID(mock.Anything, "user-1").
					Return(entities.Provider{}, providers.ErrProviderNotFound)
			},
			want: nil,
		},
		{
			name: "error resolving provider",
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetProviderByUserID(mock.Anything, "user-1").
					Return(entities.Provider{}, errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
		{
			name: "error listing requests by provider",
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetProviderByUserID(mock.Anything, "user-1").
					Return(entities.Provider{ID: "provider-1"}, nil)
				repoMock.EXPECT().GetCourseRequestsByProvider(mock.Anything, "provider-1", mock.AnythingOfType("entities.PageScope")).
					Return(nil, entities.PageScope{}, errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			if tt.prepare != nil {
				tt.prepare(repoMock)
			}
			s := NewService(repoMock, zap.NewNop())
			got, _, err := s.GetMyCourseRequests(ctx, "user-1", entities.PageScope{})
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
