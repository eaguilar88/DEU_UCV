package course_periods

import (
	"github.com/eaguilar88/deu/internal/entities"
)

type AnnouncementInfo struct {
	ID        string `json:"id"`
	Title     string `json:"titulo"`
	Content   string `json:"contenido"`
	CreatedAt string `json:"creado_el"`
	UpdatedAt string `json:"actualizado_el"`
}

type GetCoursePeriodResponse struct {
	ID              string             `json:"id"`
	Announcements   []AnnouncementInfo `json:"publicaciones,omitempty"`
	StartDate       string             `json:"fecha_inicio"`
	EndDate         string             `json:"fecha_fin"`
	InscriptionDate string             `json:"fecha_inscripcion"`
	CreatedAt       string             `json:"creado_el"`
	UpdatedAt       string             `json:"actualizado_el"`
	DeletedAt       string             `json:"eliminado_en,omitempty"`
}

// periodToResponse converts a CoursePeriod entity to GetCoursePeriodResponse.
func periodToResponse(coursePeriod entities.CoursePeriod) GetCoursePeriodResponse {
	announcements := make([]AnnouncementInfo, 0, len(coursePeriod.Announcements))
	for _, announcement := range coursePeriod.Announcements {
		announcements = append(announcements, AnnouncementInfo{
			ID:        announcement.ID,
			Title:     announcement.Title,
			Content:   announcement.Content,
			CreatedAt: announcement.CreatedAt,
			UpdatedAt: announcement.UpdatedAt,
		})
	}

	return GetCoursePeriodResponse{
		ID:              coursePeriod.ID,
		Announcements:   announcements,
		StartDate:       coursePeriod.StartDate,
		EndDate:         coursePeriod.EndDate,
		InscriptionDate: coursePeriod.InscriptionDate,
		CreatedAt:       coursePeriod.CreatedAt,
		UpdatedAt:       coursePeriod.UpdatedAt,
		DeletedAt:       coursePeriod.DeletedAt,
	}
}

type GetCoursePeriodsResponse struct {
	Periods []GetCoursePeriodResponse `json:"periodos"`
	Pages   entities.PageScope        `json:"paginas"`
}

// periodsToResponse converts a slice of CoursePeriod entities to responses.
func periodsToResponse(coursePeriods []entities.CoursePeriod) []GetCoursePeriodResponse {
	out := make([]GetCoursePeriodResponse, 0, len(coursePeriods))
	for _, coursePeriod := range coursePeriods {
		out = append(out, periodToResponse(coursePeriod))
	}
	return out
}

type CreateCoursePeriodResponse struct {
	ID string `json:"id"`
}

type UpdateCoursePeriodResponse struct{}

type DeleteCoursePeriodResponse struct{}

// Announcement responses
type GetAnnouncementResponse struct {
	ID        string `json:"id"`
	Title     string `json:"titulo"`
	Content   string `json:"contenido"`
	CreatedAt string `json:"creado_el"`
	UpdatedAt string `json:"actualizado_el"`
}

// announcementToResponse converts an Announcement entity to GetAnnouncementResponse.
func announcementToResponse(announcement entities.Announcement) GetAnnouncementResponse {
	return GetAnnouncementResponse{
		ID:        announcement.ID,
		Title:     announcement.Title,
		Content:   announcement.Content,
		CreatedAt: announcement.CreatedAt,
		UpdatedAt: announcement.UpdatedAt,
	}
}

type CreateAnnouncementResponse struct {
	ID string `json:"id"`
}

type UpdateAnnouncementResponse struct{}

type DeleteAnnouncementResponse struct{}
