/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package xdscache

import (
	"crypto/tls"
	"testing"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/anypb"

	seldontls "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"
)

type FakeCertificateStore struct {
	cert       *seldontls.CertificateWrapper
	validation *seldontls.CertificateWrapper
}

func (f FakeCertificateStore) GetServerCertificate(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
	// TODO implement me
	panic("implement me")
}

func (f FakeCertificateStore) GetClientCertificate(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
	// TODO implement me
	panic("implement me")
}

func (f FakeCertificateStore) CreateClientTLSConfig() *tls.Config {
	// TODO implement me
	panic("implement me")
}

func (f FakeCertificateStore) CreateClientTransportCredentials() credentials.TransportCredentials {
	// TODO implement me
	panic("implement me")
}

func (f FakeCertificateStore) CreateServerTLSConfig() *tls.Config {
	// TODO implement me
	panic("implement me")
}

func (f FakeCertificateStore) CreateServerTransportCredentials() credentials.TransportCredentials {
	// TODO implement me
	panic("implement me")
}

func (f FakeCertificateStore) GetCertificate() *seldontls.CertificateWrapper {
	return f.cert
}

func (f FakeCertificateStore) GetValidationCertificate() *seldontls.CertificateWrapper {
	return f.validation
}

func (f FakeCertificateStore) Stop() {
}

func NewFakeCertificateStore(addCert bool, addValidation bool) *FakeCertificateStore {
	var cert, validation *seldontls.CertificateWrapper
	if addCert {
		cert = &seldontls.CertificateWrapper{}
	}
	if addValidation {
		validation = &seldontls.CertificateWrapper{}
	}
	return &FakeCertificateStore{
		cert:       cert,
		validation: validation,
	}
}

func TestMakeSecretResource(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name               string
		certName           string
		validationName     string
		certStore          seldontls.CertificateStoreHandler
		expectedNumSecrets int
	}

	tests := []test{
		{
			name:               "cert only",
			certName:           "sec",
			validationName:     "val",
			certStore:          NewFakeCertificateStore(true, false),
			expectedNumSecrets: 1,
		},
		{
			name:               "cert and validation",
			certName:           "sec",
			validationName:     "val",
			certStore:          NewFakeCertificateStore(true, true),
			expectedNumSecrets: 2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			secs := MakeSecretResource(test.certName, test.validationName, test.certStore)
			g.Expect(len(secs)).To(Equal(test.expectedNumSecrets))
		})
	}
}

func TestCreateDownstreamTransportSocket(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		secret   *Secret
		expected *tlsv3.DownstreamTlsContext
	}

	tests := []test{
		{
			name:     "No certs",
			secret:   nil,
			expected: nil,
		},
		{
			name: "cert only",
			secret: &Secret{
				Name:                 "sec",
				ValidationSecretName: "val",
				Certificate:          NewFakeCertificateStore(true, false),
			},
			expected: &tlsv3.DownstreamTlsContext{
				CommonTlsContext: &tlsv3.CommonTlsContext{
					TlsCertificateSdsSecretConfigs: []*tlsv3.SdsSecretConfig{
						{
							Name:      "sec",
							SdsConfig: configSource,
						},
					},
				},
			},
		},
		{
			name: "cert and validation",
			secret: &Secret{
				Name:                 "sec",
				ValidationSecretName: "val",
				Certificate:          NewFakeCertificateStore(true, true),
			},
			expected: &tlsv3.DownstreamTlsContext{
				CommonTlsContext: &tlsv3.CommonTlsContext{
					TlsCertificateSdsSecretConfigs: []*tlsv3.SdsSecretConfig{
						{
							Name:      "sec",
							SdsConfig: configSource,
						},
					},
					ValidationContextType: &tlsv3.CommonTlsContext_ValidationContextSdsSecretConfig{
						ValidationContextSdsSecretConfig: &tlsv3.SdsSecretConfig{
							Name:      "val",
							SdsConfig: configSource,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := createDownstreamTransportSocket(test.secret)
			if test.expected != nil {
				tlsCtxPb, err := anypb.New(test.expected)
				if err != nil {
					panic(err)
				}

				tsExpected := &core.TransportSocket{
					Name: "envoy.transport_sockets.tls",
					ConfigType: &core.TransportSocket_TypedConfig{
						TypedConfig: tlsCtxPb,
					},
				}
				g.Expect(ts).To(Equal(tsExpected))
			} else {
				g.Expect(ts).To(BeNil())
			}
		})
	}
}

func TestCreateUpstreamTransportSocket(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		secret   *Secret
		expected *tlsv3.UpstreamTlsContext
	}

	tests := []test{
		{
			name:     "No certs",
			secret:   nil,
			expected: nil,
		},
		{
			name: "cert only",
			secret: &Secret{
				Name:                 "sec",
				ValidationSecretName: "val",
				Certificate:          NewFakeCertificateStore(true, false),
			},
			expected: &tlsv3.UpstreamTlsContext{
				CommonTlsContext: &tlsv3.CommonTlsContext{
					TlsCertificateSdsSecretConfigs: []*tlsv3.SdsSecretConfig{
						{
							Name:      "sec",
							SdsConfig: configSource,
						},
					},
				},
			},
		},
		{
			name: "cert and validation",
			secret: &Secret{
				Name:                 "sec",
				ValidationSecretName: "val",
				Certificate:          NewFakeCertificateStore(true, true),
			},
			expected: &tlsv3.UpstreamTlsContext{
				CommonTlsContext: &tlsv3.CommonTlsContext{
					TlsCertificateSdsSecretConfigs: []*tlsv3.SdsSecretConfig{
						{
							Name:      "sec",
							SdsConfig: configSource,
						},
					},
					ValidationContextType: &tlsv3.CommonTlsContext_ValidationContextSdsSecretConfig{
						ValidationContextSdsSecretConfig: &tlsv3.SdsSecretConfig{
							Name:      "val",
							SdsConfig: configSource,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := createUpstreamTransportSocket(test.secret)
			if test.expected != nil {
				tlsCtxPb, err := anypb.New(test.expected)
				if err != nil {
					panic(err)
				}

				tsExpected := &core.TransportSocket{
					Name: "envoy.transport_sockets.tls",
					ConfigType: &core.TransportSocket_TypedConfig{
						TypedConfig: tlsCtxPb,
					},
				}
				g.Expect(ts).To(Equal(tsExpected))
			} else {
				g.Expect(ts).To(BeNil())
			}
		})
	}
}
