/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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
