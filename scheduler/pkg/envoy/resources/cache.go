package resources

const (
	PipelineGatewayHttpClusterName = "pipelinegateway_http"
	PipelineGatewayGrpcClusterName = "pipelinegateway_grpc"
)

type Listener struct {
	Name    string
	Address string
	Port    uint32
	//RouteNames []string
}

type Route struct {
	RouteName   string
	LogPayloads bool
	Clusters    []TrafficSplits
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
	PipelineName string
}
