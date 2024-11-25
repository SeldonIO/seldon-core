/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package resources

import (
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"
)

const (
	xdsClusterName = "xds_cluster"
)

func MakeSecretResource(name string, validationName string, certStore tls.CertificateStoreHandler) []*tlsv3.Secret {
	var secrets []*tlsv3.Secret

	if certStore.GetValidationCertificate() != nil {
		secrets = append(secrets, &tlsv3.Secret{
			Name: validationName,
			Type: &tlsv3.Secret_ValidationContext{
				ValidationContext: &tlsv3.CertificateValidationContext{
					TrustedCa: &core.DataSource{
						Specifier: &core.DataSource_InlineBytes{InlineBytes: certStore.GetValidationCertificate().CaRaw},
					},
				},
			},
		})
	}

	secrets = append(secrets, &tlsv3.Secret{
		Name: name,
		Type: &tlsv3.Secret_TlsCertificate{
			TlsCertificate: &tlsv3.TlsCertificate{
				CertificateChain: &core.DataSource{
					Specifier: &core.DataSource_InlineBytes{InlineBytes: certStore.GetCertificate().CrtRaw},
				},
				PrivateKey: &core.DataSource{
					Specifier: &core.DataSource_InlineBytes{InlineBytes: certStore.GetCertificate().KeyRaw},
				},
			},
		},
	})
	return secrets
}

var configSource = &core.ConfigSource{
	ResourceApiVersion: core.ApiVersion_V3,
	ConfigSourceSpecifier: &core.ConfigSource_ApiConfigSource{
		ApiConfigSource: &core.ApiConfigSource{
			ApiType:             core.ApiConfigSource_GRPC,
			TransportApiVersion: core.ApiVersion_V3,
			GrpcServices: []*core.GrpcService{
				{
					TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
						EnvoyGrpc: &core.GrpcService_EnvoyGrpc{
							ClusterName: xdsClusterName,
						},
					},
				},
			},
		},
	},
}

func createDownstreamTransportSocketV2(serverSecret *tlsv3.Secret) *core.TransportSocket {
	var ts *core.TransportSocket
	if serverSecret != nil {
		tlsCtx := tlsv3.DownstreamTlsContext{
			CommonTlsContext: &tlsv3.CommonTlsContext{
				TlsCertificateSdsSecretConfigs: []*tlsv3.SdsSecretConfig{
					{
						Name:      serverSecret.Name,
						SdsConfig: configSource,
					},
				},
			},
		}
		if serverSecret.GetTlsCertificate() == nil {
			tlsCtx.CommonTlsContext.ValidationContextType = &tlsv3.CommonTlsContext_ValidationContextSdsSecretConfig{
				ValidationContextSdsSecretConfig: &tlsv3.SdsSecretConfig{
					Name:      serverSecret.Name,
					SdsConfig: configSource,
				},
			}
		}

		tlsCtxPb, err := anypb.New(&tlsCtx)
		if err != nil {
			panic(err)
		}

		ts = &core.TransportSocket{
			Name: "envoy.transport_sockets.tls",
			ConfigType: &core.TransportSocket_TypedConfig{
				TypedConfig: tlsCtxPb,
			},
		}
	}
	return ts
}

func createDownstreamTransportSocket(serverSecret *Secret) *core.TransportSocket {
	var ts *core.TransportSocket
	if serverSecret != nil {
		tlsCtx := tlsv3.DownstreamTlsContext{
			CommonTlsContext: &tlsv3.CommonTlsContext{
				TlsCertificateSdsSecretConfigs: []*tlsv3.SdsSecretConfig{
					{
						Name:      serverSecret.Name,
						SdsConfig: configSource,
					},
				},
			},
		}
		if serverSecret.Certificate.GetValidationCertificate() != nil {
			tlsCtx.CommonTlsContext.ValidationContextType = &tlsv3.CommonTlsContext_ValidationContextSdsSecretConfig{
				ValidationContextSdsSecretConfig: &tlsv3.SdsSecretConfig{
					Name:      serverSecret.ValidationSecretName,
					SdsConfig: configSource,
				},
			}
		}

		tlsCtxPb, err := anypb.New(&tlsCtx)
		if err != nil {
			panic(err)
		}

		ts = &core.TransportSocket{
			Name: "envoy.transport_sockets.tls",
			ConfigType: &core.TransportSocket_TypedConfig{
				TypedConfig: tlsCtxPb,
			},
		}
	}
	return ts
}

func createUpstreamTransportSocketV2(clientSecret *tlsv3.Secret) *core.TransportSocket {
	var ts *core.TransportSocket
	if clientSecret != nil {
		tlsCtx := tlsv3.UpstreamTlsContext{
			CommonTlsContext: &tlsv3.CommonTlsContext{
				TlsCertificateSdsSecretConfigs: []*tlsv3.SdsSecretConfig{
					{
						Name:      clientSecret.Name,
						SdsConfig: configSource,
					},
				},
			},
		}
		if clientSecret.GetTlsCertificate() != nil {
			tlsCtx.CommonTlsContext.ValidationContextType = &tlsv3.CommonTlsContext_ValidationContextSdsSecretConfig{
				ValidationContextSdsSecretConfig: &tlsv3.SdsSecretConfig{
					Name:      clientSecret.Name,
					SdsConfig: configSource,
				},
			}
		}

		tlsCtxPb, err := anypb.New(&tlsCtx)
		if err != nil {
			panic(err)
		}

		ts = &core.TransportSocket{
			Name: "envoy.transport_sockets.tls",
			ConfigType: &core.TransportSocket_TypedConfig{
				TypedConfig: tlsCtxPb,
			},
		}
	}
	return ts
}

func createUpstreamTransportSocket(clientSecret *Secret) *core.TransportSocket {
	var ts *core.TransportSocket
	if clientSecret != nil {
		tlsCtx := tlsv3.UpstreamTlsContext{
			CommonTlsContext: &tlsv3.CommonTlsContext{
				TlsCertificateSdsSecretConfigs: []*tlsv3.SdsSecretConfig{
					{
						Name:      clientSecret.Name,
						SdsConfig: configSource,
					},
				},
			},
		}
		if clientSecret.Certificate.GetValidationCertificate() != nil {
			tlsCtx.CommonTlsContext.ValidationContextType = &tlsv3.CommonTlsContext_ValidationContextSdsSecretConfig{
				ValidationContextSdsSecretConfig: &tlsv3.SdsSecretConfig{
					Name:      clientSecret.ValidationSecretName,
					SdsConfig: configSource,
				},
			}
		}

		tlsCtxPb, err := anypb.New(&tlsCtx)
		if err != nil {
			panic(err)
		}

		ts = &core.TransportSocket{
			Name: "envoy.transport_sockets.tls",
			ConfigType: &core.TransportSocket_TypedConfig{
				TypedConfig: tlsCtxPb,
			},
		}
	}
	return ts
}
