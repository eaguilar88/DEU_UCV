package entities

import "time"

// Participant is an approved participant listed in a close request's participants file.
type Participant struct {
	FirstName string
	LastName  string
	Document  string
	Email     string
}

// Certificate is issued to an approved participant when their course cycle's close request is
// approved. It is valid once IssuedAt is set and as long as RevokedAt is not.
type Certificate struct {
	ID               int64
	CloseRequestID   int64
	CourseCycleID    int64
	FirstName        string
	LastName         string
	Document         string
	Email            string
	VerificationCode string
	StorageKey       string
	IssuedAt         *time.Time
	RevokedAt        *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (c Certificate) FullName() string {
	return c.FirstName + " " + c.LastName
}

// IsValid reports whether the certificate has been issued and not revoked.
func (c Certificate) IsValid() bool {
	return c.IssuedAt != nil && c.RevokedAt == nil
}

// CertificateDetails is a certificate together with the course cycle it certifies.
type CertificateDetails struct {
	Certificate Certificate
	Course      Course
	Period      CoursePeriod
}
