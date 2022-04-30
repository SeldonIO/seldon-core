package filters

import (
	"strings"

	"github.com/seldonio/seldon-core/scheduler/pkg/store"
)

type RequirementsReplicaFilter struct{}

func (s RequirementsReplicaFilter) Name() string {
	return "RequirementsReplicaFilter"
}

func (s RequirementsReplicaFilter) Filter(model *store.ModelVersion, replica *store.ServerReplica) bool {
	for _, requirement := range model.GetRequirements() {
		requirementFound := false
		for _, capability := range replica.GetCapabilities() {
			if strings.TrimSpace(requirement) == capability {
				requirementFound = true
				break
			}
		}
		if !requirementFound {
			return false
		}
	}
	return true
}
