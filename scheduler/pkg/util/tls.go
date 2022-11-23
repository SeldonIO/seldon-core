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

package util

import seldontls "github.com/seldonio/seldon-core-v2/components/tls/pkg/tls"

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
