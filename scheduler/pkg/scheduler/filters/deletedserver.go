package filters

import (
	"fmt"

	"github.com/seldonio/seldon-core/scheduler/pkg/store"
)

type DeletedServerFilter struct{}

func (e DeletedServerFilter) Name() string {
	return "DeletedServerFilter"
}

func (e DeletedServerFilter) Filter(model *store.ModelVersion, server *store.ServerSnapshot) bool {
	return server.ExpectedReplicas != 0
}

func (e DeletedServerFilter) Description(model *store.ModelVersion, server *store.ServerSnapshot) string {
	return fmt.Sprintf("expected replicas %d != 0", server.ExpectedReplicas)
}
