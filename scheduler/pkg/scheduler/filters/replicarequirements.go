package filters

import "github.com/seldonio/seldon-core/scheduler/pkg/store"

type RequirementsReplicaFilter struct{}

func (s RequirementsReplicaFilter) Filter(model *store.ModelVersion, replica *store.ServerReplica) bool {
	for _,requirement := range model.GetRequirements() {
		requirementFound := false
		for _, capability := range replica.GetCapabilities() {
			if requirement == capability {
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



