package models

import (
	"database/sql"
	"time"
)

type GroupRenewal struct {
	ID          int64
	GroupID     string
	Payload     []byte
	SubmittedBy sql.NullString
	Status      string
	CreatedAt   time.Time
}

// GroupRenewalPayload is the proposed group data stored as JSON in deu.group_renewals.payload.
// Files are not part of it: they are stored in deu.files with owner_type 'group_renewal'.
type GroupRenewalPayload struct {
	Name                string                      `json:"name"`
	Description         string                      `json:"description"`
	Foundation          string                      `json:"foundation"`
	IsMultidisciplinary bool                        `json:"is_multidisciplinary"`
	Faculty             []string                    `json:"faculty"`
	Type                []string                    `json:"type"`
	Objective           string                      `json:"objective"`
	Location            string                      `json:"location"`
	Email               string                      `json:"email"`
	Phone               string                      `json:"phone"`
	Members             []GroupRenewalPayloadMember `json:"members"`
}

type GroupRenewalPayloadMember struct {
	Name         string `json:"name"`
	CI           int    `json:"ci"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Coordination string `json:"coordination"`
	Year         string `json:"year"`
	Faculty      string `json:"faculty"`
	School       string `json:"school"`
	IsLeader     bool   `json:"is_leader"`
	IsActive     bool   `json:"is_active"`
}
