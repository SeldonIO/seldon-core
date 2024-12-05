/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package xdscache

import (
	"fmt"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/sirupsen/logrus"

	seldontls "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/resources"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

const (
	defaultListenerName           = "seldon_http"
	defaultListenerAddress        = "0.0.0.0"
	defaultListenerPort    uint32 = 9000

	mirrorListenerName           = "seldon_mirrors"
	mirrorListenerAddress        = "0.0.0.0"
	mirrorListenerPort    uint32 = 9001

	permanentListenerCount int = 2 // seldon_service and seldon_mirrors
	permanentClusterCount  int = 4 // pipeline gateway * 2 + model gateway * 2

	EnvoyDownstreamServerCertName = "downstream_server"
	EnvoyDownstreamClientCertName = "downstream_client"
	EnvoyUpstreamServerCertName   = "upstream_server"
	EnvoyUpstreamClientCertName   = "upstream_client"
)

type SeldonXDSCache struct {
	permanentListeners     []types.Resource
	permanentClusters      []types.Resource
	Routes                 map[string]resources.Route
	Clusters               map[string]resources.Cluster
	Pipelines              map[string]resources.PipelineRoute
	Secrets                map[string]resources.Secret
	PipelineGatewayDetails *PipelineGatewayDetails
	logger                 logrus.FieldLogger
	TLSActive              bool
}

type PipelineGatewayDetails struct {
	Host     string
	HttpPort int
	GrpcPort int
}

func NewSeldonXDSCache(logger logrus.FieldLogger, pipelineGatewayDetails *PipelineGatewayDetails) *SeldonXDSCache {
	return &SeldonXDSCache{
		permanentListeners:     make([]types.Resource, permanentListenerCount),
		permanentClusters:      make([]types.Resource, permanentClusterCount),
		Clusters:               make(map[string]resources.Cluster),
		Routes:                 make(map[string]resources.Route),
		Pipelines:              make(map[string]resources.PipelineRoute),
		Secrets:                make(map[string]resources.Secret),
		PipelineGatewayDetails: pipelineGatewayDetails,
		logger:                 logger.WithField("source", "SeldonXDSCache"),
	}
}

func (xds *SeldonXDSCache) SetupTLS() error {
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

func (xds *SeldonXDSCache) AddPermanentListeners() {
	var serverSecret *resources.Secret
	if xds.TLSActive {
		if secret, ok := xds.Secrets[EnvoyDownstreamServerCertName]; ok {
			serverSecret = &secret
		}
	}
	xds.permanentListeners[0] = resources.MakeHTTPListener(defaultListenerName, defaultListenerAddress, defaultListenerPort, resources.DefaultRouteConfigurationName, serverSecret)
	xds.permanentListeners[1] = resources.MakeHTTPListener(mirrorListenerName, mirrorListenerAddress, mirrorListenerPort, resources.MirrorRouteConfigurationName, serverSecret)
}

func (xds *SeldonXDSCache) AddPermanentClusters() {
	var clientSecret *resources.Secret
	if xds.TLSActive {
		if secret, ok := xds.Secrets[EnvoyUpstreamClientCertName]; ok {
			clientSecret = &secret
		}
	}

	// Add pipeline gateway clusters
	xds.logger.Infof("Add http pipeline cluster %s host:%s port:%d", resources.PipelineGatewayHttpClusterName, xds.PipelineGatewayDetails.Host, xds.PipelineGatewayDetails.HttpPort)
	xds.permanentClusters[0] = resources.MakeCluster(resources.PipelineGatewayHttpClusterName, []resources.Endpoint{
		{
			UpstreamHost: xds.PipelineGatewayDetails.Host,
			UpstreamPort: uint32(xds.PipelineGatewayDetails.HttpPort),
		},
	}, false, clientSecret)

	xds.logger.Infof("Add grpc pipeline cluster %s host:%s port:%d", resources.PipelineGatewayGrpcClusterName, xds.PipelineGatewayDetails.Host, xds.PipelineGatewayDetails.GrpcPort)
	xds.permanentClusters[1] = resources.MakeCluster(resources.PipelineGatewayGrpcClusterName, []resources.Endpoint{
		{
			UpstreamHost: xds.PipelineGatewayDetails.Host,
			UpstreamPort: uint32(xds.PipelineGatewayDetails.GrpcPort),
		},
	}, true, clientSecret)

	// Add Mirror clusters
	xds.logger.Infof("Add http mirror cluster %s host:%s port:%d", resources.MirrorHttpClusterName, mirrorListenerAddress, mirrorListenerPort)
	xds.permanentClusters[2] = resources.MakeCluster(resources.MirrorHttpClusterName, []resources.Endpoint{
		{
			UpstreamHost: mirrorListenerAddress,
			UpstreamPort: mirrorListenerPort,
		},
	}, false, nil)
	xds.logger.Infof("Add grpc mirror cluster %s host:%s port:%d", resources.MirrorGrpcClusterName, mirrorListenerAddress, mirrorListenerPort)
	xds.permanentClusters[3] = resources.MakeCluster(resources.MirrorGrpcClusterName, []resources.Endpoint{
		{
			UpstreamHost: mirrorListenerAddress,
			UpstreamPort: mirrorListenerPort,
		},
	}, true, nil)
}

func (xds *SeldonXDSCache) ClusterContents() []types.Resource {
	var r []types.Resource

	var clientSecret *resources.Secret
	if xds.TLSActive {
		if secret, ok := xds.Secrets[EnvoyUpstreamClientCertName]; ok {
			clientSecret = &secret
		}
	}

	r = append(r, xds.permanentClusters...)

	for _, c := range xds.Clusters {
		endpoints := make([]resources.Endpoint, 0, len(c.Endpoints))
		for _, value := range c.Endpoints { // Likely to be small (<100?) as is number of model replicas
			endpoints = append(endpoints, value)
		}
		r = append(r, resources.MakeCluster(c.Name, endpoints, c.Grpc, clientSecret))
	}

	return r
}

func (xds *SeldonXDSCache) RouteContents() []types.Resource {
	defaultRoutes, mirrorRoutes := resources.MakeRoutes(xds.Routes, xds.Pipelines)
	return []types.Resource{defaultRoutes, mirrorRoutes}
}

func (xds *SeldonXDSCache) ListenerContents() []types.Resource {
	return xds.permanentListeners
}

func (xds *SeldonXDSCache) SecretContents() []types.Resource {
	logger := xds.logger.WithField("func", "SecretContents")
	var r []types.Resource

	for _, s := range xds.Secrets {
		secrets := resources.MakeSecretResource(s.Name, s.ValidationSecretName, s.Certificate)
		logger.Infof("Adding secrets for %s(%s) of length %d", s.Name, s.ValidationSecretName, len(secrets))
		for _, secret := range secrets {
			r = append(r, secret)
		}
	}

	return r
}

func (xds *SeldonXDSCache) AddPipelineRoute(routeName string, trafficSplits []resources.PipelineTrafficSplit, mirror *resources.PipelineTrafficSplit) {
	xds.RemovePipelineRoute(routeName)
	pipelineRoute, ok := xds.Pipelines[routeName]
	if !ok {
		xds.Pipelines[routeName] = resources.PipelineRoute{
			RouteName: routeName,
			Mirror:    mirror,
			Clusters:  trafficSplits,
		}
	} else {
		pipelineRoute.Mirror = mirror
		pipelineRoute.Clusters = trafficSplits
		xds.Pipelines[routeName] = pipelineRoute
	}
}

func (xds *SeldonXDSCache) RemovePipelineRoute(pipelineName string) {
	delete(xds.Pipelines, pipelineName)
}

func (xds *SeldonXDSCache) AddSecret(name string, validationSecretName string, certificate *seldontls.CertificateStore) {
	xds.Secrets[name] = resources.Secret{
		Name:                 name,
		ValidationSecretName: validationSecretName,
		Certificate:          certificate,
	}
}

func (xds *SeldonXDSCache) AddRouteClusterTraffic(
	routeName, modelName, httpClusterName, grpcClusterName string,
	modelVersion uint32,
	trafficPercent uint32,
	logPayloads bool,
	isMirror bool,
) {
	route, ok := xds.Routes[routeName]
	if !ok {
		route = resources.Route{
			RouteName:   routeName,
			LogPayloads: logPayloads,
		}
	}

	// Always log payloads if any version wants it - so during a rolling update if one wants it then it will done
	if logPayloads {
		route.LogPayloads = true
	}

	clusterTraffic := resources.TrafficSplit{
		ModelName:     modelName,
		ModelVersion:  modelVersion,
		TrafficWeight: trafficPercent,
		HttpCluster:   httpClusterName,
		GrpcCluster:   grpcClusterName,
	}

	if isMirror {
		route.Mirror = &clusterTraffic
	} else {
		route.Clusters = append(route.Clusters, clusterTraffic)
	}

	xds.Routes[routeName] = route
}

func (xds *SeldonXDSCache) AddClustersForRoute(
	routeName, modelName, httpClusterName, grpcClusterName string,
	modelVersion uint32,
	assignment []int,
	server *store.ServerSnapshot,
) {
	logger := xds.logger.WithField("func", "AddClustersForRoute")

	routeVersionKey := resources.RouteVersionKey{RouteName: routeName, ModelName: modelName, Version: modelVersion}

	httpCluster, ok := xds.Clusters[httpClusterName]
	if !ok {
		httpCluster = resources.Cluster{
			Name:      httpClusterName,
			Endpoints: make(map[string]resources.Endpoint),
			Routes:    make(map[resources.RouteVersionKey]bool),
			Grpc:      false,
		}
	}
	xds.Clusters[httpClusterName] = httpCluster
	httpCluster.Routes[routeVersionKey] = true

	grpcCluster, ok := xds.Clusters[grpcClusterName]
	if !ok {
		grpcCluster = resources.Cluster{
			Name:      grpcClusterName,
			Endpoints: make(map[string]resources.Endpoint),
			Routes:    make(map[resources.RouteVersionKey]bool),
			Grpc:      true,
		}
	}

	for _, replicaIdx := range assignment {
		replica, ok := server.Replicas[replicaIdx]
		if !ok {
			logger.Warnf("Invalid replica index %d for server %s", replicaIdx, server.Name)
		} else {
			httpEndpointName := fmt.Sprintf("%s:%d", replica.GetInferenceSvc(), replica.GetInferenceHttpPort())
			httpCluster.Endpoints[httpEndpointName] = resources.Endpoint{
				UpstreamHost: replica.GetInferenceSvc(),
				UpstreamPort: uint32(replica.GetInferenceHttpPort()),
			}
			grpcEndpointName := fmt.Sprintf("%s:%d", replica.GetInferenceSvc(), replica.GetInferenceGrpcPort())
			grpcCluster.Endpoints[grpcEndpointName] = resources.Endpoint{
				UpstreamHost: replica.GetInferenceSvc(),
				UpstreamPort: uint32(replica.GetInferenceGrpcPort()),
			}
		}
	}
	xds.Clusters[grpcClusterName] = grpcCluster
	grpcCluster.Routes[routeVersionKey] = true
}

func (xds *SeldonXDSCache) RemoveRoute(routeName string) error {
	logger := xds.logger.WithField("func", "RemoveRoute")
	logger.Infof("Remove routes for model %s", routeName)
	route, ok := xds.Routes[routeName]
	if !ok {
		logger.Warnf("No route found for model %s", routeName)
		return nil
	}
	delete(xds.Routes, routeName)
	for _, cluster := range route.Clusters {
		err := xds.removeRouteFromCluster(route, cluster)
		if err != nil {
			return err
		}
	}
	if route.Mirror != nil {
		err := xds.removeRouteFromCluster(route, *route.Mirror)
		if err != nil {
			return err
		}
	}
	return nil
}

func (xds *SeldonXDSCache) removeRouteFromCluster(route resources.Route, cluster resources.TrafficSplit) error {
	removeCluster := func(route resources.Route, clusterName string, split resources.TrafficSplit) error {
		cluster, ok := xds.Clusters[clusterName]
		if !ok {
			return fmt.Errorf("can't find cluster for route %s cluster %s route %+v", route.RouteName, clusterName, route)
		}
		delete(cluster.Routes, resources.RouteVersionKey{RouteName: route.RouteName, ModelName: split.ModelName, Version: split.ModelVersion})
		if len(cluster.Routes) == 0 {
			delete(xds.Clusters, clusterName)
		}
		return nil
	}

	err := removeCluster(route, cluster.HttpCluster, cluster)
	if err != nil {
		return err
	}
	err = removeCluster(route, cluster.GrpcCluster, cluster)
	if err != nil {
		return err
	}
	return nil
}
