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
	TemplateUserWelcome                   Template = "user_welcome"
	TemplateProviderRegistrationReceived  Template = "provider_registration_received"
	TemplateProviderRequestApproved       Template = "provider_request_approved"
	TemplateProviderApproved              Template = "provider_approved"
	TemplateProviderRejected              Template = "provider_rejected"
	TemplateProviderRegistrationSubmitted Template = "provider_registration_submitted"
	TemplateCourseRequestSubmittedFaculty Template = "course_request_submitted_faculty"
	TemplateCourseRequestSubmittedDEU     Template = "course_request_submitted_deu"
	TemplateGroupRequestSubmittedFaculty  Template = "group_request_submitted_faculty"
	TemplateGroupRequestSubmittedDEU      Template = "group_request_submitted_deu"
	TemplateGroupAdminCredentials         Template = "group_admin_credentials"
	TemplateGroupRequestRejected          Template = "group_request_rejected"
	TemplateGroupRequestApproved          Template = "group_request_approved"
	TemplateGroupResourceRequestApproved  Template = "group_resource_request_approved"
	TemplateGroupResourceRequestRejected  Template = "group_resource_request_rejected"
	TemplateCourseRequestApproved         Template = "course_request_approved"
	TemplateCourseRequestRejected         Template = "course_request_rejected"
	TemplateCourseRequestRedirected       Template = "course_request_redirected"
	TemplateCourseCycleCloseApproved      Template = "course_cycle_close_approved"
	TemplateCourseCycleCloseRejected      Template = "course_cycle_close_rejected"
	TemplateCourseCycleCertificatesReady  Template = "course_cycle_certificates_ready"
	TemplateCoordinatorPendingReminder    Template = "coordinator_pending_requests_reminder"
	TemplateGroupRenewalReminder          Template = "group_renewal_reminder"
	TemplateGroupRenewalSubmittedFaculty  Template = "group_renewal_submitted_faculty"
	TemplateGroupRenewalSubmittedDEU      Template = "group_renewal_submitted_deu"
	TemplateGroupRenewalApproved          Template = "group_renewal_approved"
)

// allTemplates lists every template parsed at startup. A template missing from this list cannot be sent.
var allTemplates = []Template{
	TemplateUserWelcome,
	TemplateProviderRegistrationReceived,
	TemplateProviderRequestApproved,
	TemplateProviderApproved,
	TemplateProviderRejected,
	TemplateProviderRegistrationSubmitted,
	TemplateCourseRequestSubmittedFaculty,
	TemplateCourseRequestSubmittedDEU,
	TemplateGroupRequestSubmittedFaculty,
	TemplateGroupRequestSubmittedDEU,
	TemplateGroupAdminCredentials,
	TemplateGroupRequestApproved,
	TemplateGroupRequestRejected,
	TemplateGroupResourceRequestApproved,
	TemplateGroupResourceRequestRejected,
	TemplateCourseRequestApproved,
	TemplateCourseRequestRejected,
	TemplateCourseRequestRedirected,
	TemplateCourseCycleCloseApproved,
	TemplateCourseCycleCloseRejected,
	TemplateCourseCycleCertificatesReady,
	TemplateCoordinatorPendingReminder,
	TemplateGroupRenewalReminder,
	TemplateGroupRenewalSubmittedFaculty,
	TemplateGroupRenewalSubmittedDEU,
	TemplateGroupRenewalApproved,
}

// Data passed to each template. Templates without variables (user_welcome,
// provider_registration_received, provider_request_approved) take nil.

type GroupAdminCredentialsData struct {
	GroupName string
	Username  string
	Password  string
}

type ProviderApprovedData struct {
	ProviderName string
}

type ProviderRejectedData struct {
	ProviderName string
	Reason       string
}

type ProviderRegistrationSubmittedData struct {
	ProviderName string
	ProviderType string
	Faculty      string
}

type CourseRequestSubmittedData struct {
	CourseName   string
	ProviderName string
	Faculty      string
}

type GroupRequestSubmittedData struct {
	GroupName string
	Faculty   string
}

type CourseRequestApprovedData struct {
	CourseName string
	Comments   string
}

type CourseRequestRejectedData struct {
	CourseName string
	Reason     string
}

type GroupRequestApprovedData struct {
	GroupName string
}
type GroupRequestRejectedData struct {
	GroupName string
	Reason    string
}

type GroupResourceRequestApprovedData struct {
	GroupName    string
	ResourceType string
}

type GroupResourceRequestRejectedData struct {
	GroupName    string
	ResourceType string
	Reason       string
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

type CourseCycleCertificatesReadyData struct {
	CourseName   string
	ZipURL       string
	Certificates []CertificateLink
}

type CoordinatorPendingReminderData struct {
	Faculty        string
	MinDays        int
	CourseRequests int
	GroupRequests  int
}

type GroupRenewalReminderData struct {
	GroupName string
	DueDate   string
	Overdue   bool
}

// CertificateLink is a participant's name and the link to their certificate's verification page.
type CertificateLink struct {
	Name string
	URL  string
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
