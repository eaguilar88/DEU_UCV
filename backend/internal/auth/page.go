package auth

import (
	"bytes"
	"embed"
	"html/template"

	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/labstack/echo/v4"
)

// ResetPasswordPagePath is where the backend serves the page the password reset email links to.
const ResetPasswordPagePath = "/restablecer-contrasena"

// minPasswordLength is the shortest password accepted. The JSON DTOs repeat it in their
// validate tags (min=8).
const minPasswordLength = 8

// resetPageCSP allows only the page's inline styles and form posts to itself.
const resetPageCSP = "default-src 'none'; style-src 'unsafe-inline'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'"

//go:embed templates/*.html
var templateFS embed.FS

var resetPasswordTemplate = template.Must(template.ParseFS(templateFS, "templates/reset_password.html"))

// resetPasswordPage fills templates/reset_password.html. With Done and Invalid unset it renders
// the form.
type resetPasswordPage struct {
	Token string
	Error string
	// Done renders the "password updated" page.
	Done bool
	// Invalid renders the "invalid or expired link" page.
	Invalid bool

	// Set by renderResetPage.
	Action    string
	MinLength int
}

// renderResetPage writes the page with headers that keep the token out of caches, the Referer
// header and other sites' frames.
func renderResetPage(c echo.Context, status int, page resetPasswordPage) error {
	page.Action = ResetPasswordPagePath
	page.MinLength = minPasswordLength

	var buf bytes.Buffer
	if err := resetPasswordTemplate.Execute(&buf, page); err != nil {
		return httperrors.NewInternal(err)
	}

	h := c.Response().Header()
	h.Set(echo.HeaderCacheControl, "no-store")
	h.Set("Referrer-Policy", "no-referrer")
	h.Set(echo.HeaderXFrameOptions, "DENY")
	h.Set(echo.HeaderContentSecurityPolicy, resetPageCSP)
	return c.HTMLBlob(status, buf.Bytes())
}
