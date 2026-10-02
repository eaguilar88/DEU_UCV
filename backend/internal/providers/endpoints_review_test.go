package providers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/eaguilar88/deu/internal/providers/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func newReviewContext(body string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPost, "/admin/providers/7/reject", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("7")
	return ctx, rec
}

func TestHandler_RejectProvider_SendsReason(t *testing.T) {
	ctx, rec := newReviewContext(`{"observaciones":"  Falta el RIF vigente  "}`)
	svc := mocks.NewMockService(t)
	svc.EXPECT().RejectProvider(mock.Anything, "7", "Falta el RIF vigente").Return(nil)

	err := NewHandler(svc, zap.NewNop()).RejectProvider(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, rec.Code)
}

func TestHandler_ReviewProvider_Errors(t *testing.T) {
	tests := []struct {
		name    string
		svcErr  error
		wantErr error
	}{
		{name: "not found", svcErr: ErrProviderNotFound, wantErr: httperrors.NewNotFound("provider not found")},
		{name: "already processed", svcErr: ErrProviderAlreadyProcessed, wantErr: httperrors.NewConflict(ErrProviderAlreadyProcessed.Error())},
	}
	for _, tt := range tests {
		t.Run("approve/"+tt.name, func(t *testing.T) {
			ctx, _ := newReviewContext("")
			svc := mocks.NewMockService(t)
			svc.EXPECT().ApproveProvider(mock.Anything, "7").Return(tt.svcErr)

			err := NewHandler(svc, zap.NewNop()).ApproveProvider(ctx)

			assert.Equal(t, tt.wantErr.Error(), err.Error())
			assert.Equal(t, tt.wantErr.(httperrors.CustomError).StatusCode(), err.(httperrors.CustomError).StatusCode())
		})
		t.Run("reject/"+tt.name, func(t *testing.T) {
			ctx, _ := newReviewContext(`{}`)
			svc := mocks.NewMockService(t)
			svc.EXPECT().RejectProvider(mock.Anything, "7", "").Return(tt.svcErr)

			err := NewHandler(svc, zap.NewNop()).RejectProvider(ctx)

			assert.Equal(t, tt.wantErr.(httperrors.CustomError).StatusCode(), err.(httperrors.CustomError).StatusCode())
		})
	}
}
