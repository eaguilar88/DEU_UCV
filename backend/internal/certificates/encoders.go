package certificates

import "github.com/eaguilar88/deu/internal/entities"

func toVerificationPage(d entities.CertificateDetails, pdfURL string) verificationPage {
	page := verificationPage{
		Valid:            true,
		FullName:         d.Certificate.FullName(),
		MaskedDocument:   maskDocument(d.Certificate.Document),
		CourseName:       d.Course.Name,
		Duration:         d.Course.Duration,
		Faculty:          facultyLabel(d.Course.Faculty),
		StartDate:        cycleDate(d.Period.StartDate),
		EndDate:          cycleDate(d.Period.EndDate),
		VerificationCode: d.Certificate.VerificationCode,
		PDFURL:           pdfURL,
	}
	if d.Certificate.IssuedAt != nil {
		page.IssueDate = spanishDate(*d.Certificate.IssuedAt)
	}
	return page
}
