package filter

import (
	"fmt"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/model"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/server"
)

type DeletedServerFilter struct{}

func (e DeletedServerFilter) Name() string {
	return "DeletedServerFilter"
}

func (e DeletedServerFilter) Filter(model *model.VersionStatus, server *server.Snapshot) bool {
	return server.ExpectedReplicas != 0
}

func (e DeletedServerFilter) Description(model *model.VersionStatus, server *server.Snapshot) string {
	return fmt.Sprintf("expected replicas %d != 0", server.ExpectedReplicas)
}
