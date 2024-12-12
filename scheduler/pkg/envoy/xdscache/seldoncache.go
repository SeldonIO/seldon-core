/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package xdscache

import (
	"context"
	"fmt"

	discoveryv3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/stream/v3"
	"github.com/sirupsen/logrus"

	seldontls "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

const (
	defaultListenerName           = "seldon_http"
	defaultListenerAddress        = "0.0.0.0"
	defaultListenerPort    uint32 = 9000

	mirrorListenerName           = "seldon_mirrors"
	mirrorListenerAddress        = "0.0.0.0"
	mirrorListenerPort    uint32 = 9001

	EnvoyDownstreamServerCertName = "downstream_server"
	EnvoyDownstreamClientCertName = "downstream_client"
	EnvoyUpstreamServerCertName   = "upstream_server"
	EnvoyUpstreamClientCertName   = "upstream_client"
)

type SeldonXDSCache struct {
	muxCache *cache.MuxCache
	cds      *cache.LinearCache
	lds      *cache.LinearCache
	rds      *cache.LinearCache
	sds      *cache.LinearCache

	Clusters               map[string]Cluster
	ClustersForRemoval     map[string]bool
	Pipelines              map[string]PipelineRoute
	PipelineGatewayDetails *PipelineGatewayDetails
	Routes                 map[string]Route
	Secrets                map[string]Secret
	logger                 logrus.FieldLogger
	TLSActive              bool
}

type PipelineGatewayDetails struct {
	Host     string
	HttpPort int
	GrpcPort int
}

func NewSeldonXDSCache(logger logrus.FieldLogger, pipelineGatewayDetails *PipelineGatewayDetails) (*SeldonXDSCache, error) {
	xdsCache := &SeldonXDSCache{
		Clusters:               make(map[string]Cluster),
		ClustersForRemoval:     make(map[string]bool),
		Pipelines:              make(map[string]PipelineRoute),
		PipelineGatewayDetails: pipelineGatewayDetails,
		Routes:                 make(map[string]Route),
		Secrets:                make(map[string]Secret),
		logger:                 logger.WithField("source", "SeldonXDSCache"),
	}
	err := xdsCache.init()
	if err != nil {
		return nil, err
	}
	return xdsCache, nil
}

func (xds *SeldonXDSCache) CreateWatch(req *cache.Request, stream stream.StreamState, responseChan chan cache.Response) (cancel func()) {
	return xds.muxCache.CreateWatch(req, stream, responseChan)
}

func (xds *SeldonXDSCache) CreateDeltaWatch(req *cache.DeltaRequest, stream stream.StreamState, responseChan chan cache.DeltaResponse) (cancel func()) {
	return xds.muxCache.CreateDeltaWatch(req, stream, responseChan)
}

func (xds *SeldonXDSCache) Fetch(ctx context.Context, req *cache.Request) (cache.Response, error) {
	return xds.muxCache.Fetch(ctx, req)
}

func (xds *SeldonXDSCache) init() error {
	logger := xds.logger.WithField("source", "LinearCache")

	secretTypeURL := resource.SecretType
	sds := cache.NewLinearCache(secretTypeURL, cache.WithLogger(logger))
	xds.sds = sds

	listenerTypeURL := resource.ListenerType
	lds := cache.NewLinearCache(listenerTypeURL, cache.WithLogger(logger))
	xds.lds = lds

	clusterTypeURL := resource.ClusterType
	cds := cache.NewLinearCache(clusterTypeURL, cache.WithLogger(logger))
	xds.cds = cds

	routeTypeURL := resource.RouteType
	rds := cache.NewLinearCache(routeTypeURL, cache.WithLogger(logger))
	xds.rds = rds

	xds.muxCache = &cache.MuxCache{
		Classify: func(req *discoveryv3.DiscoveryRequest) string {
			return req.GetTypeUrl()
		},
		ClassifyDelta: func(req *discoveryv3.DeltaDiscoveryRequest) string {
			return req.GetTypeUrl()
		},
		Caches: map[string]cache.Cache{
			secretTypeURL:   sds,
			listenerTypeURL: lds,
			clusterTypeURL:  cds,
			routeTypeURL:    rds,
		},
	}

	err := xds.SetupTLS()
	if err != nil {
		return err
	}

	xds.AddPermanentListeners()
	xds.AddPermanentClusters()
	return nil
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
		xds.createSecret(EnvoyUpstreamClientCertName, EnvoyUpstreamServerCertName, tlsUpstreamClient)

		// Envoy listener - external calls to Seldon
		logger.Info("Downstream TLS active")
		tlsDownstreamServer, err := seldontls.NewCertificateStore(seldontls.Prefix(seldontls.EnvSecurityPrefixEnvoyDownstreamServer),
			seldontls.ValidationPrefix(seldontls.EnvSecurityPrefixEnvoyDownstreamClient))
		if err != nil {
			return err
		}
		xds.createSecret(EnvoyDownstreamServerCertName, EnvoyDownstreamClientCertName, tlsDownstreamServer)
	}
	return nil
}

func (xds *SeldonXDSCache) createSecret(name string, validationSecretName string, certificate *seldontls.CertificateStore) {
	seldonSecret := Secret{
		Name:                 name,
		ValidationSecretName: validationSecretName,
		Certificate:          certificate,
	}

	xds.Secrets[name] = seldonSecret
	secrets := MakeSecretResource(seldonSecret.Name, seldonSecret.ValidationSecretName, seldonSecret.Certificate)
	for _, secret := range secrets {
		xds.sds.UpdateResource(secret.Name, secret)
	}
}

func (xds *SeldonXDSCache) AddPermanentListeners() {
	var serverSecret *Secret
	if xds.TLSActive {
		if secret, ok := xds.Secrets[EnvoyDownstreamServerCertName]; ok {
			serverSecret = &secret
		}
	}
	defaultListener := MakeHTTPListener(defaultListenerName, defaultListenerAddress, defaultListenerPort, DefaultRouteConfigurationName, serverSecret)
	xds.lds.UpdateResource(defaultListener.Name, defaultListener)
	mirrorListener := MakeHTTPListener(mirrorListenerName, mirrorListenerAddress, mirrorListenerPort, MirrorRouteConfigurationName, serverSecret)
	xds.lds.UpdateResource(mirrorListener.Name, mirrorListener)
}

func (xds *SeldonXDSCache) AddPermanentClusters() {
	var clientSecret *Secret
	if xds.TLSActive {
		if secret, ok := xds.Secrets[EnvoyUpstreamClientCertName]; ok {
			clientSecret = &secret
		}
	}

	// Add pipeline gateway clusters
	pipelineGatewayCluster := MakeCluster(PipelineGatewayHttpClusterName, []Endpoint{
		{
			UpstreamHost: xds.PipelineGatewayDetails.Host,
			UpstreamPort: uint32(xds.PipelineGatewayDetails.HttpPort),
		},
	}, false, clientSecret)
	xds.cds.UpdateResource(pipelineGatewayCluster.Name, pipelineGatewayCluster)

	pipelineGatewayGrpcCluster := MakeCluster(PipelineGatewayGrpcClusterName, []Endpoint{
		{
			UpstreamHost: xds.PipelineGatewayDetails.Host,
			UpstreamPort: uint32(xds.PipelineGatewayDetails.GrpcPort),
		},
	}, true, clientSecret)
	xds.cds.UpdateResource(pipelineGatewayGrpcCluster.Name, pipelineGatewayGrpcCluster)

	// Add Mirror clusters
	mirrorHttpCluster := MakeCluster(MirrorHttpClusterName, []Endpoint{
		{
			UpstreamHost: mirrorListenerAddress,
			UpstreamPort: mirrorListenerPort,
		},
	}, false, nil)
	xds.cds.UpdateResource(mirrorHttpCluster.Name, mirrorHttpCluster)

	mirrorGprcCluster := MakeCluster(MirrorGrpcClusterName, []Endpoint{
		{
			UpstreamHost: mirrorListenerAddress,
			UpstreamPort: mirrorListenerPort,
		},
	}, true, nil)
	xds.cds.UpdateResource(mirrorGprcCluster.Name, mirrorGprcCluster)
}

func (xds *SeldonXDSCache) ClusterContents() []types.Resource {
	var r []types.Resource
	for _, cluster := range xds.cds.GetResources() {
		r = append(r, cluster)
	}
	return r
}

func (xds *SeldonXDSCache) RouteContents() []types.Resource {
	var r []types.Resource
	for _, route := range xds.rds.GetResources() {
		r = append(r, route)
	}
	return r
}

func (xds *SeldonXDSCache) ListenerContents() []types.Resource {
	var r []types.Resource
	for _, listener := range xds.lds.GetResources() {
		r = append(r, listener)
	}
	return r
}

func (xds *SeldonXDSCache) SecretContents() []types.Resource {
	var r []types.Resource
	for _, secret := range xds.sds.GetResources() {
		r = append(r, secret)
	}
	return r
}

func (xds *SeldonXDSCache) UpdateRoutes() {
	defaultRoutes, mirrorRoutes := MakeRoutes(xds.Routes, xds.Pipelines)
	xds.rds.UpdateResource(defaultRoutes.Name, defaultRoutes)
	xds.rds.UpdateResource(mirrorRoutes.Name, mirrorRoutes)

	for clusterName := range xds.ClustersForRemoval {
		xds.cds.DeleteResource(clusterName)
	}
}

func (xds *SeldonXDSCache) AddPipelineRoute(routeName string, trafficSplits []PipelineTrafficSplit, mirror *PipelineTrafficSplit) {
	xds.RemovePipelineRoute(routeName)
	pipelineRoute, ok := xds.Pipelines[routeName]
	if !ok {
		xds.Pipelines[routeName] = PipelineRoute{
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

func (xds *SeldonXDSCache) AddRouteClusterTraffic(
	routeName, modelName, httpClusterName, grpcClusterName string,
	modelVersion uint32,
	trafficPercent uint32,
	logPayloads bool,
	isMirror bool,
) {
	route, ok := xds.Routes[routeName]
	if !ok {
		route = Route{
			RouteName:   routeName,
			LogPayloads: logPayloads,
		}
	}

	// Always log payloads if any version wants it - so during a rolling update if one wants it then it will done
	if logPayloads {
		route.LogPayloads = true
	}

	clusterTraffic := TrafficSplit{
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

	routeVersionKey := RouteVersionKey{RouteName: routeName, ModelName: modelName, Version: modelVersion}

	httpCluster, ok := xds.Clusters[httpClusterName]
	if !ok {
		httpCluster = Cluster{
			Name:      httpClusterName,
			Endpoints: make([]Endpoint, 0),
			Routes:    make(map[RouteVersionKey]bool),
			Grpc:      false,
		}
	}

	grpcCluster, ok := xds.Clusters[grpcClusterName]
	if !ok {
		grpcCluster = Cluster{
			Name:      grpcClusterName,
			Endpoints: make([]Endpoint, 0),
			Routes:    make(map[RouteVersionKey]bool),
			Grpc:      true,
		}
	}

	httpEndpoints := make([]Endpoint, 0)
	grpcEndpoints := make([]Endpoint, 0)
	for _, replicaIdx := range assignment {
		replica, ok := server.Replicas[replicaIdx]
		if !ok {
			logger.Warnf("Invalid replica index %d for server %s", replicaIdx, server.Name)
		} else {
			httpEndpoint := Endpoint{
				UpstreamHost: replica.GetInferenceSvc(),
				UpstreamPort: uint32(replica.GetInferenceHttpPort()),
			}
			httpEndpoints = append(httpEndpoints, httpEndpoint)

			grpcEndpoint := Endpoint{
				UpstreamHost: replica.GetInferenceSvc(),
				UpstreamPort: uint32(replica.GetInferenceGrpcPort()),
			}
			grpcEndpoints = append(grpcEndpoints, grpcEndpoint)
		}
	}

	httpCluster.Endpoints = httpEndpoints
	xds.Clusters[httpClusterName] = httpCluster
	httpCluster.Routes[routeVersionKey] = true
	delete(xds.ClustersForRemoval, httpClusterName)

	grpcCluster.Endpoints = grpcEndpoints
	xds.Clusters[grpcClusterName] = grpcCluster
	grpcCluster.Routes[routeVersionKey] = true
	delete(xds.ClustersForRemoval, grpcClusterName)

	var clientSecret *Secret
	if xds.TLSActive {
		if secret, ok := xds.Secrets[EnvoyUpstreamClientCertName]; ok {
			clientSecret = &secret
		}
	}

	envoyHttpCluster := MakeCluster(httpClusterName, httpEndpoints, false, clientSecret)
	xds.cds.UpdateResource(envoyHttpCluster.Name, envoyHttpCluster)
	envoyGrpcCluster := MakeCluster(grpcClusterName, grpcEndpoints, true, clientSecret)
	xds.cds.UpdateResource(envoyGrpcCluster.Name, envoyGrpcCluster)
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

func (xds *SeldonXDSCache) removeRouteFromCluster(route Route, cluster TrafficSplit) error {
	removeCluster := func(route Route, clusterName string, split TrafficSplit) error {
		cluster, ok := xds.Clusters[clusterName]
		if !ok {
			return fmt.Errorf("can't find cluster for route %s cluster %s route %+v", route.RouteName, clusterName, route)
		}
		delete(cluster.Routes, RouteVersionKey{RouteName: route.RouteName, ModelName: split.ModelName, Version: split.ModelVersion})
		if len(cluster.Routes) == 0 {
			delete(xds.Clusters, clusterName)
			xds.ClustersForRemoval[clusterName] = true
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
