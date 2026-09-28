package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strings"
)

//go:embed templates/*.html
var templateFS embed.FS

const layoutFile = "templates/layout.html"

// Template identifies a notification email. Its value is the template file name under templates/,
// without the .html extension.
type Template string

const (
	TemplateUserWelcome                  Template = "user_welcome"
	TemplateProviderRegistrationReceived Template = "provider_registration_received"
	TemplateProviderRequestApproved      Template = "provider_request_approved"
	TemplateGroupAdminCredentials        Template = "group_admin_credentials"
	TemplateCourseRequestApproved        Template = "course_request_approved"
	TemplateCourseRequestRejected        Template = "course_request_rejected"
	TemplateCourseRequestRedirected      Template = "course_request_redirected"
	TemplateCourseCycleCloseApproved     Template = "course_cycle_close_approved"
	TemplateCourseCycleCloseRejected     Template = "course_cycle_close_rejected"
)

// allTemplates lists every template parsed at startup. A template missing from this list cannot be sent.
var allTemplates = []Template{
	TemplateUserWelcome,
	TemplateProviderRegistrationReceived,
	TemplateProviderRequestApproved,
	TemplateGroupAdminCredentials,
	TemplateCourseRequestApproved,
	TemplateCourseRequestRejected,
	TemplateCourseRequestRedirected,
	TemplateCourseCycleCloseApproved,
	TemplateCourseCycleCloseRejected,
}

// Data passed to each template. Templates without variables (user_welcome,
// provider_registration_received, provider_request_approved) take nil.

type GroupAdminCredentialsData struct {
	GroupName string
	Username  string
	Password  string
}

type CourseRequestApprovedData struct {
	CourseName string
	Comments   string
}

type CourseRequestRejectedData struct {
	CourseName string
	Reason     string
}

type CourseRequestRedirectedData struct {
	CourseName string
	Faculty    string
	Reason     string
}

type CourseCycleCloseApprovedData struct {
	CourseName string
}

type CourseCycleCloseRejectedData struct {
	CourseName string
	Reason     string
}

// parseTemplates parses every template together with the shared layout. Each email gets its own
// template set, so the "subject" and "content" blocks of different emails don't clash.
func parseTemplates() (map[Template]*template.Template, error) {
	parsed := make(map[Template]*template.Template, len(allTemplates))
	for _, name := range allTemplates {
		t, err := template.ParseFS(templateFS, layoutFile, templateFile(name))
		if err != nil {
			return nil, fmt.Errorf("parsing email template %q: %w", name, err)
		}
		parsed[name] = t
	}
	return parsed, nil
}

// render executes a parsed template and returns its subject and HTML body.
func render(templates map[Template]*template.Template, name Template, data any) (string, string, error) {
	t, ok := templates[name]
	if !ok {
		return "", "", fmt.Errorf("unknown email template %q", name)
	}

	var subject bytes.Buffer
	if err := t.ExecuteTemplate(&subject, "subject", data); err != nil {
		return "", "", fmt.Errorf("rendering subject of email template %q: %w", name, err)
	}

	var body bytes.Buffer
	if err := t.ExecuteTemplate(&body, "layout", data); err != nil {
		return "", "", fmt.Errorf("rendering body of email template %q: %w", name, err)
	}

	return strings.TrimSpace(subject.String()), body.String(), nil
}

func templateFile(name Template) string {
	return "templates/" + string(name) + ".html"
}
