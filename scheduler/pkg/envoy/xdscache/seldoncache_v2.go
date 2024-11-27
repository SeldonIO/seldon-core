/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package xdscache

import (
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/sirupsen/logrus"

	seldontls "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/resources"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

type ClusterRoutes map[string]bool

type SeldonXDSCacheV2 struct {
	Listeners              []types.Resource
	Routes                 map[string]*routev3.Route
	Mirrors                map[string]*routev3.Route
	Clusters               map[string]*clusterv3.Cluster
	Secrets                map[string]types.Resource
	PipelineGatewayDetails *PipelineGatewayDetails
	routeConfigs           []*routev3.RouteConfiguration
	clusterRoutes          map[string]ClusterRoutes
	logger                 logrus.FieldLogger
	TLSActive              bool
}

var _ SeldonXDSCache = (*SeldonXDSCacheV2)(nil)

func NewSeldonXDSCacheV2(logger logrus.FieldLogger, pipelineGatewayDetails *PipelineGatewayDetails) *SeldonXDSCacheV2 {
	logger = logger.WithField("source", "XDSCache")
	cache := &SeldonXDSCacheV2{
		Clusters:               make(map[string]*clusterv3.Cluster),
		Routes:                 make(map[string]*routev3.Route),
		Mirrors:                make(map[string]*routev3.Route),
		Secrets:                make(map[string]types.Resource),
		PipelineGatewayDetails: pipelineGatewayDetails,
		clusterRoutes:          make(map[string]ClusterRoutes),
		logger:                 logger,
	}

	err := cache.SetupTLS()
	if err != nil {
		logger.Error(err)
	}
	cache.SetupListeners()
	cache.SetupClusters()
	cache.SetupRoutes()

	return cache
}

func (xds *SeldonXDSCacheV2) ClusterContents() []types.Resource {
	clusters := make([]types.Resource, len(xds.Clusters))
	for _, cluster := range xds.Clusters {
		clusters = append(clusters, cluster)
	}
	return clusters
}

func (xds *SeldonXDSCacheV2) ListenerContents() []types.Resource {
	return xds.Listeners
}

func (xds *SeldonXDSCacheV2) RouteContents() []types.Resource {
	routes := make([]*routev3.Route, len(xds.Routes))
	for _, route := range xds.Routes {
		routes = append(routes, route)
	}

	mirrors := make([]*routev3.Route, len(xds.Mirrors))
	for _, mirror := range xds.Mirrors {
		mirrors = append(mirrors, mirror)
	}

	xds.routeConfigs[0].VirtualHosts[0].Routes = routes
	xds.routeConfigs[1].VirtualHosts[0].Routes = mirrors

	return []types.Resource{xds.routeConfigs[0], xds.routeConfigs[1]}
}

func (xds *SeldonXDSCacheV2) SecretContents() []types.Resource {
	secrets := make([]types.Resource, len(xds.Secrets))
	for _, secret := range xds.Secrets {
		secrets = append(secrets, secret)
	}
	return secrets
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
	xds.clusterRoutes[resources.PipelineGatewayHttpClusterName] = map[string]bool{resources.PipelineGatewayHttpClusterName: true}

	xds.logger.Infof("Add grpc pipeline cluster %s host:%s port:%d", resources.PipelineGatewayGrpcClusterName, xds.PipelineGatewayDetails.Host, xds.PipelineGatewayDetails.GrpcPort)
	xds.Clusters[resources.PipelineGatewayGrpcClusterName] = resources.MakeClusterV2(resources.PipelineGatewayGrpcClusterName, []resources.Endpoint{
		{
			UpstreamHost: xds.PipelineGatewayDetails.Host,
			UpstreamPort: uint32(xds.PipelineGatewayDetails.GrpcPort),
		},
	}, true, clientSecret)
	xds.clusterRoutes[resources.PipelineGatewayGrpcClusterName] = map[string]bool{resources.PipelineGatewayGrpcClusterName: true}

	// Add Mirror clusters
	xds.logger.Infof("Add http mirror cluster %s host:%s port:%d", resources.MirrorHttpClusterName, mirrorListenerAddress, mirrorListenerPort)
	xds.Clusters[resources.MirrorHttpClusterName] = resources.MakeClusterV2(resources.MirrorHttpClusterName, []resources.Endpoint{
		{
			UpstreamHost: mirrorListenerAddress,
			UpstreamPort: mirrorListenerPort,
		},
	}, false, nil)
	xds.clusterRoutes[resources.MirrorHttpClusterName] = map[string]bool{resources.MirrorHttpClusterName: true}

	xds.logger.Infof("Add grpc mirror cluster %s host:%s port:%d", resources.MirrorGrpcClusterName, mirrorListenerAddress, mirrorListenerPort)
	xds.Clusters[resources.MirrorGrpcClusterName] = resources.MakeClusterV2(resources.MirrorGrpcClusterName, []resources.Endpoint{
		{
			UpstreamHost: mirrorListenerAddress,
			UpstreamPort: mirrorListenerPort,
		},
	}, true, nil)
	xds.clusterRoutes[resources.MirrorGrpcClusterName] = map[string]bool{resources.MirrorGrpcClusterName: true}
}

func (xds *SeldonXDSCacheV2) SetupListeners() {
	var serverSecret *tlsv3.Secret
	if xds.TLSActive {
		if secret, ok := xds.Secrets[EnvoyDownstreamServerCertName]; ok {
			serverSecret = secret.(*tlsv3.Secret)
		}
	}
	listeners := make([]types.Resource, 2)
	listeners[0] = resources.MakeHTTPListenerV2(defaultListenerName, defaultListenerAddress, defaultListenerPort, resources.DefaultRouteConfigurationName, serverSecret)
	listeners[1] = resources.MakeHTTPListenerV2(mirrorListenerName, mirrorListenerAddress, mirrorListenerPort, resources.MirrorRouteConfigurationName, serverSecret)
	xds.Listeners = listeners
}

func (xds *SeldonXDSCacheV2) SetupRoutes() {
	xds.routeConfigs = []*routev3.RouteConfiguration{
		{
			Name: resources.DefaultRouteConfigurationName,
			VirtualHosts: []*routev3.VirtualHost{{
				Name:    "seldon_service",
				Domains: []string{"*"},
			}},
		},
		{
			Name: resources.MirrorRouteConfigurationName,
			VirtualHosts: []*routev3.VirtualHost{{
				Name:    "seldon_mirror",
				Domains: []string{"*"},
			}},
		},
	}
}

func (xds *SeldonXDSCacheV2) AddSecret(name string, validationSecretName string, certificate *seldontls.CertificateStore) {
	secrets := resources.MakeSecretResource(name, validationSecretName, certificate)
	for _, secret := range secrets {
		xds.Secrets[secret.Name] = secret
	}
}

func (xds *SeldonXDSCacheV2) AddPipelineRoute(routeName string, pipelineName string, trafficWeight uint32, isMirror bool) {
	if !isMirror {
		// REST
		routeNameHttp := resources.GetRouteName(routeName, true, false, isMirror)
		pipelineRouteHttp, ok := xds.Routes[routeNameHttp]

		if !ok {
			pipelineRouteHttp := resources.MakePipelineRouteV2(routeNameHttp, pipelineName, trafficWeight, isMirror, false)
			xds.Routes[routeNameHttp] = pipelineRouteHttp
		} else {
			pipelineRouteHttp = resources.AddWeightedClusterToPipeline(pipelineRouteHttp, pipelineName, resources.PipelineGatewayHttpClusterName, trafficWeight)
			xds.Routes[routeNameHttp] = pipelineRouteHttp
		}
		// gRPC
		routeNameGrpc := resources.GetRouteName(routeName, true, true, isMirror)
		pipelineRouteGrpc, ok := xds.Routes[routeNameGrpc]

		if !ok {
			pipelineRouteGrpc := resources.MakePipelineRouteV2(routeNameGrpc, pipelineName, trafficWeight, isMirror, true)
			xds.Routes[routeNameGrpc] = pipelineRouteGrpc

		} else {
			pipelineRouteGrpc = resources.AddWeightedClusterToPipeline(pipelineRouteGrpc, pipelineName, resources.PipelineGatewayGrpcClusterName, trafficWeight)
			xds.Routes[routeNameGrpc] = pipelineRouteGrpc
		}

	} else {
		// REST
		routeNameHttp := resources.GetRouteName(routeName, true, false, isMirror)
		pipelineRouteHttp, ok := xds.Mirrors[routeNameHttp]

		if !ok {
			pipelineRouteHttp := resources.MakePipelineRouteV2(routeNameHttp, pipelineName, trafficWeight, isMirror, false)
			xds.Routes[routeNameHttp] = pipelineRouteHttp
		} else {
			routeAction := pipelineRouteHttp.GetRoute()
			routeAction.RequestMirrorPolicies = resources.CreateMirrorRouteAction(resources.PipelineGatewayHttpClusterName, trafficWeight)
			xds.Routes[routeNameHttp] = pipelineRouteHttp
		}
		// gRPC
		routeNameGrpc := resources.GetRouteName(routeName, true, true, isMirror)
		pipelineRouteGrpc, ok := xds.Mirrors[routeNameGrpc]

		if !ok {
			pipelineRouteGrpc := resources.MakePipelineRouteV2(routeNameGrpc, pipelineName, trafficWeight, isMirror, true)
			xds.Routes[routeNameGrpc] = pipelineRouteGrpc

		} else {
			routeAction := pipelineRouteGrpc.GetRoute()
			routeAction.RequestMirrorPolicies = resources.CreateMirrorRouteAction(resources.PipelineGatewayGrpcClusterName, trafficWeight)
			xds.Routes[routeNameGrpc] = pipelineRouteGrpc
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
	clusterTraffic := resources.TrafficSplits{
		ModelName:     modelName,
		ModelVersion:  modelVersion,
		TrafficWeight: trafficPercent,
		HttpCluster:   httpClusterName,
		GrpcCluster:   grpcClusterName,
	}

	httpRouteName := resources.GetRouteName(routeName, false, false, isMirror)

	httpRoute, ok := xds.Routes[httpRouteName]
	if !ok {
		httpRoute = resources.MakeModelRouteV2(routeName, []resources.TrafficSplits{clusterTraffic}, isMirror, false, logPayloads)
	} else {
		httpRouteAction := httpRoute.GetRoute()
		weightedClusters := httpRouteAction.ClusterSpecifier.(*routev3.RouteAction_WeightedClusters).WeightedClusters.GetClusters()
		weightedClusters = append(weightedClusters, resources.CreateModelWeightedCluster(httpClusterName, clusterTraffic))
		httpRouteAction.ClusterSpecifier.(*routev3.RouteAction_WeightedClusters).WeightedClusters.Clusters = weightedClusters
	}

	httpClusterRoutes := xds.clusterRoutes[httpClusterName]
	_, ok = httpClusterRoutes[httpRouteName]
	if !ok {
		httpClusterRoutes[httpRouteName] = true
		xds.clusterRoutes[grpcClusterName] = httpClusterRoutes
	}

	grpcRouteName := resources.GetRouteName(routeName, false, true, isMirror)

	grpcRoute, ok := xds.Routes[grpcRouteName]
	if !ok {
		grpcRoute = resources.MakeModelRouteV2(routeName, []resources.TrafficSplits{clusterTraffic}, isMirror, true, logPayloads)
	} else {
		grpcRouteAction := grpcRoute.GetRoute()
		weightedClusters := grpcRouteAction.ClusterSpecifier.(*routev3.RouteAction_WeightedClusters).WeightedClusters.GetClusters()
		weightedClusters = append(weightedClusters, resources.CreateModelWeightedCluster(grpcClusterName, clusterTraffic))
		grpcRouteAction.ClusterSpecifier.(*routev3.RouteAction_WeightedClusters).WeightedClusters.Clusters = weightedClusters
	}

	grpcClusterRoutes := xds.clusterRoutes[grpcClusterName]
	_, ok = grpcClusterRoutes[grpcRouteName]
	if !ok {
		grpcClusterRoutes[grpcRouteName] = true
		xds.clusterRoutes[grpcClusterName] = grpcClusterRoutes
	}

	if isMirror {
		httpRouteAction := httpRoute.GetRoute()
		httpRouteAction.RequestMirrorPolicies = resources.CreateMirrorRouteAction(resources.MirrorHttpClusterName, trafficPercent)

		grpcRouteAction := grpcRoute.GetRoute()
		grpcRouteAction.RequestMirrorPolicies = resources.CreateMirrorRouteAction(resources.MirrorGrpcClusterName, trafficPercent)
	}

	xds.Routes[httpRouteName] = httpRoute
	xds.Routes[grpcRouteName] = grpcRoute
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
		xds.clusterRoutes[name] = make(ClusterRoutes)
	}

	xds.Clusters[name] = cluster
}

func (xds *SeldonXDSCacheV2) RemoveRoute(routeName string) error {
	logger := xds.logger.WithField("func", "RemoveRoute")
	logger.Infof("Remove routes for model %s", routeName)

	httpRouteName := resources.GetRouteName(routeName, false, false, false)

	httpRoute, ok := xds.Routes[httpRouteName]
	if !ok {
		logger.Warnf("No route named %s found", httpRouteName)
	} else {
		xds.updateClusterRoutes(httpRouteName, httpRoute)
		delete(xds.Routes, httpRouteName)
	}

	grpcRouteName := resources.GetRouteName(routeName, false, true, false)

	grpcRoute, ok := xds.Routes[grpcRouteName]
	if !ok {
		logger.Warnf("No route named %s found", grpcRoute)
	} else {
		xds.updateClusterRoutes(grpcRouteName, grpcRoute)
		delete(xds.Routes, grpcRouteName)
	}

	httpMirrorName := resources.GetRouteName(routeName, false, false, true)

	_, ok = xds.Mirrors[httpMirrorName]
	if !ok {
		logger.Warnf("No mirror named %s found", httpMirrorName)
	} else {
		delete(xds.Mirrors, httpMirrorName)
	}

	grpcMirrorName := resources.GetRouteName(routeName, false, true, true)

	_, ok = xds.Mirrors[grpcMirrorName]
	if !ok {
		logger.Warnf("No mirror named %s found", grpcMirrorName)
	} else {
		delete(xds.Mirrors, grpcMirrorName)
	}
	return nil
}

func (xds *SeldonXDSCacheV2) updateClusterRoutes(routeName string, route *routev3.Route) {
	weightedCluster := route.GetRoute().GetWeightedClusters()
	for _, cluster := range weightedCluster.Clusters {
		clusterRoutes := xds.clusterRoutes[cluster.Name]
		delete(clusterRoutes, routeName)
		if len(clusterRoutes) == 0 {
			delete(xds.clusterRoutes, cluster.Name)
			delete(xds.Clusters, cluster.Name)
		}
	}
}

func (xds *SeldonXDSCacheV2) AddEndpoint(clusterName string, upstreamHost string, upstreamPort uint32, assignments []int, replicas map[int]*store.ServerReplica, index int) {
	// TODO: remove this optimisation, which is here to make it compatible with the current implementation
	if index > 0 {
		return
	}
	logger := xds.logger.WithField("func", "AddEndpoint")

	cluster, ok := xds.Clusters[clusterName]

	if !ok {
		logger.Warnf("No cluster named %s found", clusterName)
	} else {
		cluster.LoadAssignment = resources.MakeEndpointV2(clusterName, assignments, replicas)
	}
}

func (xds *SeldonXDSCacheV2) GetRoute(routeName string) (resources.Route, bool) {
	_, ok := xds.Routes[routeName]

	return resources.Route{RouteName: routeName}, ok
}
