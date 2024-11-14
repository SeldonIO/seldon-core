/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package processor

import (
	"fmt"
	"net"

	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	seldontls "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
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
	// endpointservice.RegisterEndpointDiscoveryServiceServer(grpcServer, server)
	// clusterservice.RegisterClusterDiscoveryServiceServer(grpcServer, server)
	// routeservice.RegisterRouteDiscoveryServiceServer(grpcServer, server)
	// listenerservice.RegisterListenerDiscoveryServiceServer(grpcServer, server)
	// secretservice.RegisterSecretDiscoveryServiceServer(grpcServer, server)
	// runtimeservice.RegisterRuntimeDiscoveryServiceServer(grpcServer, server)
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
	kaep := util.GetServerKeepAliveEnforcementPolicy()
	secure := x.certificateStore != nil
	var grpcOptions []grpc.ServerOption
	if secure {
		grpcOptions = append(grpcOptions, grpc.Creds(x.certificateStore.CreateServerTransportCredentials()))
	}
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	grpcOptions = append(grpcOptions, grpc.KeepaliveEnforcementPolicy(kaep))
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
