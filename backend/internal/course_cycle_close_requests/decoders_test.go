package course_cycle_close_requests

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/security"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var e *echo.Echo

func TestMain(m *testing.M) {
	e = echo.New()
	e.Validator = security.NewCustomValidator()
	os.Exit(m.Run())
}

func newMultipartCloseRequest(t *testing.T, fields map[string]string, omitFiles ...string) echo.Context {
	t.Helper()

	omit := make(map[string]bool, len(omitFiles))
	for _, f := range omitFiles {
		omit[f] = true
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, value := range fields {
		require.NoError(t, writer.WriteField(key, value))
	}

	allFiles := []string{
		entities.CloseRequestFileTypeParticipants,
		entities.CloseRequestFileTypeVouchers,
		entities.CloseRequestFileTypeSurvey,
	}
	for _, purpose := range allFiles {
		if omit[purpose] {
			continue
		}
		part, err := writer.CreateFormFile(purpose, purpose+".pdf")
		require.NoError(t, err)
		_, err = part.Write([]byte("contenido de prueba"))
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())

	req := httptest.NewRequest("POST", "/course-cycle-close-requests", body)
	req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func TestToCloseRequestEntity(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := newMultipartCloseRequest(t, map[string]string{
			"course_cycle_id": "5",
			"observaciones":   "cierre de cohorte",
		})

		got, err := toCloseRequestEntity(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(5), got.CourseCycleID)
		assert.Equal(t, "cierre de cohorte", got.Comments)
		require.NotNil(t, got.ParticipantsFile)
		require.NotNil(t, got.VouchersFile)
		require.NotNil(t, got.SurveyFile)
	})

	t.Run("error missing course_cycle_id", func(t *testing.T) {
		ctx := newMultipartCloseRequest(t, map[string]string{})
		_, err := toCloseRequestEntity(ctx)
		assert.EqualError(t, err, "course_cycle_id is required")
	})

	t.Run("error invalid course_cycle_id", func(t *testing.T) {
		ctx := newMultipartCloseRequest(t, map[string]string{"course_cycle_id": "not-a-number"})
		_, err := toCloseRequestEntity(ctx)
		assert.EqualError(t, err, "course_cycle_id must be a valid integer")
	})

	t.Run("error missing participants file", func(t *testing.T) {
		ctx := newMultipartCloseRequest(t, map[string]string{"course_cycle_id": "5"}, entities.CloseRequestFileTypeParticipants)
		_, err := toCloseRequestEntity(ctx)
		assert.EqualError(t, err, "archivo_participantes is required")
	})

	t.Run("error missing vouchers file", func(t *testing.T) {
		ctx := newMultipartCloseRequest(t, map[string]string{"course_cycle_id": "5"}, entities.CloseRequestFileTypeVouchers)
		_, err := toCloseRequestEntity(ctx)
		assert.EqualError(t, err, "archivo_vouchers is required")
	})

	t.Run("error missing survey file", func(t *testing.T) {
		ctx := newMultipartCloseRequest(t, map[string]string{"course_cycle_id": "5"}, entities.CloseRequestFileTypeSurvey)
		_, err := toCloseRequestEntity(ctx)
		assert.EqualError(t, err, "archivo_encuesta is required")
	})
}
