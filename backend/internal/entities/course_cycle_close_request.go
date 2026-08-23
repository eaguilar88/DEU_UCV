package entities

const (
	CloseRequestFileTypeParticipants = "archivo_participantes"
	CloseRequestFileTypeVouchers     = "archivo_vouchers"
	CloseRequestFileTypeSurvey       = "archivo_encuesta"
)

type CourseCycleCloseRequest struct {
	ID               int64
	CourseCycleID    int64
	SubmittedByID    string
	Status           RequestStatus
	Comments         string
	ReviewerID       string
	ReviewedAt       string
	CreatedAt        string
	UpdatedAt        string
	ParticipantsFile *File
	VouchersFile     *File
	SurveyFile       *File
}
