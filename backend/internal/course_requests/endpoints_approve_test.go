package course_requests

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eaguilar88/deu/internal/course_requests/mocks"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newApproveContext(req *http.Request) (echo.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("1")
	ctx.Set("userID", "reviewer-1")
	return ctx, rec
}

func TestHandler_ApproveCourseRequest_Multipart(t *testing.T) {
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	require.NoError(t, form.WriteField("tipo_curso", "life_skills"))
	require.NoError(t, form.WriteField("observaciones", "buen curso"))
	require.NoError(t, form.WriteField("calificacion", "18.5"))
	require.NoError(t, form.WriteField("clasificacion", "Formación para el trabajo"))
	part, err := form.CreateFormFile("archivo_evaluacion", "rubrica.pdf")
	require.NoError(t, err)
	_, err = part.Write([]byte("%PDF"))
	require.NoError(t, err)
	require.NoError(t, form.Close())

	req := httptest.NewRequest(http.MethodPost, "/admin/course-requests/1/approve", &body)
	req.Header.Set(echo.HeaderContentType, form.FormDataContentType())
	ctx, rec := newApproveContext(req)

	svc := mocks.NewMockService(t)
	svc.EXPECT().ApproveCourseRequest(mock.Anything, mock.MatchedBy(func(r entities.CourseRequest) bool {
		return r.ID == "1" && r.Reviewer.ID == "reviewer-1" && r.Comments == "buen curso" &&
			r.Score != nil && *r.Score == 18.5 && r.Classification == "Formación para el trabajo" &&
			r.EvaluationFile != nil && r.EvaluationFile.Name == "archivo_evaluacion.pdf"
	}), entities.CourseType_LifeSkills, mock.Anything).Return(nil)

	err = NewHandler(svc, zap.NewNop()).ApproveCourseRequest(ctx)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_ApproveCourseRequest_JSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/admin/course-requests/1/approve",
		strings.NewReader(`{"tipo_curso":"life_skills","observaciones":"ok"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	ctx, rec := newApproveContext(req)

	svc := mocks.NewMockService(t)
	svc.EXPECT().ApproveCourseRequest(mock.Anything, mock.MatchedBy(func(r entities.CourseRequest) bool {
		return r.Comments == "ok" && r.Score == nil && r.EvaluationFile == nil
	}), entities.CourseType_LifeSkills, mock.Anything).Return(nil)

	err := NewHandler(svc, zap.NewNop()).ApproveCourseRequest(ctx)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_ApproveCourseRequest_InvalidScore(t *testing.T) {
	for _, score := range []string{"abc", "-1"} {
		t.Run(score, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/admin/course-requests/1/approve",
				strings.NewReader(`{"calificacion":"`+score+`"}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			ctx, _ := newApproveContext(req)

			err := NewHandler(mocks.NewMockService(t), zap.NewNop()).ApproveCourseRequest(ctx)

			assert.Equal(t, httperrors.NewBadRequest("calificacion must be a non-negative number"), err)
		})
	}
}

func TestHandler_GetCourseRequestsByFaculty_Scope(t *testing.T) {
	tests := []struct {
		name        string
		claim       string
		query       string
		wantFaculty entities.Faculty
	}{
		{name: "DEU admin without faculty lists every faculty", claim: "DEU", wantFaculty: ""},
		{name: "DEU admin filters by faculty", claim: "DEU", query: "?faculty=Medicina", wantFaculty: entities.FacultyMedicina},
		{name: "faculty admin is always scoped to their faculty", claim: "Farmacia", query: "?faculty=Medicina", wantFaculty: entities.FacultyFarmacia},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/admin/course-requests"+tt.query, nil)
			rec := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, rec)
			ctx.Set("faculty", tt.claim)

			svc := mocks.NewMockService(t)
			svc.EXPECT().GetCourseRequestsByFaculty(mock.Anything, tt.wantFaculty, mock.Anything).
				Return([]entities.CourseRequest{{
					ID: "1", User: entities.User{ID: "owner-1"},
					Course: &entities.Course{ID: "3", Name: "Curso", HasDocumentation: true},
				}}, entities.PageScope{}, nil)

			err := NewHandler(svc, zap.NewNop()).GetCourseRequestsByFaculty(ctx)

			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
			// List items carry their course, its contract status and the owner's user.
			assert.Contains(t, rec.Body.String(), `"curso":{"id":"3","nombre":"Curso","tiene_documentacion_legal":true},"usuario_id":"owner-1"`)
		})
	}
}
