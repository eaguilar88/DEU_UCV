package certificates

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/url"
	"strings"
	"time"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/skip2/go-qrcode"
)

//go:embed templates/*.html
var templateFS embed.FS

var (
	certificateTemplate  = template.Must(template.ParseFS(templateFS, "templates/certificate.html"))
	verificationTemplate = template.Must(template.ParseFS(templateFS, "templates/verification.html"))
)

var spanishMonths = [...]string{
	"enero", "febrero", "marzo", "abril", "mayo", "junio",
	"julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre",
}

// caracas is the time zone certificate dates are shown in. Falls back to a fixed UTC-4 offset
// when the tz database is unavailable (e.g. a minimal container image).
var caracas = func() *time.Location {
	if loc, err := time.LoadLocation("America/Caracas"); err == nil {
		return loc
	}
	return time.FixedZone("VET", -4*60*60)
}()

// certificateData fills templates/certificate.html.
type certificateData struct {
	FullName         string
	Document         string
	CourseName       string
	Duration         string
	Faculty          string
	StartDate        string
	EndDate          string
	IssueDate        string
	VerificationCode string
	VerifyHost       string
	// QRCode is a data: URI; html/template would replace a plain string with "#ZgotmplZ".
	QRCode template.URL
}

func renderCertificate(data certificateData) ([]byte, error) {
	var buf bytes.Buffer
	if err := certificateTemplate.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("rendering certificate template: %w", err)
	}
	return buf.Bytes(), nil
}

// qrDataURI encodes content as a QR code PNG inlined in a data: URI.
func qrDataURI(content string) (template.URL, error) {
	png, err := qrcode.Encode(content, qrcode.Medium, 512)
	if err != nil {
		return "", fmt.Errorf("encoding QR code: %w", err)
	}
	return template.URL("data:image/png;base64," + base64.StdEncoding.EncodeToString(png)), nil //nolint:gosec // generated PNG, not user input
}

// spanishDate formats t as e.g. "29 de septiembre de 2026" in Caracas time.
func spanishDate(t time.Time) string {
	t = t.In(caracas)
	return fmt.Sprintf("%d de %s de %d", t.Day(), spanishMonths[t.Month()-1], t.Year())
}

// cycleDate formats a course cycle DATE column (scanned as "2026-01-15" or
// "2026-01-15T00:00:00Z") as "15/01/2026". It returns "" for empty or unparseable values.
func cycleDate(value string) string {
	if len(value) < len("2006-01-02") {
		return ""
	}
	t, err := time.Parse("2006-01-02", value[:len("2006-01-02")])
	if err != nil {
		return ""
	}
	return t.Format("02/01/2006")
}

// facultyLabel names the unit that offered the course.
func facultyLabel(f entities.Faculty) string {
	switch f {
	case "":
		return ""
	case entities.FacultyDEU:
		return "Dirección de Extensión Universitaria"
	default:
		return "Facultad de " + f.String()
	}
}

// maskDocument hides the middle digits of a cédula, e.g. "V-12345678" -> "V-12.3**.**8".
func maskDocument(document string) string {
	prefix, digits, found := strings.Cut(document, "-")
	if !found {
		prefix, digits = "", document
	}

	masked := []rune(digits)
	for i := range masked {
		if i >= 3 && i < len(masked)-1 {
			masked[i] = '*'
		}
	}

	// Group digits in thousands from the right: 12345678 -> 12.345.678.
	grouped := make([]rune, 0, len(masked)+len(masked)/3)
	for i, r := range masked {
		if i > 0 && (len(masked)-i)%3 == 0 {
			grouped = append(grouped, '.')
		}
		grouped = append(grouped, r)
	}

	if prefix == "" {
		return string(grouped)
	}
	return prefix + "-" + string(grouped)
}

// hostOf returns the host of rawURL, for a short "verify at" hint printed on the certificate.
func hostOf(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return rawURL
	}
	return u.Host
}
