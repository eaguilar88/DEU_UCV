package course_cycle_close_requests

import (
	"encoding/json"
	"testing"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloseRequestToResponse(t *testing.T) {
	t.Run("list item carries course and cohort, without files", func(t *testing.T) {
		resp := closeRequestToResponse(entities.CourseCycleCloseRequest{
			ID: 1, CourseCycleID: 2, CourseID: "3", CourseName: "Fotografía", CohortName: "Cohorte 1",
		})

		data, err := json.Marshal(resp)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"curso":{"id":"3","nombre":"Fotografía"},"nombre_cohorte":"Cohorte 1"`)
		assert.NotContains(t, string(data), `"archivos"`)
	})

	t.Run("detail carries the evidence file URLs", func(t *testing.T) {
		resp := closeRequestToResponse(entities.CourseCycleCloseRequest{
			ID:               1,
			ParticipantsFile: &entities.File{URL: "https://b2/p"},
			SurveyFile:       &entities.File{URL: "https://b2/s"},
		})

		assert.Equal(t, &CloseRequestFiles{Participants: "https://b2/p", Survey: "https://b2/s"}, resp.Files)
	})
}
