package certificates

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/eaguilar88/deu/internal/certificates/mocks"
	"github.com/eaguilar88/deu/internal/course_cycle_close_requests"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newRequest(path, paramName, paramValue string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames(paramName)
	c.SetParamValues(paramValue)
	return c, rec
}

func assertStatus(t *testing.T, err error, status int) {
	t.Helper()
	var customErr httperrors.CustomError
	require.True(t, errors.As(err, &customErr), "expected a CustomError, got %v", err)
	assert.Equal(t, status, customErr.StatusCode())
}

func TestHandler_VerifyCertificate(t *testing.T) {
	issuedAt := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	details := entities.CertificateDetails{
		Certificate: entities.Certificate{FirstName: "Ana María", LastName: "Pérez Gómez", Document: "V-12345678", VerificationCode: codeAna, IssuedAt: &issuedAt},
		Course:      entities.Course{Name: "Fotografía Digital", Duration: "40 horas", Faculty: entities.FacultyCiencias},
		Period:      entities.CoursePeriod{StartDate: "2026-06-01", EndDate: "2026-07-15"},
	}

	t.Run("valid certificate", func(t *testing.T) {
		svc := mocks.NewMockService(t)
		// Codes typed in lower case are accepted.
		c, rec := newRequest("/certificados/"+strings.ToLower(codeAna), "code", strings.ToLower(codeAna))
		svc.EXPECT().VerifyCertificate(mock.Anything, codeAna).Return(details, nil)

		require.NoError(t, NewHandler(svc, zap.NewNop()).VerifyCertificate(c))

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Certificado válido")
		assert.Contains(t, body, "Ana María Pérez Gómez")
		assert.Contains(t, body, "V-12.3**.**8")
		assert.NotContains(t, body, "V-12345678")
		assert.Contains(t, body, "Facultad de Ciencias")
		assert.Contains(t, body, "1 de septiembre de 2026")
		assert.Contains(t, body, `href="/certificados/`+strings.ToLower(codeAna)+`/pdf"`)
		assert.Equal(t, "no-store", rec.Header().Get(echo.HeaderCacheControl))
	})

	t.Run("revoked, unissued or unknown certificate", func(t *testing.T) {
		svc := mocks.NewMockService(t)
		c, rec := newRequest("/certificados/"+codeAna, "code", codeAna)
		svc.EXPECT().VerifyCertificate(mock.Anything, codeAna).Return(entities.CertificateDetails{}, ErrCertificateNotFound)

		require.NoError(t, NewHandler(svc, zap.NewNop()).VerifyCertificate(c))

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "Certificado no válido")
	})

	t.Run("malformed code does not reach the service", func(t *testing.T) {
		svc := mocks.NewMockService(t)
		c, rec := newRequest("/certificados/abc", "code", "abc")

		require.NoError(t, NewHandler(svc, zap.NewNop()).VerifyCertificate(c))

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "Certificado no válido")
	})

	t.Run("service error", func(t *testing.T) {
		svc := mocks.NewMockService(t)
		c, _ := newRequest("/certificados/"+codeAna, "code", codeAna)
		svc.EXPECT().VerifyCertificate(mock.Anything, codeAna).Return(entities.CertificateDetails{}, errors.New("db down"))

		assertStatus(t, NewHandler(svc, zap.NewNop()).VerifyCertificate(c), http.StatusInternalServerError)
	})
}

func TestHandler_DownloadCertificate(t *testing.T) {
	t.Run("streams the pdf as an attachment", func(t *testing.T) {
		svc := mocks.NewMockService(t)
		c, rec := newRequest("/certificados/"+codeAna+"/pdf", "code", codeAna)
		svc.EXPECT().GetCertificatePDF(mock.Anything, codeAna).Return(io.NopCloser(strings.NewReader("%PDF")), nil)

		require.NoError(t, NewHandler(svc, zap.NewNop()).DownloadCertificate(c))

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "%PDF", rec.Body.String())
		assert.Equal(t, pdfContentType, rec.Header().Get(echo.HeaderContentType))
		assert.Equal(t, `attachment; filename="certificado-`+codeAna+`.pdf"`, rec.Header().Get(echo.HeaderContentDisposition))
	})

	t.Run("invalid certificate", func(t *testing.T) {
		svc := mocks.NewMockService(t)
		c, _ := newRequest("/certificados/"+codeAna+"/pdf", "code", codeAna)
		svc.EXPECT().GetCertificatePDF(mock.Anything, codeAna).Return(nil, ErrCertificateNotFound)

		assertStatus(t, NewHandler(svc, zap.NewNop()).DownloadCertificate(c), http.StatusNotFound)
	})
}

func TestHandler_DownloadCertificatesZip(t *testing.T) {
	t.Run("streams the zip", func(t *testing.T) {
		svc := mocks.NewMockService(t)
		c, rec := newRequest("/certificados/lotes/"+testToken+"/zip", "token", testToken)
		svc.EXPECT().GetCertificatesZip(mock.Anything, testToken).Return(io.NopCloser(strings.NewReader("PK")), nil)

		require.NoError(t, NewHandler(svc, zap.NewNop()).DownloadCertificatesZip(c))

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, zipContentType, rec.Header().Get(echo.HeaderContentType))
	})

	for name, svcErr := range map[string]error{
		"unknown token":     course_cycle_close_requests.ErrCycleCloseRequestNotFound,
		"not generated yet": ErrCertificatesNotAvailable,
	} {
		t.Run(name, func(t *testing.T) {
			svc := mocks.NewMockService(t)
			c, _ := newRequest("/certificados/lotes/"+testToken+"/zip", "token", testToken)
			svc.EXPECT().GetCertificatesZip(mock.Anything, testToken).Return(nil, svcErr)

			assertStatus(t, NewHandler(svc, zap.NewNop()).DownloadCertificatesZip(c), http.StatusNotFound)
		})
	}
}

func TestHandler_RetryCertificates(t *testing.T) {
	t.Run("accepted", func(t *testing.T) {
		svc := mocks.NewMockService(t)
		c, rec := newRequest("/admin/course-cycle-close-requests/1/certificates/retry", "id", "1")
		svc.EXPECT().RetryCertificates(mock.Anything, "1").Return(nil)

		require.NoError(t, NewHandler(svc, zap.NewNop()).RetryCertificates(c))
		assert.Equal(t, http.StatusAccepted, rec.Code)
	})

	t.Run("no failed job", func(t *testing.T) {
		svc := mocks.NewMockService(t)
		c, _ := newRequest("/admin/course-cycle-close-requests/1/certificates/retry", "id", "1")
		svc.EXPECT().RetryCertificates(mock.Anything, "1").Return(ErrNoFailedCertificatesJob)

		assertStatus(t, NewHandler(svc, zap.NewNop()).RetryCertificates(c), http.StatusConflict)
	})
}
