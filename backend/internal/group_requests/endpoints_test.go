package group_requests_test

import (
	"testing"

	"github.com/eaguilar88/deu/internal/group_requests"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

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
