package constants

const (
	PU_PARAMETER_ENVVAR    = "PREDICTIVE_UNIT_PARAMETERS"
	TFServingContainerName = "tfserving"

	GRPCRegExMatchAmbassador = "/(seldon.protos.*|tensorflow.serving.*)/.*"
	GRPCRegExMatchIstio      = ".*tensorflow.*|.*seldon.protos.*"

	PrePackedServerTensorflow = "TENSORFLOW_SERVER"
	PrePackedServerSklearn    = "SKLEARN_SERVER"

	TfServingGrpcPort    = 2000
	TfServingRestPort    = 2001
	TfServingArgPort     = "--port="
	TfServingArgRestPort = "--rest_api_port="

	FirstPortNumber = 9000
)

const (
	ControllerName = "seldon-controller-manager"
)

// Event messages
const (
	EventsCreateVirtualService  = "CreateVirtualService"
	EventsUpdateVirtualService  = "UpdateVirtualService"
	EventsCreateDestinationRule = "CreateDestinationRule"
	EventsUpdateDestinationRule = "UpdateDestinationRule"
	EventsCreateService         = "CreateService"
	EventsUpdateService         = "UpdateService"
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
	ExplainerPathSuffix = "/explainer"
	ExplainerNameSuffix = "-explainer"
)
