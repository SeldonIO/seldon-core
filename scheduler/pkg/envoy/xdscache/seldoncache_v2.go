/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package xdscache

import (
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/sirupsen/logrus"

	seldontls "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

	clusterEnvoy "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoyRoute "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/resources"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

type ClusterRoutes map[string]bool

var _ SeldonXDSCache = (*SeldonXDSCacheV2)(nil)

type SeldonXDSCacheV2 struct {
	Listeners              map[string]types.Resource
	Routes                 map[string]types.Resource
	Mirrors                map[string]types.Resource
	Clusters               map[string]types.Resource
	Endpoints              map[string][]types.Resource
	Secrets                map[string]types.Resource
	clusterRoutes          map[string]ClusterRoutes
	PipelineGatewayDetails *PipelineGatewayDetails
	logger                 logrus.FieldLogger
	TLSActive              bool
}

func NewSeldonXDSCacheV2(logger logrus.FieldLogger, pipelineGatewayDetails *PipelineGatewayDetails) *SeldonXDSCacheV2 {
	return &SeldonXDSCacheV2{
		Listeners:              make(map[string]types.Resource),
		Clusters:               make(map[string]types.Resource),
		Endpoints:              make(map[string][]types.Resource),
		Routes:                 make(map[string]types.Resource),
		Mirrors:                make(map[string]types.Resource),
		Secrets:                make(map[string]types.Resource),
		clusterRoutes:          make(map[string]ClusterRoutes),
		PipelineGatewayDetails: pipelineGatewayDetails,
		logger:                 logger.WithField("source", "XDSCache"),
	}
}

func (xds *SeldonXDSCacheV2) ClusterContents() []types.Resource {
	contents := make([]types.Resource, 0)
	for _, val := range xds.Clusters {
		contents = append(contents, val)
	}
	return contents
}

func (xds *SeldonXDSCacheV2) RouteContents() []types.Resource {
	contents := make([]types.Resource, 0)
	for _, val := range xds.Routes {
		contents = append(contents, val)
	}
	for _, val := range xds.Mirrors {
		contents = append(contents, val)
	}
	return contents
}

func (xds *SeldonXDSCacheV2) ListenerContents() []types.Resource {
	contents := make([]types.Resource, 0)
	for _, val := range xds.Listeners {
		contents = append(contents, val)
	}
	return contents
}

func (xds *SeldonXDSCacheV2) SecretContents() []types.Resource {
	contents := make([]types.Resource, 0)
	for _, secret := range xds.Secrets {
		contents = append(contents, secret)
	}
	return contents
}

func (xds *SeldonXDSCacheV2) AddSecret(name string, validationSecretName string, certificate *seldontls.CertificateStore) {
	secrets := resources.MakeSecretResource(name, validationSecretName, certificate)
	for _, secret := range secrets {
		xds.Secrets[secret.Name] = secret
	}
}

func (xds *SeldonXDSCacheV2) SetupTLS() error {
	logger := xds.logger.WithField("func", "SetupTLS")
	protocol := seldontls.GetSecurityProtocolFromEnv(seldontls.EnvSecurityPrefixEnvoy)
	if protocol == seldontls.SecurityProtocolSSL {
		xds.TLSActive = true

		// Envoy client to talk to agent or Pipelinegateway
		logger.Info("Upstream TLS active")
		tlsUpstreamClient, err := seldontls.NewCertificateStore(seldontls.Prefix(seldontls.EnvSecurityPrefixEnvoyUpstreamClient),
			seldontls.ValidationPrefix(seldontls.EnvSecurityPrefixEnvoyUpstreamServer))
		if err != nil {
			return err
		}
		xds.AddSecret(EnvoyUpstreamClientCertName, EnvoyUpstreamServerCertName, tlsUpstreamClient)

		// Envoy listener - external calls to Seldon
		logger.Info("Downstream TLS active")
		tlsDownstreamServer, err := seldontls.NewCertificateStore(seldontls.Prefix(seldontls.EnvSecurityPrefixEnvoyDownstreamServer),
			seldontls.ValidationPrefix(seldontls.EnvSecurityPrefixEnvoyDownstreamClient))
		if err != nil {
			return err
		}
		xds.AddSecret(EnvoyDownstreamServerCertName, EnvoyDownstreamClientCertName, tlsDownstreamServer)
	}
	return nil
}

func (xds *SeldonXDSCacheV2) SetupClusters() {
	var clientSecret *tlsv3.Secret
	if xds.TLSActive {
		if secret, ok := xds.Secrets[EnvoyUpstreamClientCertName]; ok {
			clientSecret = secret.(*tlsv3.Secret)
		}
	}
	// Add pipeline gateway clusters
	xds.logger.Infof("Add http pipeline cluster %s host:%s port:%d", resources.PipelineGatewayHttpClusterName, xds.PipelineGatewayDetails.Host, xds.PipelineGatewayDetails.HttpPort)
	xds.Clusters[resources.PipelineGatewayHttpClusterName] = resources.MakeClusterV2(resources.PipelineGatewayHttpClusterName, []resources.Endpoint{
		{
			UpstreamHost: xds.PipelineGatewayDetails.Host,
			UpstreamPort: uint32(xds.PipelineGatewayDetails.HttpPort),
		},
	}, false, clientSecret)

	xds.logger.Infof("Add grpc pipeline cluster %s host:%s port:%d", resources.PipelineGatewayGrpcClusterName, xds.PipelineGatewayDetails.Host, xds.PipelineGatewayDetails.GrpcPort)
	xds.Clusters[resources.PipelineGatewayHttpClusterName] = resources.MakeClusterV2(resources.PipelineGatewayGrpcClusterName, []resources.Endpoint{
		{
			UpstreamHost: xds.PipelineGatewayDetails.Host,
			UpstreamPort: uint32(xds.PipelineGatewayDetails.GrpcPort),
		},
	}, true, clientSecret)

	// Add Mirror clusters
	xds.logger.Infof("Add http mirror cluster %s host:%s port:%d", resources.MirrorHttpClusterName, mirrorListenerAddress, mirrorListenerPort)
	xds.Clusters[resources.PipelineGatewayHttpClusterName] = resources.MakeClusterV2(resources.MirrorHttpClusterName, []resources.Endpoint{
		{
			UpstreamHost: mirrorListenerAddress,
			UpstreamPort: mirrorListenerPort,
		},
	}, false, nil)
	xds.logger.Infof("Add grpc mirror cluster %s host:%s port:%d", resources.MirrorGrpcClusterName, mirrorListenerAddress, mirrorListenerPort)
	xds.Clusters[resources.PipelineGatewayHttpClusterName] = resources.MakeClusterV2(resources.MirrorGrpcClusterName, []resources.Endpoint{
		{
			UpstreamHost: mirrorListenerAddress,
			UpstreamPort: mirrorListenerPort,
		},
	}, true, nil)
}

func (xds *SeldonXDSCacheV2) SetupListeners() {
	var serverSecret *tlsv3.Secret
	if xds.TLSActive {
		if secret, ok := xds.Secrets[EnvoyDownstreamServerCertName]; ok {
			serverSecret = secret.(*tlsv3.Secret)
		}
	}
	xds.Listeners[defaultListenerName] = resources.MakeHTTPListenerV2(defaultListenerName, defaultListenerAddress, defaultListenerPort, resources.DefaultRouteConfigurationName, serverSecret)
	xds.Listeners[mirrorListenerName] = resources.MakeHTTPListenerV2(mirrorListenerName, mirrorListenerAddress, mirrorListenerPort, resources.MirrorRouteConfigurationName, serverSecret)
}

func (xds *SeldonXDSCacheV2) AddPipelineRoute(routeName string, pipelineName string, trafficWeight uint32, isMirror bool) {
	if !isMirror {
		routeNameHttp := resources.GetRouteName(routeName, true, false, isMirror)
		pipelineRouteHttp, ok := xds.Routes[routeNameHttp]

		if !ok {
			pipelineRouteHttp := resources.MakePipelineRouteV2(routeNameHttp, pipelineName, trafficWeight, isMirror, false)
			route := pipelineRouteHttp.GetRoute()
			var splits []*envoyRoute.WeightedCluster_ClusterWeight
			splits = append(splits, resources.CreateWeightedCluster(resources.PipelineGatewayHttpClusterName, pipelineName, trafficWeight))
			weightedClusters := &envoyRoute.RouteAction_WeightedClusters{
				WeightedClusters: &envoyRoute.WeightedCluster{
					Clusters: splits,
				},
			}

			route.ClusterSpecifier = weightedClusters
			xds.Routes[routeNameHttp] = pipelineRouteHttp
		} else {
			route := pipelineRouteHttp.(*envoyRoute.Route).GetRoute()
			weightedClusters := route.ClusterSpecifier.(*envoyRoute.RouteAction_WeightedClusters).WeightedClusters.GetClusters()
			weightedClusters = append(weightedClusters, resources.CreateWeightedCluster(resources.PipelineGatewayHttpClusterName, pipelineName, trafficWeight))
			route.ClusterSpecifier.(*envoyRoute.RouteAction_WeightedClusters).WeightedClusters.Clusters = weightedClusters
		}

		routeNameGrpc := resources.GetRouteName(routeName, true, true, isMirror)
		pipelineRouteGrpc, ok := xds.Routes[routeNameGrpc]

		if !ok {
			pipelineRouteGrpc := resources.MakePipelineRouteV2(routeNameGrpc, pipelineName, trafficWeight, isMirror, true)
			route := pipelineRouteGrpc.GetRoute()
			var splits []*envoyRoute.WeightedCluster_ClusterWeight
			splits = append(splits, resources.CreateWeightedCluster(resources.PipelineGatewayGrpcClusterName, pipelineName, trafficWeight))
			weightedClusters := &envoyRoute.RouteAction_WeightedClusters{
				WeightedClusters: &envoyRoute.WeightedCluster{
					Clusters: splits,
				},
			}
			route.ClusterSpecifier = weightedClusters
			xds.Routes[routeNameGrpc] = pipelineRouteGrpc

		} else {
			route := pipelineRouteGrpc.(*envoyRoute.Route).GetRoute()
			weightedClusters := route.ClusterSpecifier.(*envoyRoute.RouteAction_WeightedClusters).WeightedClusters.GetClusters()
			weightedClusters = append(weightedClusters, resources.CreateWeightedCluster(resources.PipelineGatewayGrpcClusterName, pipelineName, trafficWeight))
			route.ClusterSpecifier.(*envoyRoute.RouteAction_WeightedClusters).WeightedClusters.Clusters = weightedClusters
		}

	} else {
		// convert route name (http)
		routeNameHttp := ""
		pipelineRouteHttp, ok := xds.Mirrors[routeNameHttp]

		if !ok {
			pipelineRouteHttp := resources.MakePipelineRouteV2(routeNameHttp, pipelineName, trafficWeight, isMirror, false)
			route := pipelineRouteHttp.GetRoute()
			route.RequestMirrorPolicies = resources.CreateMirrorRouteAction(resources.PipelineGatewayHttpClusterName, trafficWeight)
			xds.Routes[routeNameHttp] = pipelineRouteHttp
		} else {
			route := pipelineRouteHttp.(*envoyRoute.Route).GetRoute()
			route.RequestMirrorPolicies = resources.CreateMirrorRouteAction(resources.PipelineGatewayHttpClusterName, trafficWeight)
			pipelineRouteHttp.ProtoReflect()
		}

		// convert route name (grpc)
		routeNameGrpc := ""
		pipelineRouteGrpc, ok := xds.Mirrors[routeNameGrpc]

		if !ok {
			pipelineRouteGrpc := resources.MakePipelineRouteV2(routeNameGrpc, pipelineName, trafficWeight, isMirror, true)
			route := pipelineRouteGrpc.GetRoute()
			route.RequestMirrorPolicies = resources.CreateMirrorRouteAction(resources.PipelineGatewayGrpcClusterName, trafficWeight)
			xds.Routes[routeNameGrpc] = pipelineRouteGrpc

		} else {
			route := pipelineRouteGrpc.(*envoyRoute.Route).GetRoute()
			route.RequestMirrorPolicies = resources.CreateMirrorRouteAction(resources.PipelineGatewayGrpcClusterName, trafficWeight)
		}

	}
}

func (xds *SeldonXDSCacheV2) RemovePipelineRoute(pipelineName string) {
	// convert to grpc/http/mirror
	delete(xds.Routes, pipelineName)
}

func (xds *SeldonXDSCacheV2) AddRouteClusterTraffic(
	routeName string,
	modelName string,
	modelVersion uint32,
	trafficPercent uint32,
	httpClusterName string,
	grpcClusterName string,
	logPayloads bool,
	isMirror bool,
) {
}

func (xds *SeldonXDSCacheV2) AddCluster(
	name string,
	routeName string,
	modelName string,
	modelVersion uint32,
	isGrpc bool,
) {
	var clientSecret *tlsv3.Secret
	if xds.TLSActive {
		if secret, ok := xds.Secrets[EnvoyUpstreamClientCertName]; ok {
			clientSecret = secret.(*tlsv3.Secret)
		}
	}

	cluster, ok := xds.Clusters[name]

	if !ok {
		cluster = resources.MakeClusterV2(name, make([]resources.Endpoint, 0), isGrpc, clientSecret)
	}

	xds.Clusters[name] = cluster
}

func (xds *SeldonXDSCacheV2) RemoveRoute(routeName string) error {
	logger := xds.logger.WithField("func", "RemoveRoute")
	logger.Infof("Remove routes for model %s", routeName)
	route, ok := xds.Routes[routeName]
	if !ok {
		logger.Warnf("No route named %s found", routeName)
	} else {
		routePtr := route.(*envoyRoute.Route)
		xds.removeRoute(routeName, routePtr)
	}

	_, ok = xds.Mirrors[routeName]
	if !ok {
		logger.Warnf("No mirror named %s found", routeName)
	} else {
		delete(xds.Mirrors, routeName)
	}
	return nil
}

func (xds *SeldonXDSCacheV2) removeRoute(routeName string, route *envoyRoute.Route) {
	weightedCluster := route.GetRoute().GetWeightedClusters()
	for _, cluster := range weightedCluster.Clusters {
		// get the cluster count and decrement it
		clusterRoutes := xds.clusterRoutes[cluster.Name]
		delete(clusterRoutes, routeName)
		if len(clusterRoutes) == 0 {
			delete(xds.clusterRoutes, cluster.Name)
			delete(xds.Clusters, cluster.Name)
		}

	}
	delete(xds.Mirrors, routeName)
}

func (xds *SeldonXDSCacheV2) AddEndpoint(clusterName string, upstreamHost string, upstreamPort uint32, assignments []int, replicas map[int]*store.ServerReplica) {
	logger := xds.logger.WithField("func", "AddEndpoint")

	cluster, ok := (xds.Clusters[clusterName])

	if !ok {
		logger.Warnf("No cluster named %s found", clusterName)
	} else {
		clusterV2 := cluster.(*clusterEnvoy.Cluster)
		clusterV2.LoadAssignment = resources.MakeEndpointV2(clusterName, assignments, replicas)
	}
}
