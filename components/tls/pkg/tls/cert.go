package tls

import (
	"crypto/tls"
	"crypto/x509"
)

type CertificateWrapper struct {
	Certificate *tls.Certificate
	Ca          *x509.CertPool
	KeyPath     string
	CrtPath     string
	CaPath      string
}

type UpdateCertificateHandler interface {
	UpdateCertificate(cert *CertificateWrapper)
}

type CertificateManager interface {
	GetCertificateAndWatch(updater UpdateCertificateHandler) (*CertificateWrapper, error)
	Stop()
}
