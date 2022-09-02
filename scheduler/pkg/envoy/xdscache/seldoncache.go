package xdscache

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"
)

const (
	defaultListenerName           = "seldon_http"
	defaultListenerAddress        = "0.0.0.0"
	defaultListenerPort    uint32 = 9000

	mirrorListenerName           = "seldon_mirrors"
	mirrorListenerAddress        = "0.0.0.0"
	mirrorListenerPort    uint32 = 9001
)

type SeldonXDSCache struct {
	Listeners              map[string]resources.Listener
	Routes                 map[string]resources.Route
	Clusters               map[string]resources.Cluster
	Pipelines              map[string]resources.PipelineRoute
	PipelineGatewayDetails *PipelineGatewayDetails
	logger                 logrus.FieldLogger
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
		PipelineGatewayDetails: pipelineGatewayDetails,
		logger:                 logger.WithField("source", "XDSCache"),
	}
}

func (xds *SeldonXDSCache) ClusterContents() []types.Resource {
	var r []types.Resource

	//Add pipeline gateway clusters
	xds.logger.Infof("Add http pipeline cluster %s host:%s port:%d", resources.PipelineGatewayHttpClusterName, xds.PipelineGatewayDetails.Host, xds.PipelineGatewayDetails.HttpPort)
	r = append(r, resources.MakeCluster(resources.PipelineGatewayHttpClusterName, []resources.Endpoint{
		{
			UpstreamHost: xds.PipelineGatewayDetails.Host,
			UpstreamPort: uint32(xds.PipelineGatewayDetails.HttpPort),
		},
	}, false))
	xds.logger.Infof("Add grpc pipeline cluster %s host:%s port:%d", resources.PipelineGatewayGrpcClusterName, xds.PipelineGatewayDetails.Host, xds.PipelineGatewayDetails.GrpcPort)
	r = append(r, resources.MakeCluster(resources.PipelineGatewayGrpcClusterName, []resources.Endpoint{
		{
			UpstreamHost: xds.PipelineGatewayDetails.Host,
			UpstreamPort: uint32(xds.PipelineGatewayDetails.GrpcPort),
		},
	}, true))

	// Add Mirror clusters
	xds.logger.Infof("Add http mirror cluster %s host:%s port:%d", resources.MirrorHttpClusterName, mirrorListenerAddress, mirrorListenerPort)
	r = append(r, resources.MakeCluster(resources.MirrorHttpClusterName, []resources.Endpoint{
		{
			UpstreamHost: mirrorListenerAddress,
			UpstreamPort: mirrorListenerPort,
		},
	}, false))
	xds.logger.Infof("Add grpc mirror cluster %s host:%s port:%d", resources.MirrorGrpcClusterName, mirrorListenerAddress, mirrorListenerPort)
	r = append(r, resources.MakeCluster(resources.MirrorGrpcClusterName, []resources.Endpoint{
		{
			UpstreamHost: mirrorListenerAddress,
			UpstreamPort: mirrorListenerPort,
		},
	}, true))

	for _, c := range xds.Clusters {
		endpoints := make([]resources.Endpoint, 0, len(c.Endpoints))
		for _, value := range c.Endpoints { // Likely to be small (<100?) as is number of model replicas
			endpoints = append(endpoints, value)
		}
		r = append(r, resources.MakeCluster(c.Name, endpoints, c.Grpc))
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

	for _, l := range xds.Listeners {
		r = append(r, resources.MakeHTTPListener(l.Name, l.Address, l.Port, l.RouteConfigurationName))
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
