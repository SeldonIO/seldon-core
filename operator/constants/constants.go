package constants

const (
	PU_PARAMETER_ENVVAR    = "PREDICTIVE_UNIT_PARAMETERS"
	TFServingContainerName = "tfserving"

	GRPCRegExMatchAmbassador = "/(seldon.protos.*|tensorflow.serving.*)/.*"
	GRPCRegExMatchIstio      = ".*tensorflow.*|.*seldon.protos.*"
	GRPCPathPrefixTensorflow = "/tensorflow.serving."
	GRPCPathPrefixSeldon     = "/seldon.protos."

	PrePackedServerTensorflow = "TENSORFLOW_SERVER"
	PrePackedServerSklearn    = "SKLEARN_SERVER"

	TfServingGrpcPort    = 2000
	TfServingRestPort    = 2001
	TfServingArgPort     = "--port="
	TfServingArgRestPort = "--rest_api_port="

	FirstPortNumber       = int32(9000)
	DNSLocalHost          = "localhost"
	DNSClusterLocalSuffix = ".svc.cluster.local."
	GrpcPortName          = "grpc"
	HttpPortName          = "http"
)

// Metrics-related constants
const (
	FirstMetricsPortNumber = int32(6000)
	DefaultMetricsPortName = "metrics"
)

const (
	ControllerName = "seldon-controller-manager"
)

// Event messages
const (
	EventsCreateVirtualService  = "CreateVirtualService"
	EventsUpdateVirtualService  = "UpdateVirtualService"
	EventsDeleteVirtualService  = "DeleteVirtualService"
	EventsCreateDestinationRule = "CreateDestinationRule"
	EventsUpdateDestinationRule = "UpdateDestinationRule"
	EventsCreateHTTPProxy       = "CreateHTTPProxy"
	EventsUpdateHTTPProxy       = "UpdateHTTPProxy"
	EventsDeleteHTTPProxy       = "DeleteHTTPProxy"
	EventsCreateService         = "CreateService"
	EventsUpdateService         = "UpdateService"
	EventsDeleteService         = "DeleteService"
	EventsCreateHPA             = "CreateHPA"
	EventsUpdateHPA             = "UpdateHPA"
	EventsDeleteHPA             = "DeleteHPA"
	EventsCreateDeployment      = "CreateDeployment"
	EventsUpdateDeployment      = "UpdateDeployment"
	EventsDeleteDeployment      = "DeleteDeployment"
	EventsInternalError         = "InternalError"
	EventsUpdated               = "Updated"
	EventsUpdateFailed          = "UpdateFailed"
)

// Explainers
const (
	ExplainerPathSuffix = "-explainer"
	ExplainerNameSuffix = "-explainer"
)
