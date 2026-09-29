package certificates

import "errors"

var (
	ErrCertificateNotFound       = errors.New("certificate not found")
	ErrCertificatesNotAvailable  = errors.New("los certificados aún no están disponibles")
	ErrNoFailedCertificatesJob   = errors.New("no hay una generación de certificados fallida para esta solicitud")
	ErrInvalidCloseRequestStatus = errors.New("close request is not approved")
)
