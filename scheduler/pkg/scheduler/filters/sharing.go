package filters

import "github.com/seldonio/seldon-core/scheduler/pkg/store"

type SharingServerFilter struct{}

func (e SharingServerFilter) Name() string {
	return "SharingServerFilter"
}

func (e SharingServerFilter) Filter(model *store.ModelVersion, server *store.ServerSnapshot) bool {
	requestedServer := model.GetRequestedServer()
	return (requestedServer == nil && server.Shared) || (requestedServer != nil && *requestedServer == server.Name)
}
