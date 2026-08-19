package group_requests_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/group_requests"
	"github.com/eaguilar88/deu/internal/group_requests/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newRequestContext(faculty, queryFaculty string) echo.Context {
	req := httptest.NewRequest("GET", "/group-requests?faculty="+queryFaculty, nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.Set("faculty", faculty)
	return c
}

func TestGroupRequestEndpointsHandler_GetGroupRequestsByFaculty(t *testing.T) {
	t.Run("faculty_admin is forced to their own faculty despite a different query param", func(t *testing.T) {
		svc := &mocks.MockService{}
		svc.On("GetGroupRequestsByFaculty", context.Background(), entities.FacultyCiencias, "", entities.PageScope{Page: 1, PerPage: 10}).
			Return([]entities.GroupRequest{}, entities.PageScope{}, 0, nil)

		h := group_requests.NewHandler(svc, zap.NewNop())
		c := newRequestContext("Ciencias", "Medicina")

		err := h.GetGroupRequestsByFaculty(c)

		require.NoError(t, err)
		svc.AssertExpectations(t)
	})

	t.Run("DEU admin can filter via query param", func(t *testing.T) {
		svc := &mocks.MockService{}
		svc.On("GetGroupRequestsByFaculty", context.Background(), entities.FacultyMedicina, "", entities.PageScope{Page: 1, PerPage: 10}).
			Return([]entities.GroupRequest{}, entities.PageScope{}, 0, nil)

		h := group_requests.NewHandler(svc, zap.NewNop())
		c := newRequestContext("DEU", "Medicina")

		err := h.GetGroupRequestsByFaculty(c)

		require.NoError(t, err)
		svc.AssertExpectations(t)
	})
}

func TestGroupRequestEndpointsHandler_GetPendingGroupRequestsCounts(t *testing.T) {
	t.Run("faculty_admin only ever gets their own faculty's counts", func(t *testing.T) {
		svc := &mocks.MockService{}
		svc.On("GetPendingGroupRequestsCounts", context.Background(), entities.FacultyCiencias).
			Return([]entities.FacultyPendingCount{{Faculty: entities.FacultyCiencias, Count: 3}}, nil)

		h := group_requests.NewHandler(svc, zap.NewNop())
		c := newRequestContext("Ciencias", "Medicina")

		err := h.GetPendingGroupRequestsCounts(c)

		require.NoError(t, err)
		svc.AssertExpectations(t)
	})

	t.Run("DEU admin without a faculty filter sees all faculties", func(t *testing.T) {
		svc := &mocks.MockService{}
		svc.On("GetPendingGroupRequestsCounts", context.Background(), entities.Faculty("")).
			Return([]entities.FacultyPendingCount{
				{Faculty: entities.FacultyCiencias, Count: 3},
				{Faculty: entities.FacultyMedicina, Count: 1},
			}, nil)

		h := group_requests.NewHandler(svc, zap.NewNop())
		c := newRequestContext("DEU", "")

		err := h.GetPendingGroupRequestsCounts(c)

		require.NoError(t, err)
		svc.AssertExpectations(t)
	})
}

func TestGroupRequestEndpointsHandler_ApproveGroupRequest(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		svc group_requests.Service
		log *zap.Logger
		// Named input parameters for target function.
		c       echo.Context
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := group_requests.NewHandler(tt.svc, tt.log)
			gotErr := h.ApproveGroupRequest(tt.c)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ApproveGroupRequest() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ApproveGroupRequest() succeeded unexpectedly")
			}
		})
	}
}
