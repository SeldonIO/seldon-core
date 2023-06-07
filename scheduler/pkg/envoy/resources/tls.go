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
