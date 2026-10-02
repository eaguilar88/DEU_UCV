package course_cycle_close_requests

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/utils"
	"github.com/labstack/echo/v4"
)

func formValue(c echo.Context, key string) string {
	return strings.TrimSpace(c.FormValue(key))
}

// toCloseRequestEntity converts a multipart form request to a CourseCycleCloseRequest entity.
func toCloseRequestEntity(c echo.Context) (entities.CourseCycleCloseRequest, error) {
	cycleIDStr := formValue(c, "course_cycle_id")
	if cycleIDStr == "" {
		return entities.CourseCycleCloseRequest{}, fmt.Errorf("course_cycle_id is required")
	}
	cycleID, err := strconv.ParseInt(cycleIDStr, 10, 64)
	if err != nil {
		return entities.CourseCycleCloseRequest{}, fmt.Errorf("course_cycle_id must be a valid integer")
	}

	participants, err := utils.GetFileFrom(c, entities.CloseRequestFileTypeParticipants)
	if err != nil {
		return entities.CourseCycleCloseRequest{}, fmt.Errorf("%s is required", entities.CloseRequestFileTypeParticipants)
	}
	vouchers, err := utils.GetFileFrom(c, entities.CloseRequestFileTypeVouchers)
	if err != nil {
		return entities.CourseCycleCloseRequest{}, fmt.Errorf("%s is required", entities.CloseRequestFileTypeVouchers)
	}
	survey, err := utils.GetFileFrom(c, entities.CloseRequestFileTypeSurvey)
	if err != nil {
		return entities.CourseCycleCloseRequest{}, fmt.Errorf("%s is required", entities.CloseRequestFileTypeSurvey)
	}

	return entities.CourseCycleCloseRequest{
		CourseCycleID:    cycleID,
		Comments:         formValue(c, "observaciones"),
		ParticipantsFile: participants,
		VouchersFile:     vouchers,
		SurveyFile:       survey,
	}, nil
}
