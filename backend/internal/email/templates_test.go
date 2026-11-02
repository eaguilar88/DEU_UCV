package email

import (
	"io/fs"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sampleData returns data for a template where every string field is set to value, or nil for
// templates that take no data.
func sampleData(name Template, value string) any {
	switch name {
	case TemplateGroupAdminCredentials:
		return GroupAdminCredentialsData{GroupName: value, Username: value, Password: value}
	case TemplateGroupRequestApproved, TemplateGroupRenewalApproved:
		return GroupRequestApprovedData{GroupName: value}
	case TemplateGroupRequestRejected:
		return GroupRequestRejectedData{GroupName: value, Reason: value}
	case TemplateGroupResourceRequestApproved:
		return GroupResourceRequestApprovedData{GroupName: value, ResourceType: value}
	case TemplateGroupResourceRequestRejected:
		return GroupResourceRequestRejectedData{GroupName: value, ResourceType: value, Reason: value}
	case TemplateProviderApproved:
		return ProviderApprovedData{ProviderName: value}
	case TemplateProviderRejected:
		return ProviderRejectedData{ProviderName: value, Reason: value}
	case TemplateProviderRegistrationSubmitted:
		return ProviderRegistrationSubmittedData{ProviderName: value, ProviderType: value, Faculty: value}
	case TemplateCourseRequestSubmittedFaculty, TemplateCourseRequestSubmittedDEU:
		return CourseRequestSubmittedData{CourseName: value, ProviderName: value, Faculty: value}
	case TemplateGroupRequestSubmittedFaculty, TemplateGroupRequestSubmittedDEU,
		TemplateGroupRenewalSubmittedFaculty, TemplateGroupRenewalSubmittedDEU:
		return GroupRequestSubmittedData{GroupName: value, Faculty: value}
	case TemplateCourseRequestApproved:
		return CourseRequestApprovedData{CourseName: value, Comments: value}
	case TemplateCourseRequestRejected:
		return CourseRequestRejectedData{CourseName: value, Reason: value}
	case TemplateCourseRequestRedirected:
		return CourseRequestRedirectedData{CourseName: value, Faculty: value, Reason: value}
	case TemplateCourseCycleCloseApproved:
		return CourseCycleCloseApprovedData{CourseName: value}
	case TemplateCourseCycleCloseRejected:
		return CourseCycleCloseRejectedData{CourseName: value, Reason: value}
	case TemplateCourseCycleCertificatesReady:
		return CourseCycleCertificatesReadyData{
			CourseName:   value,
			ZipURL:       "https://example.com/zip",
			Certificates: []CertificateLink{{Name: value, URL: "https://example.com/c"}},
		}
	case TemplateCoordinatorPendingReminder:
		return CoordinatorPendingReminderData{Faculty: value, MinDays: 3, CourseRequests: 2, GroupRequests: 1}
	case TemplateGroupRenewalReminder:
		return GroupRenewalReminderData{GroupName: value, DueDate: value}
	default:
		return nil
	}
}

func TestParseTemplates(t *testing.T) {
	templates, err := parseTemplates()
	require.NoError(t, err)
	assert.Len(t, templates, len(allTemplates))
}

func TestRender_AllTemplates(t *testing.T) {
	templates, err := parseTemplates()
	require.NoError(t, err)

	for _, name := range allTemplates {
		t.Run(string(name), func(t *testing.T) {
			data := sampleData(name, "Valor de prueba")
			subject, body, err := render(templates, name, data)
			require.NoError(t, err)

			assert.NotEmpty(t, subject)
			assert.NotContains(t, subject, "\n")
			assert.Contains(t, body, "<html")
			assert.NotContains(t, subject+body, "<no value>")
			if data != nil {
				assert.Contains(t, body, "Valor de prueba", "template data should appear in the body")
			}
		})
	}
}

// TestTemplates_MatchFiles fails when a template file has no constant in allTemplates, or a
// constant has no file.
func TestTemplates_MatchFiles(t *testing.T) {
	files, err := fs.Glob(templateFS, "templates/*.html")
	require.NoError(t, err)

	var names []string
	for _, f := range files {
		if f == layoutFile {
			continue
		}
		names = append(names, strings.TrimSuffix(path.Base(f), ".html"))
	}

	var declared []string
	for _, name := range allTemplates {
		declared = append(declared, string(name))
	}

	assert.ElementsMatch(t, declared, names)
}

// TestRender_SubjectsAreStatic enforces that subjects never interpolate data: html/template would
// HTML-escape it, putting entities like "&amp;" into the subject line.
func TestRender_SubjectsAreStatic(t *testing.T) {
	templates, err := parseTemplates()
	require.NoError(t, err)

	for _, name := range allTemplates {
		t.Run(string(name), func(t *testing.T) {
			plain, _, err := render(templates, name, sampleData(name, "plain"))
			require.NoError(t, err)
			hostile, _, err := render(templates, name, sampleData(name, `A&B <"x">`))
			require.NoError(t, err)

			assert.Equal(t, plain, hostile, "subject must not depend on template data")
		})
	}
}

func TestRender_EscapesData(t *testing.T) {
	templates, err := parseTemplates()
	require.NoError(t, err)

	_, body, err := render(templates, TemplateCourseRequestRejected, CourseRequestRejectedData{
		CourseName: "<script>alert(1)</script>",
		Reason:     "no cumple",
	})
	require.NoError(t, err)

	assert.NotContains(t, body, "<script>")
	assert.Contains(t, body, "&lt;script&gt;")
}

func TestRender_OptionalBlocks(t *testing.T) {
	templates, err := parseTemplates()
	require.NoError(t, err)

	_, body, err := render(templates, TemplateCourseRequestApproved, CourseRequestApprovedData{CourseName: "Curso"})
	require.NoError(t, err)
	assert.NotContains(t, body, "Observaciones")

	_, body, err = render(templates, TemplateCourseRequestApproved, CourseRequestApprovedData{CourseName: "Curso", Comments: "buen curso"})
	require.NoError(t, err)
	assert.Contains(t, body, "Observaciones: buen curso")
}

func TestRender_UnknownTemplate(t *testing.T) {
	templates, err := parseTemplates()
	require.NoError(t, err)

	_, _, err = render(templates, Template("does_not_exist"), nil)
	assert.EqualError(t, err, `unknown email template "does_not_exist"`)
}
