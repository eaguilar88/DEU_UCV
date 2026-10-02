package certificates

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/skip2/go-qrcode"
)

//go:embed templates/*.html templates/assets/*.jpeg templates/assets/*.svg templates/assets/fonts/*.woff2
var templateFS embed.FS

const assetsDir = "templates/assets"

var (
	certificateTemplate  = template.Must(template.ParseFS(templateFS, "templates/certificate.html"))
	verificationTemplate = template.Must(template.ParseFS(templateFS, "templates/verification.html"))

	// Open Sans, the font of the designs, is inlined: the Gotenberg image does not ship it.
	fontRegular = mustAssetDataURI("fonts/OpenSans-Regular.woff2")
	fontBold    = mustAssetDataURI("fonts/OpenSans-Bold.woff2")
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

// certificateLayout is the design a certificate is printed on.
type certificateLayout struct {
	// background is the page-sized image in templates/assets.
	background string
	// staticText means the background already prints "Otorgado a:", the ribbon text and the
	// signature of the Director of University Extension.
	staticText bool
}

var deuLayout = certificateLayout{background: "certificado_deu.svg", staticText: true}

// layouts maps the faculty that approved the course request to its certificate design.
var layouts = map[entities.Faculty]certificateLayout{
	entities.FacultyDEU:                        deuLayout,
	entities.FacultyAgronomia:                  {background: "diploma_agronomia.jpeg"},
	entities.FacultyArquitecturaUrbanismo:      {background: "diploma_arquitectura.jpeg"},
	entities.FacultyCiencias:                   {background: "diploma_ciencias.jpeg"},
	entities.FacultyCienciasEconomicasSociales: {background: "diploma_faces.jpeg"},
	entities.FacultyCienciasJuridicasPoliticas: {background: "diploma_ciencias_juridicas.jpeg"},
	entities.FacultyCienciasVeterinarias:       {background: "diploma_veterinaria.jpeg"},
	entities.FacultyFarmacia:                   {background: "diploma_farmacia.jpeg"},
	entities.FacultyHumanidadesEducacion:       {background: "diploma_humanidades.jpeg"},
	entities.FacultyIngenieria:                 {background: "diploma_ingenieria.jpeg"},
	entities.FacultyMedicina:                   {background: "diploma_medicina.jpeg"},
	entities.FacultyOdontologia:                {background: "diploma_odontologia.jpeg"},
}

// layoutFor returns the design of the faculty that approved the course. It is course.Faculty,
// which a redirect updates, not course.OriginFaculty, which receives the close requests.
func layoutFor(f entities.Faculty) certificateLayout {
	if l, ok := layouts[f]; ok {
		return l
	}
	return deuLayout
}

// backgroundDataURI inlines an embedded background image in a data: URI.
func backgroundDataURI(file string) (template.URL, error) {
	return assetDataURI(file)
}

var assetMediaTypes = map[string]string{
	".jpeg":  "image/jpeg",
	".svg":   "image/svg+xml",
	".woff2": "font/woff2",
}

// assetDataURI inlines a file of templates/assets in a data: URI.
func assetDataURI(file string) (template.URL, error) {
	data, err := templateFS.ReadFile(path.Join(assetsDir, file))
	if err != nil {
		return "", fmt.Errorf("reading certificate asset %q: %w", file, err)
	}
	return template.URL("data:" + assetMediaTypes[path.Ext(file)] + ";base64," + base64.StdEncoding.EncodeToString(data)), nil //nolint:gosec // embedded asset, not user input
}

func mustAssetDataURI(file string) template.URL {
	uri, err := assetDataURI(file)
	if err != nil {
		panic(err)
	}
	return uri
}

// certificateData fills templates/certificate.html.
type certificateData struct {
	// Background is a data: URI; html/template would replace a plain string with "#ZgotmplZ".
	Background template.URL
	StaticText bool

	FullName   string
	Document   string
	CourseName string
	Duration   string
	Modality   string
	EndMonth   string
	EndYear    string
	// Endorsement names who endorses the course, e.g. "la Facultad de Farmacia y la Dirección de
	// Extensión Universitaria".
	Endorsement string
	// FacultyName is the faculty whose dean signs; empty for DEU courses.
	FacultyName string

	IssueDate        string
	VerificationCode string
	VerifyHost       string
	// QRCode is a data: URI; html/template would replace a plain string with "#ZgotmplZ".
	QRCode template.URL

	// Set by renderCertificate.
	FontRegular template.URL
	FontBold    template.URL
}

func renderCertificate(data certificateData) ([]byte, error) {
	data.FontRegular, data.FontBold = fontRegular, fontBold
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

// cycleMonthYear returns the Spanish month and the year of a course cycle DATE column, e.g.
// ("julio", "2026"). It returns empty strings for empty or unparseable values.
func cycleMonthYear(value string) (string, string) {
	if len(value) < len("2006-01-02") {
		return "", ""
	}
	t, err := time.Parse("2006-01-02", value[:len("2006-01-02")])
	if err != nil {
		return "", ""
	}
	return spanishMonths[t.Month()-1], fmt.Sprint(t.Year())
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

// endorsement names who endorses a course approved by f, for the certificate's text.
func endorsement(f entities.Faculty) string {
	const deu = "la Dirección de Extensión Universitaria"
	if f == "" || f == entities.FacultyDEU {
		return deu
	}
	return "la Facultad de " + f.String() + " y " + deu
}

// modalityLabel translates a course location into the modality printed on the certificate.
func modalityLabel(location string) string {
	switch location {
	case "online":
		return "en línea"
	case "in_site":
		return "presencial"
	case "mixed":
		return "mixta"
	default:
		return ""
	}
}

// durationLabel prints a bare number of hours as academic hours, e.g. "40" -> "40 horas
// académicas". Any other duration is printed as entered.
func durationLabel(duration string) string {
	duration = strings.TrimSpace(duration)
	if duration == "" || strings.Trim(duration, "0123456789") != "" {
		return duration
	}
	return duration + " horas académicas"
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
