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
