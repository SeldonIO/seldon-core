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
	"time"

	"github.com/golang/protobuf/ptypes/duration"

	matcher "github.com/envoyproxy/go-control-plane/envoy/config/common/matcher/v3"
	envoy_extensions_common_tap_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/common/tap/v3"
	"github.com/seldonio/seldon-core/scheduler/pkg/util"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	tap "github.com/envoyproxy/go-control-plane/envoy/config/tap/v3"
	accesslog_file "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/file/v3"
	tapfilter "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/tap/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	http "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	wrappers "github.com/golang/protobuf/ptypes/wrappers"
)

const (
	RouteConfigurationName     = "listener_0"
	SeldonLoggingHeader        = "Seldon-Logging"
	EnvoyLogPathPrefix         = "/tmp/request-log"
	SeldonModelHeader          = "seldon-model"
	SeldonPipelineHeader       = "pipeline"
	SeldonInternalModelHeader  = "seldon-internal-model"
	SeldonRouteHeader          = "seldon-route"
	SeldonModelHeaderSuffix    = "model"
	SeldonPipelineHeaderSuffix = "pipeline"
	DefaultRouteTimeoutSecs    = 60 //TODO allow configurable override
)

func MakeCluster(clusterName string, eps []Endpoint, isGrpc bool) *cluster.Cluster {
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
		}
	} else {
		return &cluster.Cluster{
			Name:                 clusterName,
			ConnectTimeout:       durationpb.New(5 * time.Second),
			ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_STRICT_DNS},
			LbPolicy:             cluster.Cluster_LEAST_REQUEST,
			LoadAssignment:       MakeEndpoint(clusterName, eps),
			DnsLookupFamily:      cluster.Cluster_V4_ONLY,
		}
	}
}

func makeEDSCluster() *cluster.Cluster_EdsClusterConfig {
	return &cluster.Cluster_EdsClusterConfig{
		EdsConfig: makeConfigSource(),
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

// weighted clusters do not play well with session affinity see https://github.com/envoyproxy/envoy/issues/8167
// Traffic shifting may need to be reinvesigated https://github.com/envoyproxy/envoy/pull/18207
func createWeightedClusterAction(clusterTraffics []TrafficSplits, rest bool) *route.Route_Route {
	// Add Weighted Clusters with given traffic percentages to each internal model
	var clusters []*route.WeightedCluster_ClusterWeight
	var totWeight uint32
	for _, clusterTraffic := range clusterTraffics {
		clusterName := clusterTraffic.HttpCluster
		if !rest {
			clusterName = clusterTraffic.GrpcCluster
		}
		totWeight = totWeight + clusterTraffic.TrafficWeight
		clusters = append(clusters,
			&route.WeightedCluster_ClusterWeight{
				Name: clusterName,
				Weight: &wrappers.UInt32Value{
					Value: clusterTraffic.TrafficWeight,
				},
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
							Value: util.XXHash(util.GetVersionedModelName(
								clusterTraffic.ModelName, clusterTraffic.ModelVersion)),
						},
					},
				},
			})

	}
	action := &route.Route_Route{
		Route: &route.RouteAction{
			Timeout: &duration.Duration{Seconds: DefaultRouteTimeoutSecs},
			ClusterSpecifier: &route.RouteAction_WeightedClusters{
				WeightedClusters: &route.WeightedCluster{
					Clusters:    clusters,
					TotalWeight: &wrappers.UInt32Value{Value: totWeight},
				},
			},
		},
	}
	return action
}

var modelRouteMatchPathHttp = &route.RouteMatch_Prefix{Prefix: "/v2"}
var modelRouteMatchPathGrpc = &route.RouteMatch_Prefix{Prefix: "/inference.GRPCInferenceService"}
var modelRouteHeaders = []*core.HeaderValueOption{
	{Header: &core.HeaderValue{Key: SeldonLoggingHeader, Value: "true"}},
}

func makeModelHttpRoute(r *Route, rt *route.Route) {
	rt.Name = r.RouteName + "_http"
	rt.Match.PathSpecifier = modelRouteMatchPathHttp
	rt.Match.Headers[0] = &route.HeaderMatcher{
		Name: SeldonModelHeader, // Header name we will match on
		HeaderMatchSpecifier: &route.HeaderMatcher_ExactMatch{
			ExactMatch: r.RouteName,
		},
		//TODO: https://github.com/envoyproxy/envoy/blob/c75c1410c8682cb44c9136ce4ad01e6a58e16e8e/api/envoy/api/v2/route/route_components.proto#L1513
		//HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
		//	StringMatch: &matcher.StringMatcher{
		//		MatchPattern: &matcher.StringMatcher_Exact{
		//			Exact: r.Host,
		//		},
		//	},
		//},
	}
	rt.Action = createWeightedClusterAction(r.Clusters, true)
	if r.LogPayloads {
		rt.ResponseHeadersToAdd = modelRouteHeaders
	}
}

func makeModelExperimentRoute(r *Route, clusterTraffic *TrafficSplits, rt *route.Route, isGrpc bool) {
	if isGrpc {
		rt.Name = r.RouteName + "_grpc_experiment"
		rt.Match.PathSpecifier = modelRouteMatchPathGrpc
	} else {
		rt.Name = r.RouteName + "_http_experiment"
		rt.Match.PathSpecifier = modelRouteMatchPathHttp
	}

	rt.Match.Headers[0] = &route.HeaderMatcher{
		Name: SeldonRouteHeader, // Header name we will match on
		HeaderMatchSpecifier: &route.HeaderMatcher_ExactMatch{
			ExactMatch: util.XXHash(util.GetVersionedModelName(
				clusterTraffic.ModelName, clusterTraffic.ModelVersion)),
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
		rt.ResponseHeadersToAdd = modelRouteHeaders
	}
}

func makeModelGrpcRoute(r *Route, rt *route.Route) {
	//TODO there is no easy way to implement version specific gRPC calls so this could mean we need to implement
	//latest model policy on V2 servers and therefore also for REST as well
	rt.Name = r.RouteName + "_grpc"
	rt.Match.PathSpecifier = modelRouteMatchPathGrpc
	rt.Match.Headers[0] = &route.HeaderMatcher{
		Name: SeldonModelHeader, // Header name we will match on
		HeaderMatchSpecifier: &route.HeaderMatcher_ExactMatch{
			ExactMatch: r.RouteName,
		},
		//TODO: https://github.com/envoyproxy/envoy/blob/c75c1410c8682cb44c9136ce4ad01e6a58e16e8e/api/envoy/api/v2/route/route_components.proto#L1513
		//HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
		//	StringMatch: &matcher.StringMatcher{
		//		MatchPattern: &matcher.StringMatcher_Exact{
		//			Exact: r.Host,
		//		},
		//	},
		//},
	}
	rt.Action = createWeightedClusterAction(r.Clusters, false)
	if r.LogPayloads {
		rt.ResponseHeadersToAdd = modelRouteHeaders
	}
}

var pipelineRoutePathHttp = &route.RouteMatch_Prefix{Prefix: "/v2"}
var pipelineRoutePathGrpc = &route.RouteMatch_Prefix{Prefix: "/inference.GRPCInferenceService"}
var pipelineRouteActionHttp = &route.Route_Route{
	Route: &route.RouteAction{
		Timeout: &duration.Duration{Seconds: DefaultRouteTimeoutSecs},
		ClusterSpecifier: &route.RouteAction_Cluster{
			Cluster: PipelineGatewayHttpClusterName,
		},
	},
}
var pipelineRouteActionGrpc = &route.Route_Route{
	Route: &route.RouteAction{
		Timeout: &duration.Duration{Seconds: DefaultRouteTimeoutSecs},
		ClusterSpecifier: &route.RouteAction_Cluster{
			Cluster: PipelineGatewayGrpcClusterName,
		},
	},
}

func makePipelineHttpRoute(r *PipelineRoute, rt *route.Route) {
	rt.Name = r.PipelineName + "_pipeline_http"
	rt.Match.PathSpecifier = pipelineRoutePathHttp
	rt.Match.Headers[0] = &route.HeaderMatcher{
		Name: SeldonModelHeader, // Header name we will match on
		HeaderMatchSpecifier: &route.HeaderMatcher_ExactMatch{
			ExactMatch: r.PipelineName + ".pipeline",
		},
	}
	rt.Action = pipelineRouteActionHttp
}

func makePipelineGrpcRoute(r *PipelineRoute, rt *route.Route) {
	rt.Name = r.PipelineName + "_pipeline_grpc"
	rt.Match.PathSpecifier = pipelineRoutePathGrpc
	rt.Match.Headers[0] = &route.HeaderMatcher{
		Name: SeldonModelHeader, // Header name we will match on
		HeaderMatchSpecifier: &route.HeaderMatcher_ExactMatch{
			ExactMatch: r.PipelineName + ".pipeline",
		},
	}
	rt.Action = pipelineRouteActionGrpc
}

func isExperiment(r *Route) bool {
	return len(r.Clusters) > 1
}

func calcNumberOfStickySessionsNeeded(modelRoutes []*Route) int {
	count := 0
	for _, r := range modelRoutes {
		if isExperiment(r) {
			count = count + (len(r.Clusters) * 2) // REST and GRPC routes for each model in an experiment
		}
	}
	return count
}

func MakeRoute(modelRoutes []*Route, pipelineRoutes []*PipelineRoute) *route.RouteConfiguration {
	rts := make([]*route.Route, 2*(len(modelRoutes)+len(pipelineRoutes))+calcNumberOfStickySessionsNeeded(modelRoutes))
	// Pre-allocate objects for better CPU pipelining
	// Warning: assumes a fixes number of route-match headers
	for i := 0; i < len(rts); i++ {
		rts[i] = &route.Route{
			Match: &route.RouteMatch{
				Headers: make([]*route.HeaderMatcher, 1),
			},
		}
	}

	idx := 0

	// Create Model Routes
	for _, r := range modelRoutes {
		makeModelHttpRoute(r, rts[idx])
		idx++
		makeModelGrpcRoute(r, rts[idx])
		idx++
		if isExperiment(r) {
			for _, clusterTraffic := range r.Clusters {
				makeModelExperimentRoute(r, &clusterTraffic, rts[idx], false)
				idx++
				makeModelExperimentRoute(r, &clusterTraffic, rts[idx], true)
				idx++
			}
		}
	}

	// Create Pipeline Routes
	for _, r := range pipelineRoutes {
		makePipelineHttpRoute(r, rts[idx])
		idx++
		makePipelineGrpcRoute(r, rts[idx])
		idx++
	}

	return &route.RouteConfiguration{
		Name: RouteConfigurationName,
		VirtualHosts: []*route.VirtualHost{{
			Name:    "seldon_service",
			Domains: []string{"*"},
			Routes:  rts,
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

		/*
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
		*/
	}

	accessAny, err := anypb.New(&accessFilter)
	if err != nil {
		panic(err)
	}
	return accessAny
}

func MakeHTTPListener(listenerName, address string, port uint32) *listener.Listener {

	// HTTP filter configuration
	manager := &hcm.HttpConnectionManager{
		CodecType:                    hcm.HttpConnectionManager_AUTO,
		StatPrefix:                   "http",
		AlwaysSetRequestIdInResponse: true,
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				ConfigSource:    makeConfigSource(),
				RouteConfigName: RouteConfigurationName,
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
				Name: wellknown.Router,
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
		FilterChains: []*listener.FilterChain{{
			Filters: []*listener.Filter{
				{
					Name: wellknown.HTTPConnectionManager,
					ConfigType: &listener.Filter_TypedConfig{
						TypedConfig: pbst,
					},
				},
			},
		}},
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
