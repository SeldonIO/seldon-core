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
	Listeners map[string]resources.Listener
	Routes    map[string]resources.Route
	Clusters  map[string]resources.Cluster
	logger    logrus.FieldLogger
}

func NewSeldonXDSCache(logger logrus.FieldLogger) *SeldonXDSCache {
	return &SeldonXDSCache{
		Listeners: make(map[string]resources.Listener),
		Clusters:  make(map[string]resources.Cluster),
		Routes:    make(map[string]resources.Route),
		logger:    logger.WithField("source", "XDSCache"),
	}
}

func (xds *SeldonXDSCache) ClusterContents() []types.Resource {
	var r []types.Resource

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

	var routesArray []resources.Route
	for _, r := range xds.Routes { //This could be very large as is equal to number of models (100k?)
		routesArray = append(routesArray, r)
	}

	return []types.Resource{resources.MakeRoute(routesArray)}
}

func (xds *SeldonXDSCache) ListenerContents() []types.Resource {
	var r []types.Resource

	for _, l := range xds.Listeners {
		r = append(r, resources.MakeHTTPListener(l.Name, l.Address, l.Port))
	}

	return r
}

//Note: We don;t use endpoints at present as Envoy does not allow strict_dns with EDS
func (xds *SeldonXDSCache) EndpointsContents() []types.Resource {
	var r []types.Resource

	for _, c := range xds.Clusters {
		endpoints := make([]resources.Endpoint, 0, len(c.Endpoints))
		for _, value := range c.Endpoints {
			endpoints = append(endpoints, value)
		}
		r = append(r, resources.MakeEndpoint(c.Name, endpoints))
	}

	return r
}

func (xds *SeldonXDSCache) AddListener(name string) {
	xds.Listeners[name] = resources.Listener{
		Name:    name,
		Address: DefaultListenerAddress,
		Port:    DefaultListenerPort,
	}
}

func (xds *SeldonXDSCache) AddRouteClusterTraffic(modelName string, modelVersion uint32, trafficPercent uint32, httpClusterName string, grpcClusterName string, logPayloads bool) {
	route, ok := xds.Routes[modelName]
	if !ok {
		route = resources.Route{
			ModelName:   modelName,
			LogPayloads: logPayloads,
		}
	}
	// Always log payloads if any version wants it - so during a rolling update if one wants it then it will done
	if logPayloads {
		route.LogPayloads = true
	}

	clusterTraffic := resources.TrafficSplits{
		ModelName:      modelName,
		ModelVersion:   modelVersion,
		TrafficPercent: trafficPercent,
		HttpCluster:    httpClusterName,
		GrpcCluster:    grpcClusterName,
	}

	route.Clusters = append(route.Clusters, clusterTraffic)
	xds.Routes[modelName] = route
}

func (xds *SeldonXDSCache) AddCluster(name string, modelName string, isGrpc bool) {
	cluster, ok := xds.Clusters[name]
	if !ok {
		cluster = resources.Cluster{
			Name:      name,
			Endpoints: make(map[string]resources.Endpoint),
			Routes:    make(map[string]bool),
			Grpc:      isGrpc,
		}
	}
	cluster.Routes[modelName] = true
	xds.Clusters[name] = cluster
}

func (xds *SeldonXDSCache) RemoveRoute(modelName string) error {
	logger := xds.logger.WithField("func", "RemoveRoute")
	logger.Infof("Remove routes for model %s", modelName)
	route, ok := xds.Routes[modelName]
	if !ok {
		logger.Warnf("No route found for model %s", modelName)
		return nil
	}
	delete(xds.Routes, modelName)
	for _, cluster := range route.Clusters {
		httpCluster, ok := xds.Clusters[cluster.HttpCluster]
		if !ok {
			return fmt.Errorf("Can't find http cluster for model %s route %+v", modelName, route)
		}
		delete(httpCluster.Routes, route.ModelName)
		if len(httpCluster.Routes) == 0 {
			delete(xds.Clusters, cluster.HttpCluster)
		}

		grpcCluster, ok := xds.Clusters[cluster.GrpcCluster]
		if !ok {
			return fmt.Errorf("Can't find grpc cluster for model %s", modelName)
		}
		delete(grpcCluster.Routes, route.ModelName)
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
