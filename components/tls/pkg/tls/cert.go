package tls

import (
	"crypto/tls"
	"crypto/x509"
	"google.golang.org/grpc/credentials"
)

type CertificateWrapper struct {
	Certificate *tls.Certificate
	Ca          *x509.CertPool
	KeyPath     string
	CrtPath     string
	CaPath      string
	KeyRaw      []byte
	CrtRaw      []byte
	CaRaw       []byte
}

type CertificateManager interface {
	GetCertificateAndWatch() error
	GetCertificate() *CertificateWrapper
	Stop()
}

type CertificateStoreHandler interface {
	GetServerCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error)
	GetClientCertificate(*tls.CertificateRequestInfo) (*tls.Certificate, error)
	CreateClientTLSConfig() *tls.Config
	CreateClientTransportCredentials() credentials.TransportCredentials
	CreateServerTLSConfig() *tls.Config
	CreateServerTransportCredentials() credentials.TransportCredentials
	GetCertificate() *CertificateWrapper
	GetValidationCertificate() *CertificateWrapper
	Stop()
}
