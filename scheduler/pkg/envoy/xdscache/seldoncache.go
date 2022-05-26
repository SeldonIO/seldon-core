package xdscache

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"
)

const (
	DefaultListenerAddress        = "0.0.0.0"
	DefaultListenerPort    uint32 = 9000
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
	r = append(r, resources.MakeCluster(resources.PipelineGatewayGrpcClusterName, []resources.Endpoint{
		{
			UpstreamHost: xds.PipelineGatewayDetails.Host,
			UpstreamPort: uint32(xds.PipelineGatewayDetails.GrpcPort),
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

	return []types.Resource{resources.MakeRoute(routesArray, pipelinesArray)}
}

func (xds *SeldonXDSCache) ListenerContents() []types.Resource {
	var r []types.Resource

	for _, l := range xds.Listeners {
		r = append(r, resources.MakeHTTPListener(l.Name, l.Address, l.Port))
	}

	return r
}

//Note: We don;t use endpoints at present as Envoy does not allow strict_dns with EDS
//func (xds *SeldonXDSCache) EndpointsContents() []types.Resource {
//	var r []types.Resource
//
//	for _, c := range xds.Clusters {
//		endpoints := make([]resources.Endpoint, 0, len(c.Endpoints))
//		for _, value := range c.Endpoints {
//			endpoints = append(endpoints, value)
//		}
//		r = append(r, resources.MakeEndpoint(c.Name, endpoints))
//	}
//
//	return r
//}

func (xds *SeldonXDSCache) AddPipelineRoute(pipelineName string) {
	xds.Pipelines[pipelineName] = resources.PipelineRoute{PipelineName: pipelineName}
}

func (xds *SeldonXDSCache) RemovePipelineRoute(pipelineName string) {
	delete(xds.Pipelines, pipelineName)
}

func (xds *SeldonXDSCache) AddListener(name string) {
	xds.Listeners[name] = resources.Listener{
		Name:    name,
		Address: DefaultListenerAddress,
		Port:    DefaultListenerPort,
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
	route.Clusters = append(route.Clusters, clusterTraffic)
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
