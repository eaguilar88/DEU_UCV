package certificates

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/eaguilar88/deu/internal/course_cycle_close_requests"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// codePattern matches verification codes and certificates tokens: 128 bits in unpadded base32.
var codePattern = regexp.MustCompile(`^[A-Z2-7]{26}$`)

type Service interface {
	GenerateForCloseRequest(ctx context.Context, payload []byte) error
	VerifyCertificate(ctx context.Context, code string) (entities.CertificateDetails, error)
	GetCertificatePDF(ctx context.Context, code string) (io.ReadCloser, error)
	GetCertificatesZip(ctx context.Context, token string) (io.ReadCloser, error)
	RetryCertificates(ctx context.Context, closeRequestID string) error
}

type Handler struct {
	svc Service
	log *zap.Logger
}

func NewHandler(svc Service, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// RegisterPublicEndpoints registers the pages reached from a certificate's QR code and from the
// links emailed to the submitter. They are public: the random code or token is the credential.
func (h *Handler) RegisterPublicEndpoints(e *echo.Echo) {
	g := e.Group("/certificados")
	g.GET("/lotes/:token/zip", h.DownloadCertificatesZip)
	g.GET("/:code", h.VerifyCertificate)
	g.GET("/:code/pdf", h.DownloadCertificate)
}

func (h *Handler) RegisterAdminEndpoints(g *echo.Group) {
	g.POST("/course-cycle-close-requests/:id/certificates/retry", h.RetryCertificates)
}

// VerifyCertificate renders the public verification page. Unknown, unissued and revoked
// certificates get the same "not valid" page.
func (h *Handler) VerifyCertificate(c echo.Context) error {
	code := normalizeCode(c.Param("code"))
	c.Response().Header().Set(echo.HeaderCacheControl, "no-store")

	if !codePattern.MatchString(code) {
		return h.renderVerification(c, http.StatusNotFound, verificationPage{})
	}

	details, err := h.svc.VerifyCertificate(c.Request().Context(), code)
	if err != nil {
		if errors.Is(err, ErrCertificateNotFound) {
			return h.renderVerification(c, http.StatusNotFound, verificationPage{})
		}
		return httperrors.NewInternal(err)
	}

	return h.renderVerification(c, http.StatusOK, toVerificationPage(details, c.Request().URL.Path+"/pdf"))
}

func (h *Handler) DownloadCertificate(c echo.Context) error {
	code := normalizeCode(c.Param("code"))
	if !codePattern.MatchString(code) {
		return httperrors.NewNotFound("certificado no encontrado")
	}

	body, err := h.svc.GetCertificatePDF(c.Request().Context(), code)
	if err != nil {
		if errors.Is(err, ErrCertificateNotFound) {
			return httperrors.NewNotFound("certificado no encontrado")
		}
		return httperrors.NewInternal(err)
	}
	defer h.closeBody(body)

	c.Response().Header().Set(echo.HeaderCacheControl, "no-store")
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%q", "certificado-"+code+".pdf"))
	return c.Stream(http.StatusOK, pdfContentType, body)
}

func (h *Handler) DownloadCertificatesZip(c echo.Context) error {
	token := normalizeCode(c.Param("token"))
	if !codePattern.MatchString(token) {
		return httperrors.NewNotFound("certificados no encontrados")
	}

	body, err := h.svc.GetCertificatesZip(c.Request().Context(), token)
	if err != nil {
		switch {
		case errors.Is(err, course_cycle_close_requests.ErrCycleCloseRequestNotFound):
			return httperrors.NewNotFound("certificados no encontrados")
		case errors.Is(err, ErrCertificatesNotAvailable):
			return httperrors.NewNotFound(ErrCertificatesNotAvailable.Error())
		}
		return httperrors.NewInternal(err)
	}
	defer h.closeBody(body)

	c.Response().Header().Set(echo.HeaderCacheControl, "no-store")
	c.Response().Header().Set(echo.HeaderContentDisposition, `attachment; filename="certificados.zip"`)
	return c.Stream(http.StatusOK, zipContentType, body)
}

func (h *Handler) RetryCertificates(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return httperrors.NewBadRequest("request ID is required")
	}

	if err := h.svc.RetryCertificates(c.Request().Context(), id); err != nil {
		if errors.Is(err, ErrNoFailedCertificatesJob) {
			return httperrors.NewConflict(ErrNoFailedCertificatesJob.Error())
		}
		return httperrors.NewInternal(err)
	}
	return c.NoContent(http.StatusAccepted)
}

func (h *Handler) renderVerification(c echo.Context, status int, page verificationPage) error {
	var buf bytes.Buffer
	if err := verificationTemplate.Execute(&buf, page); err != nil {
		return httperrors.NewInternal(err)
	}
	return c.HTMLBlob(status, buf.Bytes())
}

func (h *Handler) closeBody(body io.Closer) {
	if err := body.Close(); err != nil {
		h.log.Warn("failed to close certificate body", zap.Error(err))
	}
}

// normalizeCode accepts codes typed by hand in lower case or with surrounding spaces.
func normalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}
