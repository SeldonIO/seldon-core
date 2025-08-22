// Copyright 2020 Envoyproxy Authors
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package xdscache

import (
	"fmt"
	"time"

	accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	matcher "github.com/envoyproxy/go-control-plane/envoy/config/common/matcher/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	tap "github.com/envoyproxy/go-control-plane/envoy/config/tap/v3"
	accesslog_file "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/file/v3"
	envoy_extensions_common_tap_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/common/tap/v3"
	luav3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/lua/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	tapfilter "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/tap/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	http "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	duration "google.golang.org/protobuf/types/known/durationpb"
	wrappers "google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

const (
	SeldonLoggingHeader           = "Seldon-Logging"
	EnvoyLogPathPrefix            = "/tmp/request-log"
	EnvoyAccessLogPath            = "/tmp/envoy-accesslog.txt"
	SeldonRouteSeparator          = ":" // Tried % but this seemed to break envoy matching. Maybe % is a special character or connected to regexp. A bug?
	DefaultRouteTimeoutSecs       = 0   // TODO allow configurable override
	DefaultRouteConfigurationName = "listener_0"
	MirrorRouteConfigurationName  = "listener_1"
	TLSRouteConfigurationName     = "listener_tls"
)

var (
	pipelineRoutePathHttp = &route.RouteMatch_Prefix{Prefix: "/v2"}
	pipelineRoutePathGrpc = &route.RouteMatch_Prefix{Prefix: "/inference.GRPCInferenceService"}
)

func makeHTTPListener(listenerName, address string,
	port uint32,
	routeConfigurationName string,
	serverSecret *Secret,
	config *EnvoyConfig,
) *listener.Listener {
	routerConfig, _ := anypb.New(&router.Router{})
	// HTTP filter configuration
	manager := &hcm.HttpConnectionManager{
		CodecType:                    hcm.HttpConnectionManager_AUTO,
		StatPrefix:                   listenerName,
		AlwaysSetRequestIdInResponse: false,
		GenerateRequestId:            &wrappers.BoolValue{Value: false},
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				ConfigSource:    makeConfigSource(),
				RouteConfigName: routeConfigurationName,
			},
		},
		HttpFilters: []*hcm.HttpFilter{
			{
				Name: "envoy.filters.http.tap",
				ConfigType: &hcm.HttpFilter_TypedConfig{
					TypedConfig: createTapConfig(),
				},
			},
			{
				Name: "envoy.filters.http.lua",
				ConfigType: &hcm.HttpFilter_TypedConfig{
					TypedConfig: createHeaderFilter(),
				},
			},
			{
				Name:       wellknown.Router,
				ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: routerConfig},
			},
		},
	}
	if config != nil && config.EnableAccessLog {
		var filter *accesslog.AccessLogFilter = nil
		if !config.IncludeSuccessfulRequests {
			filter = createAccessLogFilterMatchErrors()
		}
		manager.AccessLog = []*accesslog.AccessLog{
			{
				Name: "envoy.access_loggers.file",
				// log only errors if required
				Filter: filter,
				ConfigType: &accesslog.AccessLog_TypedConfig{
					TypedConfig: createAccessLogConfig(config.AccessLogPath),
				},
			},
		}
	}
	pbst, err := anypb.New(manager)
	if err != nil {
		panic(err)
	}

	return &listener.Listener{
		Name: listenerName,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  address,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		},
		FilterChains: []*listener.FilterChain{
			{
				Filters: []*listener.Filter{
					{
						Name: wellknown.HTTPConnectionManager,
						ConfigType: &listener.Filter_TypedConfig{
							TypedConfig: pbst,
						},
					},
				},
				TransportSocket: createDownstreamTransportSocket(serverSecret), // Add TLS if needed
			},
		},
	}
}

func makeCluster(clusterName string, eps map[string]Endpoint, isGrpc bool, clientSecret *Secret) *cluster.Cluster {
	if isGrpc {
		// Need to ensure http 2 is used
		// https://github.com/envoyproxy/go-control-plane/blob/d1a10d9a9366e8ab48f3f76b44a35930bac46fec/envoy/extensions/upstreams/http/v3/http_protocol_options.pb.go#L165-L166
		httpProtocolOptions := http.HttpProtocolOptions{
			UpstreamProtocolOptions: &http.HttpProtocolOptions_ExplicitHttpConfig_{
				ExplicitHttpConfig: &http.HttpProtocolOptions_ExplicitHttpConfig{
					ProtocolConfig: &http.HttpProtocolOptions_ExplicitHttpConfig_Http2ProtocolOptions{
						Http2ProtocolOptions: &core.Http2ProtocolOptions{},
					},
				},
			},
		}
		hpoMarshalled, err := anypb.New(&httpProtocolOptions)
		if err != nil {
			panic(err)
		}
		return &cluster.Cluster{
			Name:                          clusterName,
			ConnectTimeout:                duration.New(5 * time.Second),
			ClusterDiscoveryType:          &cluster.Cluster_Type{Type: cluster.Cluster_STRICT_DNS},
			LbPolicy:                      cluster.Cluster_LEAST_REQUEST,
			LoadAssignment:                makeEndpoint(clusterName, eps),
			DnsLookupFamily:               cluster.Cluster_V4_ONLY,
			TypedExtensionProtocolOptions: map[string]*anypb.Any{"envoy.extensions.upstreams.http.v3.HttpProtocolOptions": hpoMarshalled},
			TransportSocket:               createUpstreamTransportSocket(clientSecret),
			DnsRefreshRate:                duration.New(2 * time.Second),
			CircuitBreakers: &cluster.CircuitBreakers{
				Thresholds: []*cluster.CircuitBreakers_Thresholds{
					{
						MaxRetries: &wrappers.UInt32Value{Value: 5},
					},
				},
			},
		}
	} else {
		return &cluster.Cluster{
			Name:                 clusterName,
			ConnectTimeout:       duration.New(5 * time.Second),
			ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_STRICT_DNS},
			LbPolicy:             cluster.Cluster_LEAST_REQUEST,
			LoadAssignment:       makeEndpoint(clusterName, eps),
			DnsLookupFamily:      cluster.Cluster_V4_ONLY,
			TransportSocket:      createUpstreamTransportSocket(clientSecret),
			DnsRefreshRate:       duration.New(2 * time.Second),
			CircuitBreakers: &cluster.CircuitBreakers{
				Thresholds: []*cluster.CircuitBreakers_Thresholds{
					{
						MaxRetries: &wrappers.UInt32Value{Value: 5},
					},
				},
			},
		}
	}
}

func makeEndpoint(clusterName string, eps map[string]Endpoint) *endpoint.ClusterLoadAssignment {
	var endpoints []*endpoint.LbEndpoint

	for _, e := range eps {
		endpoints = append(endpoints, &endpoint.LbEndpoint{
			HostIdentifier: &endpoint.LbEndpoint_Endpoint{
				Endpoint: &endpoint.Endpoint{
					Address: &core.Address{
						Address: &core.Address_SocketAddress{
							SocketAddress: &core.SocketAddress{
								Protocol: core.SocketAddress_TCP,
								Address:  e.UpstreamHost,
								PortSpecifier: &core.SocketAddress_PortValue{
									PortValue: e.UpstreamPort,
								},
							},
						},
					},
				},
			},
		})
	}

	return &endpoint.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*endpoint.LocalityLbEndpoints{{
			LbEndpoints: endpoints,
		}},
	}
}

func makeRoutes(routes map[string]Route, pipelines map[string]PipelineRoute) (*route.RouteConfiguration, *route.RouteConfiguration) {
	rts := make([]*route.Route, 2*(len(routes)+len(pipelines))+
		countModelStickySessions(routes)+
		countPipelineStickySessions(pipelines))

	for i := range rts {
		rts[i] = &route.Route{
			Match: &route.RouteMatch{
				Headers: make([]*route.HeaderMatcher, 2), // We always do 2 header matches
			},
		}
	}

	rtsMirrors := make([]*route.Route, countModelMirrors(routes)+
		countPipelineMirrors(pipelines))

	for i := range rtsMirrors {
		rtsMirrors[i] = &route.Route{
			Match: &route.RouteMatch{
				Headers: make([]*route.HeaderMatcher, 2), // We always do 2 header matches
			},
		}
	}

	rtsIndex := 0
	mirrorsIndex := 0

	for _, r := range routes {
		for _, clusterTraffic := range r.Clusters { // it's an experiment, so create some sticky session routes
			if isExperiment(r.Clusters) {
				makeModelStickySessionEnvoyRoute(r.RouteName, rts[rtsIndex], r.LogPayloads, &clusterTraffic, false)
				rtsIndex++
				makeModelStickySessionEnvoyRoute(r.RouteName, rts[rtsIndex], r.LogPayloads, &clusterTraffic, true)
				rtsIndex++
			}
		}
		makeModelEnvoyRoute(&r, rts[rtsIndex], false, false)
		rtsIndex++
		makeModelEnvoyRoute(&r, rts[rtsIndex], true, false)
		rtsIndex++

		if r.Mirror != nil {
			makeModelEnvoyRoute(&r, rtsMirrors[mirrorsIndex], false, true)
			mirrorsIndex++
			makeModelEnvoyRoute(&r, rtsMirrors[mirrorsIndex], true, true)
			mirrorsIndex++
		}
	}

	// Create Pipeline Routes
	for _, r := range pipelines {
		if isExperiment(r.Clusters) { // it's an experiment, so create some sticky session routes
			for _, clusterTraffic := range r.Clusters {
				makePipelineStickySessionEnvoyRoute(r.RouteName, rts[rtsIndex], &clusterTraffic, false)
				rtsIndex++
				makePipelineStickySessionEnvoyRoute(r.RouteName, rts[rtsIndex], &clusterTraffic, true)
				rtsIndex++
			}
		}

		makePipelineEnvoyRoute(&r, rts[rtsIndex], false, false)
		rtsIndex++
		makePipelineEnvoyRoute(&r, rts[rtsIndex], true, false)
		rtsIndex++

		if r.Mirror != nil {
			makePipelineEnvoyRoute(&r, rtsMirrors[mirrorsIndex], false, true)
			mirrorsIndex++
			makePipelineEnvoyRoute(&r, rtsMirrors[mirrorsIndex], true, true)
			mirrorsIndex++
		}
	}

	return &route.RouteConfiguration{
			Name: DefaultRouteConfigurationName,
			VirtualHosts: []*route.VirtualHost{{
				Name:    "seldon_service",
				Domains: []string{"*"},
				Routes:  rts,
			}},
		},
		&route.RouteConfiguration{
			Name: MirrorRouteConfigurationName,
			VirtualHosts: []*route.VirtualHost{{
				Name:    "seldon_mirror",
				Domains: []string{"*"},
				Routes:  rtsMirrors,
			}},
		}
}

func wrapRouteHeader(key string) string {
	return fmt.Sprintf("%s%s%s", SeldonRouteSeparator, key, SeldonRouteSeparator)
}

func createMirrorRouteAction(trafficWeight uint32, isGrpc bool) []*route.RouteAction_RequestMirrorPolicy {
	var mirrors []*route.RouteAction_RequestMirrorPolicy
	clusterName := MirrorHttpClusterName
	if isGrpc {
		clusterName = MirrorGrpcClusterName
	}
	mirrors = append(mirrors, &route.RouteAction_RequestMirrorPolicy{
		Cluster: clusterName,
		RuntimeFraction: &core.RuntimeFractionalPercent{
			DefaultValue: &typev3.FractionalPercent{
				Numerator:   trafficWeight, // Just take the first one - at present all will be same
				Denominator: typev3.FractionalPercent_HUNDRED,
			},
		},
	})
	return mirrors
}

// weighted clusters do not play well with session affinity see https://github.com/envoyproxy/envoy/issues/8167
// Traffic shifting may need to be reinvesigated https://github.com/envoyproxy/envoy/pull/18207
func createWeightedModelClusterAction(clusterTraffics []TrafficSplit, mirrorTraffic *TrafficSplit, isGrpc bool) *route.Route_Route {
	// Add Weighted Clusters with given traffic percentages to each internal model
	var splits []*route.WeightedCluster_ClusterWeight
	var mirrors []*route.RouteAction_RequestMirrorPolicy
	var totWeight uint32
	for _, clusterTraffic := range clusterTraffics {
		clusterName := clusterTraffic.HttpCluster
		if isGrpc {
			clusterName = clusterTraffic.GrpcCluster
		}
		totWeight = totWeight + clusterTraffic.TrafficWeight
		splits = append(splits,
			&route.WeightedCluster_ClusterWeight{
				Name: clusterName,
				Weight: &wrappers.UInt32Value{
					Value: clusterTraffic.TrafficWeight,
				},
				RequestHeadersToRemove: []string{util.SeldonInternalModelHeader},
				RequestHeadersToAdd: []*core.HeaderValueOption{
					{
						Header: &core.HeaderValue{
							Key: util.SeldonInternalModelHeader,
							// note: this is implementation specific for agent and it is exposed here
							// basically the model versions are flattened and it is loaded as
							// <model_name>_<model_version>
							// TODO: is there a nicer way of doing it?
							// check client.go for how different model versions are treated internally
							Value: util.GetVersionedModelName(
								clusterTraffic.ModelName, clusterTraffic.ModelVersion),
						},
					},
				},
				ResponseHeadersToAdd: []*core.HeaderValueOption{
					{
						Header: &core.HeaderValue{
							Key: util.SeldonRouteHeader,
							Value: wrapRouteHeader(util.GetVersionedModelName(
								clusterTraffic.ModelName, clusterTraffic.ModelVersion)),
						},
					},
				},
			})

	}
	if mirrorTraffic != nil {
		mirrors = createMirrorRouteAction(mirrorTraffic.TrafficWeight, isGrpc)
	}

	action := &route.Route_Route{
		Route: &route.RouteAction{
			Timeout: &duration.Duration{Seconds: DefaultRouteTimeoutSecs},
			ClusterSpecifier: &route.RouteAction_WeightedClusters{
				WeightedClusters: &route.WeightedCluster{
					Clusters: splits,
				},
			},
			RequestMirrorPolicies: mirrors,
			RetryPolicy: &route.RetryPolicy{
				RetryOn:    "5xx,connect-failure",
				NumRetries: &wrappers.UInt32Value{Value: 5},
				RetryBackOff: &route.RetryPolicy_RetryBackOff{
					BaseInterval: &duration.Duration{
						Nanos: int32(time.Millisecond * 500),
					},
				},
			},
		},
	}
	return action
}

var (
	modelRouteMatchPathHttp = &route.RouteMatch_Prefix{Prefix: "/v2"}
	modelRouteMatchPathGrpc = &route.RouteMatch_Prefix{Prefix: "/inference.GRPCInferenceService"}
	modelRouteHeaders       = []*core.HeaderValueOption{
		{Header: &core.HeaderValue{Key: SeldonLoggingHeader, Value: "true"}},
	}
)

func getPipelineClusterName(clusterPrefix string, isGrpc bool) string {
	if isGrpc {
		return fmt.Sprintf("%s_%s_grpc", clusterPrefix, util.SeldonPipelineHeaderSuffix)
	}
	return fmt.Sprintf("%s_%s_http", clusterPrefix, util.SeldonPipelineHeaderSuffix)
}

func getRouteName(routeName string, isGrpc bool, isMirror bool) string {
	mirrorSuffix := ""
	if isMirror {
		mirrorSuffix = "_mirror"
	}
	httpSuffix := "_http"
	if isGrpc {
		httpSuffix = "_grpc"
	}
	return fmt.Sprintf("%s%s%s", routeName, httpSuffix, mirrorSuffix)
}

func makeModelStickySessionEnvoyRoute(routeName string, envoyRoute *route.Route, logPayloads bool, clusterTraffic *TrafficSplit, isGrpc bool) {
	if isGrpc {
		envoyRoute.Name = routeName + "_grpc_experiment"
		envoyRoute.Match.PathSpecifier = modelRouteMatchPathGrpc
	} else {
		envoyRoute.Name = routeName + "_http_experiment"
		envoyRoute.Match.PathSpecifier = modelRouteMatchPathHttp
	}

	envoyRoute.Match.Headers[0] = &route.HeaderMatcher{
		Name: util.SeldonModelHeader, // Header name we will match on
		HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
			StringMatch: &matcherv3.StringMatcher{
				MatchPattern: &matcherv3.StringMatcher_Exact{
					Exact: routeName,
				},
			},
		},
	}
	envoyRoute.Match.Headers[1] = &route.HeaderMatcher{
		Name: util.SeldonRouteHeader, // Header name we will match on
		HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
			StringMatch: &matcherv3.StringMatcher{
				MatchPattern: &matcherv3.StringMatcher_Contains{
					Contains: wrapRouteHeader(util.GetVersionedModelName(
						clusterTraffic.ModelName, clusterTraffic.ModelVersion)),
				},
			},
		},
	}

	envoyRoute.RequestHeadersToAdd = []*core.HeaderValueOption{
		{
			Header: &core.HeaderValue{
				Key: util.SeldonInternalModelHeader,
				Value: util.GetVersionedModelName(
					clusterTraffic.ModelName, clusterTraffic.ModelVersion),
			},
		},
		{
			Header: &core.HeaderValue{
				Key:   util.SeldonModelHeader,
				Value: clusterTraffic.ModelName,
			},
		},
	}
	envoyRoute.ResponseHeadersToAdd = []*core.HeaderValueOption{
		{
			Header: &core.HeaderValue{
				Key: util.SeldonRouteHeader,
				Value: wrapRouteHeader(util.GetVersionedModelName(
					clusterTraffic.ModelName, clusterTraffic.ModelVersion)),
			},
		},
	}
	if isGrpc {
		envoyRoute.Action = &route.Route_Route{
			Route: &route.RouteAction{
				Timeout: &duration.Duration{Seconds: DefaultRouteTimeoutSecs},
				ClusterSpecifier: &route.RouteAction_Cluster{
					Cluster: clusterTraffic.GrpcCluster,
				},
			},
		}
	} else {
		envoyRoute.Action = &route.Route_Route{
			Route: &route.RouteAction{
				Timeout: &duration.Duration{Seconds: DefaultRouteTimeoutSecs},
				ClusterSpecifier: &route.RouteAction_Cluster{
					Cluster: clusterTraffic.HttpCluster,
				},
			},
		}
	}
	if logPayloads {
		envoyRoute.ResponseHeadersToAdd = append(envoyRoute.RequestHeadersToAdd, modelRouteHeaders...)
	}
}

func makeModelEnvoyRoute(r *Route, envoyRoute *route.Route, isGrpc, isMirror bool) {
	envoyRoute.Name = getRouteName(r.RouteName, isGrpc, isMirror)
	if isGrpc {
		envoyRoute.Match.PathSpecifier = modelRouteMatchPathGrpc
	} else {
		envoyRoute.Match.PathSpecifier = modelRouteMatchPathHttp
	}
	envoyRoute.Match.Headers[0] = &route.HeaderMatcher{
		Name: util.SeldonModelHeader, // Header name we will match on
		HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
			StringMatch: &matcherv3.StringMatcher{
				MatchPattern: &matcherv3.StringMatcher_Exact{
					Exact: r.RouteName,
				},
			},
		},
	}
	envoyRoute.Match.Headers[1] = &route.HeaderMatcher{
		Name: util.SeldonRouteHeader,
		HeaderMatchSpecifier: &route.HeaderMatcher_PresentMatch{
			PresentMatch: false,
		},
	}

	if isMirror {
		envoyRoute.Action = createWeightedModelClusterAction([]TrafficSplit{*r.Mirror}, nil, isGrpc)
	} else {
		envoyRoute.Action = createWeightedModelClusterAction(r.Clusters, r.Mirror, isGrpc)
	}

	if r.LogPayloads {
		envoyRoute.ResponseHeadersToAdd = modelRouteHeaders
	}
}

func makePipelineEnvoyRoute(r *PipelineRoute, envoyRoute *route.Route, isGrpc, isMirror bool) {
	envoyRoute.Name = getRouteName(r.RouteName, isGrpc, isMirror)
	envoyRoute.Match.PathSpecifier = pipelineRoutePathHttp
	if isGrpc {
		envoyRoute.Match.PathSpecifier = pipelineRoutePathGrpc
	}
	envoyRoute.Match.Headers[0] = &route.HeaderMatcher{
		Name: util.SeldonModelHeader, // Header name we will match on
		HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
			StringMatch: &matcherv3.StringMatcher{
				MatchPattern: &matcherv3.StringMatcher_Exact{
					Exact: r.RouteName,
				},
			},
		},
	}
	envoyRoute.Match.Headers[1] = &route.HeaderMatcher{
		Name: util.SeldonRouteHeader,
		HeaderMatchSpecifier: &route.HeaderMatcher_PresentMatch{
			PresentMatch: false,
		},
	}

	if isMirror {
		envoyRoute.Action = createWeightedPipelineClusterAction([]PipelineTrafficSplit{*r.Mirror}, nil, isGrpc)
	} else {
		envoyRoute.Action = createWeightedPipelineClusterAction(r.Clusters, r.Mirror, isGrpc)
	}
}

func getPipelineModelName(pipelineName string) string {
	return fmt.Sprintf("%s.%s", pipelineName, util.SeldonPipelineHeaderSuffix)
}

func createWeightedPipelineClusterAction(clusterTraffics []PipelineTrafficSplit, mirrorTraffic *PipelineTrafficSplit, isGrpc bool) *route.Route_Route {
	// Add Weighted Clusters with given traffic percentages to each internal model
	var splits []*route.WeightedCluster_ClusterWeight
	var mirrors []*route.RouteAction_RequestMirrorPolicy
	var totWeight uint32
	for _, clusterTraffic := range clusterTraffics {
		clusterName := getPipelineClusterName(clusterTraffic.PipelineName, isGrpc)
		totWeight = totWeight + clusterTraffic.TrafficWeight
		splits = append(splits,
			&route.WeightedCluster_ClusterWeight{
				Name: clusterName,
				Weight: &wrappers.UInt32Value{
					Value: clusterTraffic.TrafficWeight,
				},
				RequestHeadersToAdd: []*core.HeaderValueOption{
					{
						Header: &core.HeaderValue{
							Key:   util.SeldonInternalModelHeader,
							Value: getPipelineModelName(clusterTraffic.PipelineName),
						},
					},
				},
				ResponseHeadersToAdd: []*core.HeaderValueOption{
					{
						Header: &core.HeaderValue{
							Key:   util.SeldonRouteHeader,
							Value: wrapRouteHeader(getPipelineModelName(clusterTraffic.PipelineName)),
						},
					},
				},
			})

	}
	if mirrorTraffic != nil {
		mirrors = createMirrorRouteAction(mirrorTraffic.TrafficWeight, isGrpc)
	}
	action := &route.Route_Route{
		Route: &route.RouteAction{
			Timeout: &duration.Duration{Seconds: DefaultRouteTimeoutSecs},
			ClusterSpecifier: &route.RouteAction_WeightedClusters{
				WeightedClusters: &route.WeightedCluster{
					Clusters: splits,
				},
			},
			RequestMirrorPolicies: mirrors,
		},
	}
	return action
}

func makePipelineStickySessionEnvoyRoute(routeName string, envoyRoute *route.Route, clusterTraffic *PipelineTrafficSplit, isGrpc bool) {
	if isGrpc {
		envoyRoute.Name = routeName + "_grpc_experiment"
		envoyRoute.Match.PathSpecifier = pipelineRoutePathGrpc
	} else {
		envoyRoute.Name = routeName + "_http_experiment"
		envoyRoute.Match.PathSpecifier = pipelineRoutePathHttp
	}

	envoyRoute.Match.Headers[0] = &route.HeaderMatcher{
		Name: util.SeldonRouteHeader, // Header name we will match on
		HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
			StringMatch: &matcherv3.StringMatcher{
				MatchPattern: &matcherv3.StringMatcher_Contains{
					Contains: wrapRouteHeader(getPipelineModelName(clusterTraffic.PipelineName)),
				},
			},
		},
	}
	envoyRoute.Match.Headers[1] = &route.HeaderMatcher{
		Name: util.SeldonModelHeader, // Header name we will match on
		HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
			StringMatch: &matcherv3.StringMatcher{
				MatchPattern: &matcherv3.StringMatcher_Exact{
					Exact: routeName,
				},
			},
		},
	}
	envoyRoute.RequestHeadersToAdd = []*core.HeaderValueOption{
		{
			Header: &core.HeaderValue{
				Key:   util.SeldonInternalModelHeader,
				Value: getPipelineModelName(clusterTraffic.PipelineName),
			},
		},
	}
	envoyRoute.ResponseHeadersToAdd = []*core.HeaderValueOption{
		{
			Header: &core.HeaderValue{
				Key:   util.SeldonRouteHeader,
				Value: wrapRouteHeader(getPipelineModelName(clusterTraffic.PipelineName)),
			},
		},
	}
	envoyRoute.Action = &route.Route_Route{
		Route: &route.RouteAction{
			Timeout: &duration.Duration{Seconds: DefaultRouteTimeoutSecs},
			ClusterSpecifier: &route.RouteAction_Cluster{
				Cluster: getPipelineClusterName(routeName, isGrpc),
			},
		},
	}
}

func countModelStickySessions(routes map[string]Route) int {
	count := 0
	for _, r := range routes {
		if isExperiment(r.Clusters) {
			count = count + (len(r.Clusters) * 2) // REST and GRPC routes for each model in an experiment
		}
	}
	return count
}

func countPipelineStickySessions(pipelineRoutes map[string]PipelineRoute) int {
	count := 0
	for _, r := range pipelineRoutes {
		if isExperiment(r.Clusters) {
			count = count + (len(r.Clusters) * 2) // REST and GRPC routes for each model in an experiment
		}
	}
	return count
}

func isExperiment[T any](clusters []T) bool {
	return len(clusters) > 1
}

func countModelMirrors(models map[string]Route) int {
	count := 0
	for _, r := range models {
		if r.Mirror != nil {
			count = count + 2 // REST and gRPC
		}
	}
	return count
}

func countPipelineMirrors(pipelines map[string]PipelineRoute) int {
	count := 0
	for _, r := range pipelines {
		if r.Mirror != nil {
			count = count + 2 // REST and gRPC
		}
	}
	return count
}

func createTapConfig() *anypb.Any {
	// Create Tap Config
	tapFilter := tapfilter.Tap{
		CommonConfig: &envoy_extensions_common_tap_v3.CommonExtensionConfig{
			ConfigType: &envoy_extensions_common_tap_v3.CommonExtensionConfig_StaticConfig{
				StaticConfig: &tap.TapConfig{
					Match: &matcher.MatchPredicate{
						Rule: &matcher.MatchPredicate_OrMatch{ // Either match request or response header
							OrMatch: &matcher.MatchPredicate_MatchSet{
								Rules: []*matcher.MatchPredicate{
									{
										Rule: &matcher.MatchPredicate_HttpResponseHeadersMatch{ // Response header
											HttpResponseHeadersMatch: &matcher.HttpHeadersMatch{
												Headers: []*route.HeaderMatcher{
													{
														Name:                 SeldonLoggingHeader,
														HeaderMatchSpecifier: &route.HeaderMatcher_PresentMatch{PresentMatch: true},
													},
												},
											},
										},
									},
									{
										Rule: &matcher.MatchPredicate_HttpRequestHeadersMatch{ // Request header
											HttpRequestHeadersMatch: &matcher.HttpHeadersMatch{
												Headers: []*route.HeaderMatcher{
													{
														Name:                 SeldonLoggingHeader,
														HeaderMatchSpecifier: &route.HeaderMatcher_PresentMatch{PresentMatch: true},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					OutputConfig: &tap.OutputConfig{
						Sinks: []*tap.OutputSink{
							{
								OutputSinkType: &tap.OutputSink_FilePerTap{
									FilePerTap: &tap.FilePerTapSink{
										PathPrefix: EnvoyLogPathPrefix,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	tapAny, err := anypb.New(&tapFilter)
	if err != nil {
		panic(err)
	}
	return tapAny
}

func createAccessLogConfig(path string) *anypb.Any {
	config := accesslog_file.FileAccessLog{
		Path: path,
		AccessLogFormat: &accesslog_file.FileAccessLog_LogFormat{
			LogFormat: &core.SubstitutionFormatString{
				Format: &core.SubstitutionFormatString_TextFormatSource{
					TextFormatSource: &core.DataSource{
						Specifier: &core.DataSource_InlineString{
							InlineString: "[%START_TIME%] \"%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%\" %RESPONSE_CODE% %GRPC_STATUS_NUMBER% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% \"%REQ(X-FORWARDED-FOR)%\" \"%REQ(USER-AGENT)%\" \"%REQ(X-REQUEST-ID)%\" \"%REQ(:AUTHORITY)%\" \"%UPSTREAM_HOST%\"\n",
						},
					},
				},
			},
		},
	}
	accessAny, err := anypb.New(&config)
	if err != nil {
		panic(err)
	}
	return accessAny
}

func createAccessLogFilterMatchErrors() *accesslog.AccessLogFilter {
	grpcMatcher := &route.HeaderMatcher_StringMatch{
		StringMatch: &matcherv3.StringMatcher{
			MatchPattern: &matcherv3.StringMatcher_Prefix{
				// for grpc content-type starts with application/grpc:
				// Content-Type → “content-type” “application/grpc” [(“+proto” / “+json” / {custom})]
				Prefix: "application/grpc",
			},
			IgnoreCase: true,
		},
	}
	return &accesslog.AccessLogFilter{
		FilterSpecifier: &accesslog.AccessLogFilter_OrFilter{
			OrFilter: &accesslog.OrFilter{
				Filters: []*accesslog.AccessLogFilter{
					// http
					{
						FilterSpecifier: &accesslog.AccessLogFilter_AndFilter{
							AndFilter: &accesslog.AndFilter{
								Filters: []*accesslog.AccessLogFilter{
									{
										FilterSpecifier: &accesslog.AccessLogFilter_StatusCodeFilter{
											StatusCodeFilter: &accesslog.StatusCodeFilter{
												Comparison: &accesslog.ComparisonFilter{
													Op:    accesslog.ComparisonFilter_GE,
													Value: &core.RuntimeUInt32{DefaultValue: 400, RuntimeKey: "status_code"},
												},
											},
										},
									},
									{
										FilterSpecifier: &accesslog.AccessLogFilter_HeaderFilter{
											HeaderFilter: &accesslog.HeaderFilter{
												Header: &route.HeaderMatcher{
													Name:                 "content-type",
													HeaderMatchSpecifier: grpcMatcher,
													InvertMatch:          true,
												},
											},
										},
									},
								},
							},
						},
					},
					// grpc
					{
						FilterSpecifier: &accesslog.AccessLogFilter_AndFilter{
							AndFilter: &accesslog.AndFilter{
								Filters: []*accesslog.AccessLogFilter{
									{
										FilterSpecifier: &accesslog.AccessLogFilter_GrpcStatusFilter{
											GrpcStatusFilter: &accesslog.GrpcStatusFilter{
												Statuses: []accesslog.GrpcStatusFilter_Status{accesslog.GrpcStatusFilter_OK},
												Exclude:  true,
											},
										},
									},
									{
										FilterSpecifier: &accesslog.AccessLogFilter_HeaderFilter{
											HeaderFilter: &accesslog.HeaderFilter{
												Header: &route.HeaderMatcher{
													Name:                 "content-type",
													HeaderMatchSpecifier: grpcMatcher,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// A filter to add the seldon-model header from the http path if its not passed
// Does this for model and pipeline paths allowing us to keep our current header based routing
func createHeaderFilter() *anypb.Any {
	luaFilter := luav3.Lua{
		DefaultSourceCode: &core.DataSource{
			Specifier: &core.DataSource_InlineString{
				InlineString: `function envoy_on_request(request_handle)
  local modelHeader = request_handle:headers():get("` + util.SeldonModelHeader + `")
  local routeHeader = request_handle:headers():get("` + util.SeldonRouteHeader + `")
  if (modelHeader == nil or modelHeader == '') and (routeHeader == nil or routeHeader == '') then
    local path = request_handle:headers():get(":path")
    local i, j = string.find(path,"/v2/models/")
    if i == 1 then
      local s = string.sub(path,j+1)
      i, j = string.find(s, "/")
      if i then
        local model = string.sub(s,0,i-1)
        request_handle:headers():add("` + util.SeldonModelHeader + `",model)
      else
        request_handle:headers():add("` + util.SeldonModelHeader + `",s)
      end
    else
      i, j = string.find(path,"/v2/pipelines/")
      if i == 1 then
        local s = string.sub(path,j+1)
        i, j = string.find(s, "/")
        local model = string.sub(s,0,i-1)
        request_handle:headers():add("` + util.SeldonModelHeader + `",model..".` + util.SeldonPipelineHeaderSuffix + `")
      end
    end
  end
end
`,
			},
		},
	}
	luaAny, err := anypb.New(&luaFilter)
	if err != nil {
		panic(err)
	}
	return luaAny
}

func makeConfigSource() *core.ConfigSource {
	source := &core.ConfigSource{}
	source.ResourceApiVersion = resource.DefaultAPIVersion
	source.ConfigSourceSpecifier = &core.ConfigSource_Ads{}
	return source
}
