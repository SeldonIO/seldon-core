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

package server

import (
	"fmt"
	"net"

	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	runtimeservice "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	secretservice "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	seldontls "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"
)

const (
	grpcMaxConcurrentStreams = 1000000
)

type XDSServer struct {
	srv3             serverv3.Server
	certificateStore *seldontls.CertificateStore
	logger           log.FieldLogger
}

func NewXDSServer(server serverv3.Server, logger log.FieldLogger) *XDSServer {
	return &XDSServer{
		srv3:   server,
		logger: logger,
	}
}

func registerServer(grpcServer *grpc.Server, server serverv3.Server) {
	// register services
	discoverygrpc.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)
	//endpointservice.RegisterEndpointDiscoveryServiceServer(grpcServer, server)
	clusterservice.RegisterClusterDiscoveryServiceServer(grpcServer, server)
	routeservice.RegisterRouteDiscoveryServiceServer(grpcServer, server)
	listenerservice.RegisterListenerDiscoveryServiceServer(grpcServer, server)
	secretservice.RegisterSecretDiscoveryServiceServer(grpcServer, server)
	runtimeservice.RegisterRuntimeDiscoveryServiceServer(grpcServer, server)
}

// StartXDSServer starts an xDS server at the given port.
func (x *XDSServer) StartXDSServer(port uint) error {
	logger := x.logger.WithField("func", "StartXDSServer")
	var err error
	protocol := seldontls.GetSecurityProtocolFromEnv(seldontls.EnvSecurityPrefixEnvoy)
	if protocol == seldontls.SecurityProtocolSSL {
		x.certificateStore, err = seldontls.NewCertificateStore(seldontls.Prefix(seldontls.EnvSecurityPrefixControlPlaneServer),
			seldontls.ValidationPrefix(seldontls.EnvSecurityPrefixControlPlaneClient))
		if err != nil {
			return err
		}
	}
	secure := x.certificateStore != nil
	var grpcOptions []grpc.ServerOption
	if secure {
		grpcOptions = append(grpcOptions, grpc.Creds(x.certificateStore.CreateServerTransportCredentials()))
	}
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	grpcServer := grpc.NewServer(grpcOptions...)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	registerServer(grpcServer, x.srv3)
	logger.Infof("Starting xDS envoy server on port %d with secure: %v", port, secure)
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			logger.WithError(err).Fatalf("Envoy xDS server failed on port %d mtls:%v", port, secure)
		}
	}()
	return nil
}
