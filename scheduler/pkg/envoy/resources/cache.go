package resources

import "github.com/seldonio/seldon-core-v2/components/tls/pkg/tls"

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
	//RouteNames []string
}

type Route struct {
	RouteName   string
	LogPayloads bool
	Clusters    []TrafficSplits
	Mirrors     []TrafficSplits
}

type TrafficSplits struct {
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
	Clusters  []PipelineTrafficSplits
	Mirrors   []PipelineTrafficSplits
}

type PipelineTrafficSplits struct {
	PipelineName  string
	TrafficWeight uint32
}

type Secret struct {
	Name                 string
	ValidationSecretName string
	Certificate          tls.CertificateStoreHandler
}
