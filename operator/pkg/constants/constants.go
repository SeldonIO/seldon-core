package constants

import "os"

const (
	ModelFinalizerName  = "seldon.model.finalizer"
	ServerFinalizerName = "seldon.server.finalizer"
)

func getEnvOrDefault(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var (
	SeldonNamespace = getEnvOrDefault("POD_NAMESPACE", "seldon-mesh")
)

// Label selector
const (
	AppKey                    = "app"
	ServerLabelValue          = "seldon-server"
	ServerLabelNameKey        = "seldon-server-name"
	ServerReplicaLabelKey     = "seldon-server-replica"
	ServerReplicaNameLabelKey = "seldon-server-replica-name"
)

// Reconcilliation operations
type ReconcileOperation uint32

const (
	ReconcileUnknown ReconcileOperation = iota
	ReconcileNoChange
	ReconcileUpdateNeeded
	ReconcileCreateNeeded
)
