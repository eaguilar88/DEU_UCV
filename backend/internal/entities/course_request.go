package entities

// CourseRequestFileTypeEvaluation is the evaluation document attached when approving a request.
const CourseRequestFileTypeEvaluation = "archivo_evaluacion"

type CourseRequest struct {
	ID         string
	User       User // The course owner (the provider's user account)
	Reviewer   User
	Status     RequestStatus
	Comments   string
	ReviewedAt string
	CreatedAt  string
	UpdatedAt  string
	Course     *Course // Associated course information

	// Evaluation recorded on approval. Score is nil when no score was given.
	Score          *float64
	Classification string
	EvaluationFile *File
}
