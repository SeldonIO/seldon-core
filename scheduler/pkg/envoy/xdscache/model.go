/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package xdscache

import "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

const (
	PipelineGatewayHttpClusterName = "pipelinegateway_http"
	PipelineGatewayGrpcClusterName = "pipelinegateway_grpc"
	MirrorHttpClusterName          = "mirror_http"
	MirrorGrpcClusterName          = "mirror_grpc"
)

type Listener struct {
	Name                   string
	Address                string
	Port                   uint32
	RouteConfigurationName string
}

type Route struct {
	RouteName   string
	LogPayloads bool
	Clusters    []TrafficSplit
	Mirror      *TrafficSplit
}

type TrafficSplit struct {
	ModelName     string
	ModelVersion  uint32
	TrafficWeight uint32
	HttpCluster   string
	GrpcCluster   string
}

type RouteVersionKey struct {
	RouteName string
	ModelName string
	Version   uint32
}

type Cluster struct {
	Name      string
	Grpc      bool
	Endpoints map[string]Endpoint
	Routes    map[RouteVersionKey]bool
}

type Endpoint struct {
	UpstreamHost string
	UpstreamPort uint32
}

type PipelineRoute struct {
	RouteName string
	Clusters  []PipelineTrafficSplit
	Mirror    *PipelineTrafficSplit
}

type PipelineTrafficSplit struct {
	PipelineName  string
	TrafficWeight uint32
}

type Secret struct {
	Name                 string
	ValidationSecretName string
	Certificate          tls.CertificateStoreHandler
}
