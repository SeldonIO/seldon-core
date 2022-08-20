package tls

import (
	"crypto/tls"
	"crypto/x509"
)

const (
	DefaultKeyName = "tls.key"
	DefaultCrtName = "tls.crt"
	DefaultCaName  = "ca.crt"
)

type Certificate struct {
	Certificate *tls.Certificate
	Ca          *x509.CertPool
}

type UpdateCertificateHandler interface {
	UpdateCertificate(cert *Certificate)
}

type CertificateManager interface {
	GetCertificateAndWatch(updater UpdateCertificateHandler) (*Certificate, error)
	Stop()
}
