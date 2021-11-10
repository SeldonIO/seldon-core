package xdscache

import (
	"fmt"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"
)

const (
	DefaultListenerAddress = "0.0.0.0"
	DefaultListenerPort uint32 = 9000
)

type SeldonXDSCache struct {
	Listeners map[string]resources.Listener
	Routes    map[string]resources.Route
	Clusters  map[string]resources.Cluster
	Endpoints map[string]resources.Endpoint
}

func (xds *SeldonXDSCache) ClusterContents() []types.Resource {
	var r []types.Resource

	for _, c := range xds.Clusters {
		endpoints := make([]resources.Endpoint, 0, len(c.Endpoints))
		for  _, value := range c.Endpoints { // Likely to be small (<100?) as is number of model replicas
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
		for  _, value := range c.Endpoints {
			endpoints = append(endpoints, value)
		}
		r = append(r, resources.MakeEndpoint(c.Name, endpoints))
	}

	return r
}


func (xds *SeldonXDSCache) AddListener(name string) {
	xds.Listeners[name] = resources.Listener{
		Name:       name,
		Address:    DefaultListenerAddress,
		Port:       DefaultListenerPort,
	}
}

func (xds *SeldonXDSCache) AddRoute(name, modelName string, httpClusterName string, grpcClusterName string) {
	xds.Routes[name] = resources.Route{
		Name:        name,
		Host:        modelName,
		HttpCluster: httpClusterName,
		GrpcCluster: grpcClusterName,
	}
}

func (xds *SeldonXDSCache) HasCluster(name string) bool {
	_, ok := xds.Clusters[name]
	return ok
}

func (xds *SeldonXDSCache) AddCluster(name string, route string, isGrpc bool) {
	cluster, ok := xds.Clusters[name]
	if !ok {
		cluster = resources.Cluster{
			Name: name,
			Endpoints: make(map[string]resources.Endpoint),
			Routes: make(map[string]bool),
			Grpc: isGrpc,
		}
	}
	cluster.Routes[route] = true
	xds.Clusters[name] = cluster
}


func (xds *SeldonXDSCache) RemoveRoute(modelName string) {
	route, ok := xds.Routes[modelName]
	if !ok {
		return
	}
	cluster, ok := xds.Clusters[route.HttpCluster]
	if !ok {
		return
	}
	delete(cluster.Routes, modelName)
	if len(cluster.Routes) == 0 {
		delete(xds.Clusters, route.HttpCluster)
	} else {
		xds.Clusters[route.HttpCluster] = cluster
	}
	delete(xds.Routes, modelName)
}

func (xds *SeldonXDSCache) AddEndpoint(clusterName, upstreamHost string, upstreamPort uint32) {
	cluster := xds.Clusters[clusterName]
	k := fmt.Sprintf("%s:%d",upstreamHost,upstreamPort)
	cluster.Endpoints[k] = resources.Endpoint{
		UpstreamHost: upstreamHost,
		UpstreamPort: upstreamPort,
	}

	xds.Clusters[clusterName] = cluster
}
