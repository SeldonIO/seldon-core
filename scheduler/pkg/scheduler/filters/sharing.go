package filters

import (
	"fmt"

	"github.com/seldonio/seldon-core/scheduler/pkg/store"
)

type SharingServerFilter struct{}

func (e SharingServerFilter) Name() string {
	return "SharingServerFilter"
}

func (e SharingServerFilter) Filter(model *store.ModelVersion, server *store.ServerSnapshot) bool {
	requestedServer := model.GetRequestedServer()
	return (requestedServer == nil && server.Shared) || (requestedServer != nil && *requestedServer == server.Name)
}

func (e SharingServerFilter) Description(model *store.ModelVersion, server *store.ServerSnapshot) string {
	requestedServer := model.GetRequestedServer()
	if requestedServer != nil {
		return fmt.Sprintf("requested server %s == %s", *requestedServer, server.Name)
	} else {
		return fmt.Sprintf("sharing %v", server.Shared)
	}
}
