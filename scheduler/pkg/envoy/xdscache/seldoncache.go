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

package xdscache

import (
	"fmt"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/sirupsen/logrus"

	seldontls "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/resources"
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
	Listeners              map[string]resources.Listener
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
		Listeners:              make(map[string]resources.Listener),
		Clusters:               make(map[string]resources.Cluster),
		Routes:                 make(map[string]resources.Route),
		Pipelines:              make(map[string]resources.PipelineRoute),
		Secrets:                make(map[string]resources.Secret),
		PipelineGatewayDetails: pipelineGatewayDetails,
		logger:                 logger.WithField("source", "XDSCache"),
	}
}

func (xds *SeldonXDSCache) ClusterContents() []types.Resource {
	var r []types.Resource

	var clientSecret *resources.Secret
	if xds.TLSActive {
		if secret, ok := xds.Secrets[EnvoyUpstreamClientCertName]; ok {
			clientSecret = &secret
		}
	}

	//Add pipeline gateway clusters
	xds.logger.Infof("Add http pipeline cluster %s host:%s port:%d", resources.PipelineGatewayHttpClusterName, xds.PipelineGatewayDetails.Host, xds.PipelineGatewayDetails.HttpPort)
	r = append(r, resources.MakeCluster(resources.PipelineGatewayHttpClusterName, []resources.Endpoint{
		{
			UpstreamHost: xds.PipelineGatewayDetails.Host,
			UpstreamPort: uint32(xds.PipelineGatewayDetails.HttpPort),
		},
	}, false, clientSecret))
	xds.logger.Infof("Add grpc pipeline cluster %s host:%s port:%d", resources.PipelineGatewayGrpcClusterName, xds.PipelineGatewayDetails.Host, xds.PipelineGatewayDetails.GrpcPort)
	r = append(r, resources.MakeCluster(resources.PipelineGatewayGrpcClusterName, []resources.Endpoint{
		{
			UpstreamHost: xds.PipelineGatewayDetails.Host,
			UpstreamPort: uint32(xds.PipelineGatewayDetails.GrpcPort),
		},
	}, true, clientSecret))

	// Add Mirror clusters
	xds.logger.Infof("Add http mirror cluster %s host:%s port:%d", resources.MirrorHttpClusterName, mirrorListenerAddress, mirrorListenerPort)
	r = append(r, resources.MakeCluster(resources.MirrorHttpClusterName, []resources.Endpoint{
		{
			UpstreamHost: mirrorListenerAddress,
			UpstreamPort: mirrorListenerPort,
		},
	}, false, nil))
	xds.logger.Infof("Add grpc mirror cluster %s host:%s port:%d", resources.MirrorGrpcClusterName, mirrorListenerAddress, mirrorListenerPort)
	r = append(r, resources.MakeCluster(resources.MirrorGrpcClusterName, []resources.Endpoint{
		{
			UpstreamHost: mirrorListenerAddress,
			UpstreamPort: mirrorListenerPort,
		},
	}, true, nil))

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
	routesArray := make([]*resources.Route, len(xds.Routes))
	rIdx := 0
	for _, r := range xds.Routes { // This could be very large as is equal to number of models (100k?)
		modelRoute := r
		routesArray[rIdx] = &modelRoute
		rIdx++
	}

	pipelinesArray := make([]*resources.PipelineRoute, len(xds.Pipelines))
	pIdx := 0
	for _, r := range xds.Pipelines { // Likely to be less pipelines than models
		pipelineRoute := r
		pipelinesArray[pIdx] = &pipelineRoute
		pIdx++
	}

	defaultRoutes, mirrorRoutes := resources.MakeRoute(routesArray, pipelinesArray)
	return []types.Resource{defaultRoutes, mirrorRoutes}
}

func (xds *SeldonXDSCache) ListenerContents() []types.Resource {
	var r []types.Resource

	var serverSecret *resources.Secret
	if xds.TLSActive {
		if secret, ok := xds.Secrets[EnvoyDownstreamServerCertName]; ok {
			serverSecret = &secret
		}
	}

	for _, l := range xds.Listeners {
		r = append(r, resources.MakeHTTPListener(l.Name, l.Address, l.Port, l.RouteConfigurationName, serverSecret))
	}

	return r
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

func (xds *SeldonXDSCache) AddPipelineRoute(routeName string, pipelineName string, trafficWeight uint32, isMirror bool) {
	pipelineRoute, ok := xds.Pipelines[routeName]
	if !ok {
		if isMirror {
			xds.Pipelines[routeName] = resources.PipelineRoute{
				RouteName: routeName,
				Mirrors: []resources.PipelineTrafficSplits{
					{
						PipelineName:  pipelineName,
						TrafficWeight: trafficWeight,
					},
				},
			}
		} else {
			xds.Pipelines[routeName] = resources.PipelineRoute{
				RouteName: routeName,
				Clusters: []resources.PipelineTrafficSplits{
					{
						PipelineName:  pipelineName,
						TrafficWeight: trafficWeight,
					},
				},
			}
		}
	} else {
		if isMirror {
			pipelineRoute.Mirrors = append(pipelineRoute.Mirrors, resources.PipelineTrafficSplits{
				PipelineName:  pipelineName,
				TrafficWeight: trafficWeight,
			})
		} else {
			pipelineRoute.Clusters = append(pipelineRoute.Clusters, resources.PipelineTrafficSplits{
				PipelineName:  pipelineName,
				TrafficWeight: trafficWeight,
			})
		}
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

func (xds *SeldonXDSCache) AddListeners() {
	xds.Listeners[defaultListenerName] = resources.Listener{
		Name:                   defaultListenerName,
		Address:                defaultListenerAddress,
		Port:                   defaultListenerPort,
		RouteConfigurationName: resources.DefaultRouteConfigurationName,
	}
	xds.Listeners[mirrorListenerName] = resources.Listener{
		Name:                   mirrorListenerName,
		Address:                mirrorListenerAddress,
		Port:                   mirrorListenerPort,
		RouteConfigurationName: resources.MirrorRouteConfigurationName,
	}
}

func (xds *SeldonXDSCache) AddRouteClusterTraffic(
	routeName string,
	modelName string,
	modelVersion uint32,
	trafficPercent uint32,
	httpClusterName string,
	grpcClusterName string,
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

	clusterTraffic := resources.TrafficSplits{
		ModelName:     modelName,
		ModelVersion:  modelVersion,
		TrafficWeight: trafficPercent,
		HttpCluster:   httpClusterName,
		GrpcCluster:   grpcClusterName,
	}
	if isMirror {
		route.Mirrors = append(route.Mirrors, clusterTraffic)
	} else {
		route.Clusters = append(route.Clusters, clusterTraffic)
	}
	xds.Routes[routeName] = route
}

func (xds *SeldonXDSCache) AddCluster(
	name string,
	routeName string,
	modelName string,
	modelVersion uint32,
	isGrpc bool,
) {
	cluster, ok := xds.Clusters[name]
	if !ok {
		cluster = resources.Cluster{
			Name:      name,
			Endpoints: make(map[string]resources.Endpoint),
			Routes:    make(map[resources.RouteVersionKey]bool),
			Grpc:      isGrpc,
		}
	}
	cluster.Routes[resources.RouteVersionKey{RouteName: routeName, ModelName: modelName, Version: modelVersion}] = true
	xds.Clusters[name] = cluster
}

func (xds *SeldonXDSCache) removeRouteFromCluster(routeName string, route resources.Route, cluster resources.TrafficSplits) error {
	httpCluster, ok := xds.Clusters[cluster.HttpCluster]
	if !ok {
		return fmt.Errorf("Can't find http cluster for route %s cluster %s route %+v", routeName, cluster.HttpCluster, route)
	}
	delete(httpCluster.Routes, resources.RouteVersionKey{RouteName: routeName, ModelName: cluster.ModelName, Version: cluster.ModelVersion})
	if len(httpCluster.Routes) == 0 {
		delete(xds.Clusters, cluster.HttpCluster)
	}

	grpcCluster, ok := xds.Clusters[cluster.GrpcCluster]
	if !ok {
		return fmt.Errorf("Can't find grpc cluster for route %s cluster %s route %+v", routeName, cluster.GrpcCluster, route)
	}
	delete(grpcCluster.Routes, resources.RouteVersionKey{RouteName: routeName, ModelName: cluster.ModelName, Version: cluster.ModelVersion})
	if len(grpcCluster.Routes) == 0 {
		delete(xds.Clusters, cluster.GrpcCluster)
	}
	return nil
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
		err := xds.removeRouteFromCluster(routeName, route, cluster)
		if err != nil {
			return err
		}
	}
	for _, mirror := range route.Mirrors {
		err := xds.removeRouteFromCluster(routeName, route, mirror)
		if err != nil {
			return err
		}
	}
	return nil
}

func (xds *SeldonXDSCache) AddEndpoint(clusterName, upstreamHost string, upstreamPort uint32) {
	cluster := xds.Clusters[clusterName]
	k := fmt.Sprintf("%s:%d", upstreamHost, upstreamPort)
	cluster.Endpoints[k] = resources.Endpoint{
		UpstreamHost: upstreamHost,
		UpstreamPort: upstreamPort,
	}

	xds.Clusters[clusterName] = cluster
}
