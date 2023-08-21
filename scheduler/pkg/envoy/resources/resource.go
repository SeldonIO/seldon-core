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

package resources

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
	"github.com/golang/protobuf/ptypes/duration"
	wrappers "github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

const (
	SeldonLoggingHeader           = "Seldon-Logging"
	EnvoyLogPathPrefix            = "/tmp/request-log"
	SeldonModelHeader             = "seldon-model"
	SeldonPipelineHeader          = "pipeline"
	SeldonInternalModelHeader     = "seldon-internal-model"
	SeldonRouteHeader             = "x-seldon-route"
	SeldonRouteSeparator          = ":" // Tried % but this seemed to break envoy matching. Maybe % is a special character or connected to regexp. A bug?
	SeldonModelHeaderSuffix       = "model"
	SeldonPipelineHeaderSuffix    = "pipeline"
	DefaultRouteTimeoutSecs       = 0 //TODO allow configurable override
	ExternalHeaderPrefix          = "x-"
	DefaultRouteConfigurationName = "listener_0"
	MirrorRouteConfigurationName  = "listener_1"
	TLSRouteConfigurationName     = "listener_tls"
)

func MakeCluster(clusterName string, eps []Endpoint, isGrpc bool, clientSecret *Secret) *cluster.Cluster {
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
			ConnectTimeout:                durationpb.New(5 * time.Second),
			ClusterDiscoveryType:          &cluster.Cluster_Type{Type: cluster.Cluster_STRICT_DNS},
			LbPolicy:                      cluster.Cluster_LEAST_REQUEST,
			LoadAssignment:                MakeEndpoint(clusterName, eps),
			DnsLookupFamily:               cluster.Cluster_V4_ONLY,
			TypedExtensionProtocolOptions: map[string]*anypb.Any{"envoy.extensions.upstreams.http.v3.HttpProtocolOptions": hpoMarshalled},
			TransportSocket:               createUpstreamTransportSocket(clientSecret),
		}
	} else {
		return &cluster.Cluster{
			Name:                 clusterName,
			ConnectTimeout:       durationpb.New(5 * time.Second),
			ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_STRICT_DNS},
			LbPolicy:             cluster.Cluster_LEAST_REQUEST,
			LoadAssignment:       MakeEndpoint(clusterName, eps),
			DnsLookupFamily:      cluster.Cluster_V4_ONLY,
			TransportSocket:      createUpstreamTransportSocket(clientSecret),
		}
	}
}

func MakeEndpoint(clusterName string, eps []Endpoint) *endpoint.ClusterLoadAssignment {
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

func wrapRouteHeader(key string) string {
	return fmt.Sprintf("%s%s%s", SeldonRouteSeparator, key, SeldonRouteSeparator)
}

func createMirrorRouteAction(trafficWeight uint32, rest bool) []*route.RouteAction_RequestMirrorPolicy {
	var mirrors []*route.RouteAction_RequestMirrorPolicy
	clusterName := MirrorHttpClusterName
	if !rest {
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
func createWeightedModelClusterAction(clusterTraffics []TrafficSplits, mirrorTraffics []TrafficSplits, rest bool) *route.Route_Route {
	// Add Weighted Clusters with given traffic percentages to each internal model
	var splits []*route.WeightedCluster_ClusterWeight
	var mirrors []*route.RouteAction_RequestMirrorPolicy
	var totWeight uint32
	for _, clusterTraffic := range clusterTraffics {
		clusterName := clusterTraffic.HttpCluster
		if !rest {
			clusterName = clusterTraffic.GrpcCluster
		}
		totWeight = totWeight + clusterTraffic.TrafficWeight
		splits = append(splits,
			&route.WeightedCluster_ClusterWeight{
				Name: clusterName,
				Weight: &wrappers.UInt32Value{
					Value: clusterTraffic.TrafficWeight,
				},
				RequestHeadersToRemove: []string{SeldonInternalModelHeader},
				RequestHeadersToAdd: []*core.HeaderValueOption{
					{
						Header: &core.HeaderValue{
							Key: SeldonInternalModelHeader,
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
							Key: SeldonRouteHeader,
							Value: wrapRouteHeader(util.GetVersionedModelName(
								clusterTraffic.ModelName, clusterTraffic.ModelVersion)),
						},
					},
				},
			})

	}
	if len(mirrorTraffics) > 0 {
		mirrors = createMirrorRouteAction(mirrorTraffics[0].TrafficWeight, rest)
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

var modelRouteMatchPathHttp = &route.RouteMatch_Prefix{Prefix: "/v2"}
var modelRouteMatchPathGrpc = &route.RouteMatch_Prefix{Prefix: "/inference.GRPCInferenceService"}
var modelRouteHeaders = []*core.HeaderValueOption{
	{Header: &core.HeaderValue{Key: SeldonLoggingHeader, Value: "true"}},
}

func getRouteName(routeName string, isPipeline bool, isGrpc bool, isMirror bool) string {
	pipelineSuffix := ""
	if isPipeline {
		pipelineSuffix = "_pipeline"
	}
	mirrorSuffix := ""
	if isMirror {
		mirrorSuffix = "_mirror"
	}
	httpSuffix := "_http"
	if isGrpc {
		httpSuffix = "_grpc"
	}
	return fmt.Sprintf("%s%s%s%s", routeName, pipelineSuffix, httpSuffix, mirrorSuffix)
}

func makeModelHttpRoute(r *Route, rt *route.Route, isMirror bool) {
	rt.Name = getRouteName(r.RouteName, false, false, isMirror)
	rt.Match.PathSpecifier = modelRouteMatchPathHttp
	rt.Match.Headers[0] = &route.HeaderMatcher{
		Name: SeldonModelHeader, // Header name we will match on
		HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
			StringMatch: &matcherv3.StringMatcher{
				MatchPattern: &matcherv3.StringMatcher_Exact{
					Exact: r.RouteName,
				},
			},
		},
	}
	rt.Match.Headers[1] = &route.HeaderMatcher{
		Name: SeldonRouteHeader,
		HeaderMatchSpecifier: &route.HeaderMatcher_PresentMatch{
			PresentMatch: false,
		},
	}

	if isMirror {
		rt.Action = createWeightedModelClusterAction(r.Mirrors, []TrafficSplits{}, true)
	} else {
		rt.Action = createWeightedModelClusterAction(r.Clusters, r.Mirrors, true)
	}

	if r.LogPayloads {
		rt.ResponseHeadersToAdd = modelRouteHeaders
	}
}

func makeModelStickySessionRoute(r *Route, clusterTraffic *TrafficSplits, rt *route.Route, isGrpc bool) {
	if isGrpc {
		rt.Name = r.RouteName + "_grpc_experiment"
		rt.Match.PathSpecifier = modelRouteMatchPathGrpc
	} else {
		rt.Name = r.RouteName + "_http_experiment"
		rt.Match.PathSpecifier = modelRouteMatchPathHttp
	}

	rt.Match.Headers[0] = &route.HeaderMatcher{
		Name: SeldonModelHeader, // Header name we will match on
		HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
			StringMatch: &matcherv3.StringMatcher{
				MatchPattern: &matcherv3.StringMatcher_Exact{
					Exact: r.RouteName,
				},
			},
		},
	}
	rt.Match.Headers[1] = &route.HeaderMatcher{
		Name: SeldonRouteHeader, // Header name we will match on
		HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
			StringMatch: &matcherv3.StringMatcher{
				MatchPattern: &matcherv3.StringMatcher_Contains{
					Contains: wrapRouteHeader(util.GetVersionedModelName(
						clusterTraffic.ModelName, clusterTraffic.ModelVersion)),
				},
			},
		},
	}

	rt.RequestHeadersToAdd = []*core.HeaderValueOption{
		{
			Header: &core.HeaderValue{
				Key: SeldonInternalModelHeader,
				Value: util.GetVersionedModelName(
					clusterTraffic.ModelName, clusterTraffic.ModelVersion),
			},
		},
		{
			Header: &core.HeaderValue{
				Key:   SeldonModelHeader,
				Value: clusterTraffic.ModelName,
			},
		},
	}
	rt.ResponseHeadersToAdd = []*core.HeaderValueOption{
		{
			Header: &core.HeaderValue{
				Key: SeldonRouteHeader,
				Value: wrapRouteHeader(util.GetVersionedModelName(
					clusterTraffic.ModelName, clusterTraffic.ModelVersion)),
			},
		},
	}
	if isGrpc {
		rt.Action = &route.Route_Route{
			Route: &route.RouteAction{
				Timeout: &duration.Duration{Seconds: DefaultRouteTimeoutSecs},
				ClusterSpecifier: &route.RouteAction_Cluster{
					Cluster: clusterTraffic.GrpcCluster,
				},
			},
		}
	} else {
		rt.Action = &route.Route_Route{
			Route: &route.RouteAction{
				Timeout: &duration.Duration{Seconds: DefaultRouteTimeoutSecs},
				ClusterSpecifier: &route.RouteAction_Cluster{
					Cluster: clusterTraffic.HttpCluster,
				},
			},
		}
	}
	if r.LogPayloads {
		rt.ResponseHeadersToAdd = append(rt.RequestHeadersToAdd, modelRouteHeaders...)
	}
}

func makeModelGrpcRoute(r *Route, rt *route.Route, isMirror bool) {
	rt.Name = getRouteName(r.RouteName, false, true, isMirror)
	rt.Match.PathSpecifier = modelRouteMatchPathGrpc
	rt.Match.Headers[0] = &route.HeaderMatcher{
		Name: SeldonModelHeader, // Header name we will match on
		HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
			StringMatch: &matcherv3.StringMatcher{
				MatchPattern: &matcherv3.StringMatcher_Exact{
					Exact: r.RouteName,
				},
			},
		},
	}
	rt.Match.Headers[1] = &route.HeaderMatcher{
		Name: SeldonRouteHeader,
		HeaderMatchSpecifier: &route.HeaderMatcher_PresentMatch{
			PresentMatch: false,
		},
	}

	if isMirror {
		rt.Action = createWeightedModelClusterAction(r.Mirrors, []TrafficSplits{}, false)
	} else {
		rt.Action = createWeightedModelClusterAction(r.Clusters, r.Mirrors, false)
	}

	if r.LogPayloads {
		rt.ResponseHeadersToAdd = append(rt.RequestHeadersToAdd, modelRouteHeaders...)
	}
}

var pipelineRoutePathHttp = &route.RouteMatch_Prefix{Prefix: "/v2"}
var pipelineRoutePathGrpc = &route.RouteMatch_Prefix{Prefix: "/inference.GRPCInferenceService"}

func getPipelineModelName(pipelineName string) string {
	return fmt.Sprintf("%s.%s", pipelineName, SeldonPipelineHeaderSuffix)
}

func createWeightedPipelineClusterAction(clusterTraffics []PipelineTrafficSplits, mirrorTraffics []PipelineTrafficSplits, rest bool) *route.Route_Route {
	// Add Weighted Clusters with given traffic percentages to each internal model
	var splits []*route.WeightedCluster_ClusterWeight
	var mirrors []*route.RouteAction_RequestMirrorPolicy
	var totWeight uint32
	for _, clusterTraffic := range clusterTraffics {
		clusterName := PipelineGatewayHttpClusterName
		if !rest {
			clusterName = PipelineGatewayGrpcClusterName
		}
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
							Key:   SeldonInternalModelHeader,
							Value: getPipelineModelName(clusterTraffic.PipelineName),
						},
					},
				},
				ResponseHeadersToAdd: []*core.HeaderValueOption{
					{
						Header: &core.HeaderValue{
							Key:   SeldonRouteHeader,
							Value: wrapRouteHeader(getPipelineModelName(clusterTraffic.PipelineName)),
						},
					},
				},
			})

	}
	if len(mirrorTraffics) > 0 {
		mirrors = createMirrorRouteAction(mirrorTraffics[0].TrafficWeight, rest)
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

func makePipelineHttpRoute(r *PipelineRoute, rt *route.Route, isMirror bool) {
	rt.Name = getRouteName(r.RouteName, true, false, isMirror)
	rt.Match.PathSpecifier = pipelineRoutePathHttp
	rt.Match.Headers[0] = &route.HeaderMatcher{
		Name: SeldonModelHeader, // Header name we will match on
		HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
			StringMatch: &matcherv3.StringMatcher{
				MatchPattern: &matcherv3.StringMatcher_Exact{
					Exact: r.RouteName,
				},
			},
		},
	}
	rt.Match.Headers[1] = &route.HeaderMatcher{
		Name: SeldonRouteHeader,
		HeaderMatchSpecifier: &route.HeaderMatcher_PresentMatch{
			PresentMatch: false,
		},
	}

	if isMirror {
		rt.Action = createWeightedPipelineClusterAction(r.Mirrors, []PipelineTrafficSplits{}, true)
	} else {
		rt.Action = createWeightedPipelineClusterAction(r.Clusters, r.Mirrors, true)
	}
}

func makePipelineGrpcRoute(r *PipelineRoute, rt *route.Route, isMirror bool) {
	rt.Name = getRouteName(r.RouteName, true, true, isMirror)
	rt.Match.PathSpecifier = pipelineRoutePathGrpc
	rt.Match.Headers[0] = &route.HeaderMatcher{
		Name: SeldonModelHeader, // Header name we will match on
		HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
			StringMatch: &matcherv3.StringMatcher{
				MatchPattern: &matcherv3.StringMatcher_Exact{
					Exact: r.RouteName,
				},
			},
		},
	}
	rt.Match.Headers[1] = &route.HeaderMatcher{
		Name: SeldonRouteHeader,
		HeaderMatchSpecifier: &route.HeaderMatcher_PresentMatch{
			PresentMatch: false,
		},
	}

	if isMirror {
		rt.Action = createWeightedPipelineClusterAction(r.Mirrors, []PipelineTrafficSplits{}, false)
	} else {
		rt.Action = createWeightedPipelineClusterAction(r.Clusters, r.Mirrors, false)
	}
}

func makePipelineStickySessionRoute(r *PipelineRoute, clusterTraffic *PipelineTrafficSplits, rt *route.Route, isGrpc bool) {
	if isGrpc {
		rt.Name = r.RouteName + "_grpc_experiment"
		rt.Match.PathSpecifier = pipelineRoutePathGrpc
	} else {
		rt.Name = r.RouteName + "_http_experiment"
		rt.Match.PathSpecifier = pipelineRoutePathHttp
	}

	rt.Match.Headers[0] = &route.HeaderMatcher{
		Name: SeldonRouteHeader, // Header name we will match on
		HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
			StringMatch: &matcherv3.StringMatcher{
				MatchPattern: &matcherv3.StringMatcher_Contains{
					Contains: wrapRouteHeader(getPipelineModelName(clusterTraffic.PipelineName)),
				},
			},
		},
	}
	rt.Match.Headers[1] = &route.HeaderMatcher{
		Name: SeldonModelHeader, // Header name we will match on
		HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
			StringMatch: &matcherv3.StringMatcher{
				MatchPattern: &matcherv3.StringMatcher_Exact{
					Exact: r.RouteName,
				},
			},
		},
	}
	rt.RequestHeadersToAdd = []*core.HeaderValueOption{
		{
			Header: &core.HeaderValue{
				Key:   SeldonInternalModelHeader,
				Value: getPipelineModelName(clusterTraffic.PipelineName),
			},
		},
	}
	rt.ResponseHeadersToAdd = []*core.HeaderValueOption{
		{
			Header: &core.HeaderValue{
				Key:   SeldonRouteHeader,
				Value: wrapRouteHeader(getPipelineModelName(clusterTraffic.PipelineName)),
			},
		},
	}
	if isGrpc {
		rt.Action = &route.Route_Route{
			Route: &route.RouteAction{
				Timeout: &duration.Duration{Seconds: DefaultRouteTimeoutSecs},
				ClusterSpecifier: &route.RouteAction_Cluster{
					Cluster: PipelineGatewayGrpcClusterName,
				},
			},
		}
	} else {
		rt.Action = &route.Route_Route{
			Route: &route.RouteAction{
				Timeout: &duration.Duration{Seconds: DefaultRouteTimeoutSecs},
				ClusterSpecifier: &route.RouteAction_Cluster{
					Cluster: PipelineGatewayHttpClusterName,
				},
			},
		}
	}
}

// This will allow sticky sessions for a) any experiment b) any in progression rollout
func isModelExperiment(r *Route) bool {
	return len(r.Clusters) > 1
}

func isPipelineExperiment(r *PipelineRoute) bool {
	return len(r.Clusters) > 1
}

func calcNumberOfModelStickySessionsNeeded(modelRoutes []*Route) int {
	count := 0
	for _, r := range modelRoutes {
		if isModelExperiment(r) {
			count = count + (len(r.Clusters) * 2) // REST and GRPC routes for each model in an experiment
		}
	}
	return count
}

func calcNumberOfPipelineStickySessionsNeeded(pipelineRoutes []*PipelineRoute) int {
	count := 0
	for _, r := range pipelineRoutes {
		if isPipelineExperiment(r) {
			count = count + (len(r.Clusters) * 2) // REST and GRPC routes for each model in an experiment
		}
	}
	return count
}

func calcNumberOfModelMirrorsNeeded(modelRoutes []*Route) int {
	count := 0
	for _, r := range modelRoutes {
		if len(r.Mirrors) > 0 {
			count = count + 2 // REST and gRPC
		}
	}
	return count
}

func calcNumberOfPipelineMirrorsNeeded(pipelineRoutes []*PipelineRoute) int {
	count := 0
	for _, r := range pipelineRoutes {
		if len(r.Mirrors) > 0 {
			count = count + 2 // REST and gRPC
		}
	}
	return count
}

func MakeRoute(modelRoutes []*Route, pipelineRoutes []*PipelineRoute) (*route.RouteConfiguration, *route.RouteConfiguration) {
	rts := make([]*route.Route, 2*(len(modelRoutes)+
		len(pipelineRoutes))+
		calcNumberOfModelStickySessionsNeeded(modelRoutes)+
		calcNumberOfPipelineStickySessionsNeeded(pipelineRoutes))
	// Pre-allocate objects for better CPU pipelining
	// Warning: assumes a fixes number of route-match headers
	for i := 0; i < len(rts); i++ {
		rts[i] = &route.Route{
			Match: &route.RouteMatch{
				Headers: make([]*route.HeaderMatcher, 2), // We always do 2 header matches
			},
		}
	}

	idx := 0

	// Create Model Routes
	for _, r := range modelRoutes {
		for _, clusterTraffic := range r.Clusters {
			if isModelExperiment(r) {
				makeModelStickySessionRoute(r, &clusterTraffic, rts[idx], false)
				idx++
				makeModelStickySessionRoute(r, &clusterTraffic, rts[idx], true)
				idx++
			}
		}
		makeModelHttpRoute(r, rts[idx], false)
		idx++
		makeModelGrpcRoute(r, rts[idx], false)
		idx++
	}

	// Create Pipeline Routes
	for _, r := range pipelineRoutes {
		if isPipelineExperiment(r) {
			for _, clusterTraffic := range r.Clusters {
				makePipelineStickySessionRoute(r, &clusterTraffic, rts[idx], false)
				idx++
				makePipelineStickySessionRoute(r, &clusterTraffic, rts[idx], true)
				idx++
			}
		}
		makePipelineHttpRoute(r, rts[idx], false)
		idx++
		makePipelineGrpcRoute(r, rts[idx], false)
		idx++
	}

	rtsMirrors := make([]*route.Route, calcNumberOfModelMirrorsNeeded(modelRoutes)+
		calcNumberOfPipelineMirrorsNeeded(pipelineRoutes))
	// Pre-allocate objects for better CPU pipelining
	// Warning: assumes a fixes number of route-match headers
	for i := 0; i < len(rtsMirrors); i++ {
		rtsMirrors[i] = &route.Route{
			Match: &route.RouteMatch{
				Headers: make([]*route.HeaderMatcher, 2), // We always do 2 header matches
			},
		}
	}

	idx = 0

	// Create Model Mirror Routes
	for _, r := range modelRoutes {
		if len(r.Mirrors) > 0 {
			makeModelHttpRoute(r, rtsMirrors[idx], true)
			idx++
			makeModelGrpcRoute(r, rtsMirrors[idx], true)
			idx++
		}
	}

	// Create Pipeline Mirror Routes
	for _, r := range pipelineRoutes {
		if len(r.Mirrors) > 0 {
			makePipelineHttpRoute(r, rtsMirrors[idx], true)
			idx++
			makePipelineGrpcRoute(r, rtsMirrors[idx], true)
			idx++
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

func createAccessLogConfig() *anypb.Any {
	accessFilter := accesslog_file.FileAccessLog{
		Path: "/tmp/envoy-accesslog.txt",
		AccessLogFormat: &accesslog_file.FileAccessLog_LogFormat{
			LogFormat: &core.SubstitutionFormatString{
				Format: &core.SubstitutionFormatString_TextFormatSource{
					TextFormatSource: &core.DataSource{
						Specifier: &core.DataSource_InlineString{
							InlineString: "[%START_TIME%] \"%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%\" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% \"%REQ(X-FORWARDED-FOR)%\" \"%REQ(USER-AGENT)%\" \"%REQ(X-REQUEST-ID)%\" \"%REQ(:AUTHORITY)%\" \"%UPSTREAM_HOST%\"\n",
						},
					},
				},
			},
		},
	}
	accessAny, err := anypb.New(&accessFilter)
	if err != nil {
		panic(err)
	}
	return accessAny
}

// A filter to add the seldon-model header from the http path if its not passed
// Does this for model and pipeline paths allowing us to keep our current header based routing
func createHeaderFilter() *anypb.Any {
	luaFilter := luav3.Lua{
		DefaultSourceCode: &core.DataSource{
			Specifier: &core.DataSource_InlineString{
				InlineString: `function envoy_on_request(request_handle)
  local modelHeader = request_handle:headers():get("` + SeldonModelHeader + `")
  local routeHeader = request_handle:headers():get("` + SeldonRouteHeader + `")
  if (modelHeader == nil or modelHeader == '') and (routeHeader == nil or routeHeader == '') then
    local path = request_handle:headers():get(":path")
    local i, j = string.find(path,"/v2/models/")
    if i == 1 then
      local s = string.sub(path,j+1)
      i, j = string.find(s, "/")
      if i then
        local model = string.sub(s,0,i-1)
        request_handle:headers():add("` + SeldonModelHeader + `",model)
      else
        request_handle:headers():add("` + SeldonModelHeader + `",s)
      end
    else
      i, j = string.find(path,"/v2/pipelines/")
      if i == 1 then
        local s = string.sub(path,j+1)
        i, j = string.find(s, "/")
        local model = string.sub(s,0,i-1)
        request_handle:headers():add("` + SeldonModelHeader + `",model..".` + SeldonPipelineHeaderSuffix + `")
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

func MakeHTTPListener(listenerName, address string,
	port uint32,
	routeConfigurationName string,
	serverSecret *Secret) *listener.Listener {
	routerConfig, _ := anypb.New(&router.Router{})
	// HTTP filter configuration
	manager := &hcm.HttpConnectionManager{
		CodecType:                    hcm.HttpConnectionManager_AUTO,
		StatPrefix:                   "http",
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
		AccessLog: []*accesslog.AccessLog{
			{
				Name: "envoy.access_loggers.file",
				ConfigType: &accesslog.AccessLog_TypedConfig{
					TypedConfig: createAccessLogConfig(),
				},
			},
		},
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

func makeConfigSource() *core.ConfigSource {
	source := &core.ConfigSource{}
	source.ResourceApiVersion = resource.DefaultAPIVersion
	source.ConfigSourceSpecifier = &core.ConfigSource_ApiConfigSource{
		ApiConfigSource: &core.ApiConfigSource{
			TransportApiVersion:       resource.DefaultAPIVersion,
			ApiType:                   core.ApiConfigSource_DELTA_GRPC,
			SetNodeOnFirstMessageOnly: true,
			GrpcServices: []*core.GrpcService{{
				TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &core.GrpcService_EnvoyGrpc{ClusterName: "xds_cluster"},
				},
			}},
		},
	}
	return source
}
