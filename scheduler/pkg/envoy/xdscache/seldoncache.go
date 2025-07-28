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
	"math"
	"math/rand"
	"strconv"

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
	// https://github.com/envoyproxy/go-control-plane?tab=readme-ov-file#resource-caching
	// each envoy resourece is managed independently, using ADS (aggregated discovery service), so
	// updates can be sequenced in a way that reduces the susceptibility to "no cluster found"
	// responses
	muxCache      *cache.MuxCache
	snapshotCache cache.SnapshotCache // routes
	Cds           *cache.LinearCache  // clusters
	Lds           *cache.LinearCache  // listener
	Sds           *cache.LinearCache  // secrets

	Clusters               map[string]Cluster
	Pipelines              map[string]PipelineRoute
	Routes                 map[string]Route
	PipelineGatewayDetails *PipelineGatewayDetails
	TlsActive              bool
	Secrets                map[string]Secret
	clustersToAdd          map[string]struct{}
	clustersToRemove       map[string]struct{}
	logger                 logrus.FieldLogger
	snapshotVersion        int64
	config                 *EnvoyConfig
}

type PipelineGatewayDetails struct {
	Host     string
	HttpPort int
	GrpcPort int
}

type EnvoyConfig struct {
	AccessLogPath             string
	IncludeSuccessfulRequests bool
	EnableAccessLog           bool
}

func NewSeldonXDSCache(logger logrus.FieldLogger, pipelineGatewayDetails *PipelineGatewayDetails, config *EnvoyConfig) (*SeldonXDSCache, error) {
	xdsCache := &SeldonXDSCache{
		Clusters:               make(map[string]Cluster),
		Pipelines:              make(map[string]PipelineRoute),
		Routes:                 make(map[string]Route),
		PipelineGatewayDetails: pipelineGatewayDetails,
		Secrets:                make(map[string]Secret),
		clustersToAdd:          make(map[string]struct{}),
		clustersToRemove:       make(map[string]struct{}),
		logger:                 logger.WithField("source", "SeldonXDSCache"),
		snapshotVersion:        rand.Int63n(1000),
		config:                 config,
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

func (xds *SeldonXDSCache) newSnapshotVersion() string {
	// Reset the snapshotVersion if it ever hits max size.
	if xds.snapshotVersion == math.MaxInt64 {
		xds.snapshotVersion = 0
	}

	// Increment the snapshot version & return as string.
	xds.snapshotVersion++
	return strconv.FormatInt(xds.snapshotVersion, 10)
}

func (xds *SeldonXDSCache) init() error {
	const snapshotType = "snapshot"
	linearLogger := xds.logger.WithField("source", "LinearCache")
	snapshotLogger := xds.logger.WithField("source", "SnapshotCache")

	secretTypeURL := resource.SecretType
	sds := cache.NewLinearCache(secretTypeURL, cache.WithLogger(linearLogger))
	xds.Sds = sds

	listenerTypeURL := resource.ListenerType
	lds := cache.NewLinearCache(listenerTypeURL, cache.WithLogger(linearLogger))
	xds.Lds = lds

	clusterTypeURL := resource.ClusterType
	cds := cache.NewLinearCache(clusterTypeURL, cache.WithLogger(linearLogger))
	xds.Cds = cds

	snapshotCache := cache.NewSnapshotCache(true, cache.IDHash{}, snapshotLogger)
	xds.snapshotCache = snapshotCache

	classify := func(typeUrl string) string {
		switch typeUrl {
		case secretTypeURL:
			return secretTypeURL
		case listenerTypeURL:
			return listenerTypeURL
		case clusterTypeURL:
			return clusterTypeURL
		default:
			return snapshotType
		}
	}

	xds.muxCache = &cache.MuxCache{
		Classify: func(req *cache.Request) string {
			return classify(req.GetTypeUrl())
		},
		ClassifyDelta: func(req *cache.DeltaRequest) string {
			return classify(req.GetTypeUrl())
		},
		Caches: map[string]cache.Cache{
			secretTypeURL:   sds,
			listenerTypeURL: lds,
			clusterTypeURL:  cds,
			snapshotType:    snapshotCache,
		},
	}

	err := xds.setupTLS()
	if err != nil {
		return err
	}
	err = xds.addPermanentListeners()
	if err != nil {
		return err
	}
	err = xds.addPermanentClusters()
	if err != nil {
		return err
	}
	return nil
}

func (xds *SeldonXDSCache) setupTLS() error {
	logger := xds.logger.WithField("func", "SetupTLS")
	protocol := seldontls.GetSecurityProtocolFromEnv(seldontls.EnvSecurityPrefixEnvoy)
	if protocol == seldontls.SecurityProtocolSSL {
		xds.TlsActive = true

		// Envoy client to talk to agent or Pipelinegateway
		logger.Info("Upstream TLS active")
		tlsUpstreamClient, err := seldontls.NewCertificateStore(seldontls.Prefix(seldontls.EnvSecurityPrefixEnvoyUpstreamClient),
			seldontls.ValidationPrefix(seldontls.EnvSecurityPrefixEnvoyUpstreamServer))
		if err != nil {
			return err
		}
		err = xds.createSecret(EnvoyUpstreamClientCertName, EnvoyUpstreamServerCertName, tlsUpstreamClient)
		if err != nil {
			return err
		}

		// Envoy listener - external calls to Seldon
		logger.Info("Downstream TLS active")
		tlsDownstreamServer, err := seldontls.NewCertificateStore(seldontls.Prefix(seldontls.EnvSecurityPrefixEnvoyDownstreamServer),
			seldontls.ValidationPrefix(seldontls.EnvSecurityPrefixEnvoyDownstreamClient))
		if err != nil {
			return err
		}
		err = xds.createSecret(EnvoyDownstreamServerCertName, EnvoyDownstreamClientCertName, tlsDownstreamServer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (xds *SeldonXDSCache) createSecret(name string, validationSecretName string, certificate *seldontls.CertificateStore) error {
	seldonSecret := Secret{
		Name:                 name,
		ValidationSecretName: validationSecretName,
		Certificate:          certificate,
	}

	xds.Secrets[name] = seldonSecret
	secrets := MakeSecretResource(seldonSecret.Name, seldonSecret.ValidationSecretName, seldonSecret.Certificate)
	for _, secret := range secrets {
		err := xds.Sds.UpdateResource(secret.Name, secret)
		if err != nil {
			return err
		}
	}
	return nil
}

func (xds *SeldonXDSCache) addPermanentListeners() error {
	var serverSecret *Secret
	if xds.TlsActive {
		if secret, ok := xds.Secrets[EnvoyDownstreamServerCertName]; ok {
			serverSecret = &secret
		}
	}
	resources := make(map[string]types.Resource)
	resources[defaultListenerName] = makeHTTPListener(defaultListenerName, defaultListenerAddress, defaultListenerPort, DefaultRouteConfigurationName, serverSecret, xds.config)
	resources[mirrorListenerName] = makeHTTPListener(mirrorListenerName, mirrorListenerAddress, mirrorListenerPort, MirrorRouteConfigurationName, serverSecret, xds.config)
	return xds.Lds.UpdateResources(resources, nil)
}

func (xds *SeldonXDSCache) addPermanentClusters() error {
	resources := make(map[string]types.Resource)

	// Add Mirror clusters
	mirrorHttpEndpointName := fmt.Sprintf("%s:%d", mirrorListenerAddress, mirrorListenerPort)
	mirrorHttpCluster := MakeCluster(MirrorHttpClusterName, map[string]Endpoint{
		mirrorHttpEndpointName: {
			UpstreamHost: mirrorListenerAddress,
			UpstreamPort: mirrorListenerPort,
		},
	}, false, nil)
	resources[mirrorHttpCluster.Name] = mirrorHttpCluster

	mirrorGrpcEndpointName := fmt.Sprintf("%s:%d", mirrorListenerAddress, mirrorListenerPort)
	mirrorGprcCluster := MakeCluster(MirrorGrpcClusterName, map[string]Endpoint{
		mirrorGrpcEndpointName: {
			UpstreamHost: mirrorListenerAddress,
			UpstreamPort: mirrorListenerPort,
		},
	}, true, nil)
	resources[mirrorGprcCluster.Name] = mirrorGprcCluster

	return xds.Cds.UpdateResources(resources, nil)
}

func (xds *SeldonXDSCache) ClusterResources() []types.Resource {
	var r []types.Resource
	for _, cluster := range xds.Cds.GetResources() {
		r = append(r, cluster)
	}
	return r
}

func (xds *SeldonXDSCache) RouteResources() []types.Resource {
	defaultRoutes, mirrorRoutes := makeRoutes(xds.Routes, xds.Pipelines)
	return []types.Resource{defaultRoutes, mirrorRoutes}
}

func (xds *SeldonXDSCache) UpdateRoutes(nodeId string) error {
	logger := xds.logger.WithField("func", "UpdateRoutes")
	// Create the snapshot that we'll serve to Envoy
	snapshot, err := cache.NewSnapshot(
		xds.newSnapshotVersion(), // version
		map[resource.Type][]types.Resource{
			resource.RouteType: xds.RouteResources(), // routes
		})
	if err != nil {
		logger.Errorf("could not create snapshot %+v", snapshot)
		return err
	}

	logger.Debugf("will serve snapshot %+v", snapshot)

	// Add the snapshot to the cache
	if err := xds.snapshotCache.SetSnapshot(context.Background(), nodeId, snapshot); err != nil {
		return err
	}
	return nil
}

func (xds *SeldonXDSCache) AddClusters() error {
	var clientSecret *Secret
	if xds.TlsActive {
		if secret, ok := xds.Secrets[EnvoyUpstreamClientCertName]; ok {
			clientSecret = &secret
		}
	}

	resources := make(map[string]types.Resource)
	for clusterName := range xds.clustersToAdd {
		cluster, ok := xds.Clusters[clusterName]
		if ok {
			resource := MakeCluster(cluster.Name, cluster.Endpoints, cluster.Grpc, clientSecret)
			resources[cluster.Name] = resource

		}
		delete(xds.clustersToAdd, clusterName)
	}

	return xds.Cds.UpdateResources(resources, nil)
}

func (xds *SeldonXDSCache) RemoveClusters() error {
	clustersToRemove := make([]string, 0)
	for clusterName := range xds.clustersToRemove {
		if xds.shouldRemoveCluster(clusterName) {
			clustersToRemove = append(clustersToRemove, clusterName)
		}
		delete(xds.clustersToRemove, clusterName)
	}

	return xds.Cds.UpdateResources(nil, clustersToRemove)
}

// updates are batched - always check if the state has changed
func (xds *SeldonXDSCache) shouldRemoveCluster(name string) bool {
	cluster, ok := xds.Clusters[name]
	return !ok || len(cluster.Routes) < 1
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
			Endpoints: make(map[string]Endpoint),
			Routes:    make(map[RouteVersionKey]bool),
			Grpc:      false,
		}
	}

	grpcCluster, ok := xds.Clusters[grpcClusterName]
	if !ok {
		grpcCluster = Cluster{
			Name:      grpcClusterName,
			Endpoints: make(map[string]Endpoint),
			Routes:    make(map[RouteVersionKey]bool),
			Grpc:      true,
		}
	}

	for _, replicaIdx := range assignment {
		replica, ok := server.Replicas[replicaIdx]
		if !ok {
			logger.Warnf("Invalid replica index %d for server %s", replicaIdx, server.Name)
		} else {
			httpEndpointName := fmt.Sprintf("%s:%d", replica.GetInferenceSvc(), replica.GetInferenceHttpPort())
			httpCluster.Endpoints[httpEndpointName] = Endpoint{
				UpstreamHost: replica.GetInferenceSvc(),
				UpstreamPort: uint32(replica.GetInferenceHttpPort()),
			}

			grpcEndpointName := fmt.Sprintf("%s:%d", replica.GetInferenceSvc(), replica.GetInferenceGrpcPort())
			grpcCluster.Endpoints[grpcEndpointName] = Endpoint{
				UpstreamHost: replica.GetInferenceSvc(),
				UpstreamPort: uint32(replica.GetInferenceGrpcPort()),
			}

		}
	}

	xds.Clusters[httpClusterName] = httpCluster
	httpCluster.Routes[routeVersionKey] = true
	xds.clustersToAdd[httpClusterName] = struct{}{}

	xds.Clusters[grpcClusterName] = grpcCluster
	grpcCluster.Routes[routeVersionKey] = true
	xds.clustersToAdd[grpcClusterName] = struct{}{}
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
			xds.clustersToRemove[clusterName] = struct{}{}
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
