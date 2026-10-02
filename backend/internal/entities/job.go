package entities

import (
	"encoding/json"
	"time"
)

type JobStatus string

const (
	JobStatusPending JobStatus = "pending"
	JobStatusRunning JobStatus = "running"
	JobStatusDone    JobStatus = "done"
	JobStatusFailed  JobStatus = "failed"
)

// JobKindCourseCycleCertificates generates and delivers the certificates of an approved close request.
const JobKindCourseCycleCertificates = "course_cycle_certificates"

// Job is a unit of background work stored in the jobs queue.
type Job struct {
	ID          int64
	Kind        string
	Payload     json.RawMessage
	Status      JobStatus
	Attempts    int
	MaxAttempts int
	RunAt       time.Time
	LastError   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CourseCycleCertificatesPayload is the payload of a JobKindCourseCycleCertificates job.
type CourseCycleCertificatesPayload struct {
	CloseRequestID int64 `json:"close_request_id"`
}
