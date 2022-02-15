package resources

type Listener struct {
	Name    string
	Address string
	Port    uint32
	//RouteNames []string
}

type Route struct {
	ModelName   string
	LogPayloads bool
	Clusters    []TrafficSplits
}

type TrafficSplits struct {
	ModelName      string
	ModelVersion   uint32
	TrafficPercent uint32
	HttpCluster    string
	GrpcCluster    string
}

type Cluster struct {
	Name      string
	Grpc      bool
	Endpoints map[string]Endpoint
	Routes    map[string]bool
}

type Endpoint struct {
	UpstreamHost string
	UpstreamPort uint32
}
