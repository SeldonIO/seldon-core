package filters

import "github.com/seldonio/seldon-core/scheduler/pkg/store"

type DeletedServerFilter struct{}

func (e DeletedServerFilter) Name() string {
	return "DeletedServerFilter"
}

func (e DeletedServerFilter) Filter(model *store.ModelVersion, server *store.ServerSnapshot) bool {
	return server.ExpectedReplicas != 0
}
