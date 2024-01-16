/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

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
