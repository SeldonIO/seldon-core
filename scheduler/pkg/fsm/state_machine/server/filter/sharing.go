package filter

import (
	"fmt"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/model"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/server"
)

type SharingServerFilter struct{}

func (e SharingServerFilter) Name() string {
	return "SharingServerFilter"
}

func (e SharingServerFilter) Filter(model *model.VersionStatus, server *server.Snapshot) bool {
	requestedServer := model.ModelDefn.GetModelSpec().Server
	return (requestedServer == nil && server.Shared) || (requestedServer != nil && *requestedServer == server.Name)
}

func (e SharingServerFilter) Description(model *model.VersionStatus, server *server.Snapshot) string {
	requestedServer := model.ModelDefn.GetModelSpec().Server
	if requestedServer != nil {
		return fmt.Sprintf("requested server %s == %s", *requestedServer, server.Name)
	} else {
		return fmt.Sprintf("sharing %v", server.Shared)
	}
}
