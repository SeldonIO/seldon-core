package resources

type Listener struct {
	Name       string
	Address    string
	Port       uint32
	//RouteNames []string
}

type Route struct {
	Name    string
	Host    string
	Cluster string
}

type Cluster struct {
	Name      string
	Endpoints map[string]Endpoint
	Routes map[string]bool
}

type Endpoint struct {
	UpstreamHost string
	UpstreamPort uint32
}
