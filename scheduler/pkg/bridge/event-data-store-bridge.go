package bridge

import "github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"

type EventDataStoreBridge interface {
	AddServerReplica(models []coordinator.ModelEventMsg, msg coordinator.ServerEventMsg) error
}
