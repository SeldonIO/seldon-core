/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package util

import seldontls "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

type TLSOptions struct {
	TLS  bool
	Cert *seldontls.CertificateStore
}

func CreateUpstreamDataplaneServerTLSOptions() (TLSOptions, error) {
	var err error
	tlsOptions := TLSOptions{}
	protocol := seldontls.GetSecurityProtocolFromEnv(seldontls.EnvSecurityPrefixEnvoy)
	if protocol == seldontls.SecurityProtocolSSL {
		tlsOptions.TLS = true

		// Servers to receive requests from Envoy
		tlsOptions.Cert, err = seldontls.NewCertificateStore(seldontls.Prefix(seldontls.EnvSecurityPrefixEnvoyUpstreamServer),
			seldontls.ValidationPrefix(seldontls.EnvSecurityPrefixEnvoyUpstreamClient))
		if err != nil {
			return tlsOptions, err
		}
	}
	return tlsOptions, nil
}

func CreateTLSClientOptions() (*TLSOptions, error) {
	protocol := seldontls.GetSecurityProtocolFromEnv(seldontls.EnvSecurityPrefixEnvoy)
	if protocol == seldontls.SecurityProtocolSSL {
		mTLS, err := GetBoolEnvar(seldontls.EnvSecurityDownstreamClientMTLS, false)
		if err != nil {
			return nil, err
		}
		certStore, err := seldontls.NewCertificateStore(seldontls.Prefix(seldontls.EnvSecurityPrefixEnvoyDownstreamClient),
			seldontls.ValidationPrefix(seldontls.EnvSecurityPrefixEnvoyDownstreamServer),
			seldontls.ValidationOnly(!mTLS))
		if err != nil {
			return nil, err
		}
		return &TLSOptions{
			TLS:  true,
			Cert: certStore,
		}, nil
	}
	return &TLSOptions{}, nil
}
