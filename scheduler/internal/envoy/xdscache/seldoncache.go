package xdscache


import (
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/seldonio/seldon-core/scheduler/internal/envoy/resources"
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
		r = append(r, resources.MakeCluster(c.Name))
	}

	return r
}

func (xds *SeldonXDSCache) RouteContents() []types.Resource {

	var routesArray []resources.Route
	for _, r := range xds.Routes {
		routesArray = append(routesArray, r)
	}

	return []types.Resource{resources.MakeRoute(routesArray)}
}

func (xds *SeldonXDSCache) ListenerContents() []types.Resource {
	var r []types.Resource

	for _, l := range xds.Listeners {
		r = append(r, resources.MakeHTTPListener(l.Name, l.RouteNames[0], l.Address, l.Port))
	}

	return r
}

func (xds *SeldonXDSCache) EndpointsContents() []types.Resource {
	var r []types.Resource

	for _, c := range xds.Clusters {
		r = append(r, resources.MakeEndpoint(c.Name, c.Endpoints))
	}

	return r
}

func (xds *SeldonXDSCache) AddListener(name string, modelNames []string) {
	xds.Listeners[name] = resources.Listener{
		Name:       name,
		Address:    DefaultListenerAddress,
		Port:       DefaultListenerPort,
		RouteNames: modelNames,
	}
}

func (xds *SeldonXDSCache) AddRoute(name, modelName string, server string) {
	xds.Routes[name] = resources.Route{
		Name:    name,
		Prefix:  modelName,
		Cluster: server,
	}
}

func (xds *SeldonXDSCache) AddCluster(name string) {
	xds.Clusters[name] = resources.Cluster{
		Name: name,
	}
}

func (xds *SeldonXDSCache) AddEndpoint(clusterName, upstreamHost string, upstreamPort uint32) {
	cluster := xds.Clusters[clusterName]

	cluster.Endpoints = append(cluster.Endpoints, resources.Endpoint{
		UpstreamHost: upstreamHost,
		UpstreamPort: upstreamPort,
	})

	xds.Clusters[clusterName] = cluster
}
