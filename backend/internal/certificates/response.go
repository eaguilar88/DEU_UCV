package certificates

// verificationPage fills templates/verification.html. The zero value renders the "not valid" page.
type verificationPage struct {
	Valid            bool
	FullName         string
	MaskedDocument   string
	CourseName       string
	Duration         string
	Faculty          string
	StartDate        string
	EndDate          string
	IssueDate        string
	VerificationCode string
	PDFURL           string
}
