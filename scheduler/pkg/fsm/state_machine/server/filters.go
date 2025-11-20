package server

import (
	"slices"
	"strings"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/model"
)

type Filter string

const (
	FilterReplicas      Filter = "filter_replicas"
	FilterDeletedServer Filter = "filter_deleted_server"
)

func (ss *Snapshot) FilterServer(model model.VersionStatus) {

}

func (ss *Snapshot) Filter(filter Filter) bool {
	switch filter {
	case FilterReplicas:
		return len(ss.Replicas) > 0
	case FilterDeletedServer:
		return ss.ExpectedReplicas != 0
	}

	return false
}

func (ss *Snapshot) FilterSharingServer(model model.VersionStatus) bool {
	requestedServer := model.ModelDefn.GetModelSpec().Server
	return (requestedServer == nil && ss.Shared) || (requestedServer != nil && *requestedServer == ss.Name)
}

func (ss *Snapshot) FilterServerRequirements(model model.VersionStatus) bool {
	if len(ss.Replicas) == 0 {
		// Capabilities are currently stored on replicas, so no replicas means no capabilities can be determined.
		return false
	}

	requirements := model.ModelDefn.GetModelSpec().GetRequirements()

	capabilities := ss.getCapabilities()
	for _, req := range requirements {
		if slices.Contains(capabilities, strings.TrimSpace(req)) {
			return true // Server has at least one required capability
		}
	}

	return false // Server doesn't have any of the required capabilities
}

func (ss *Snapshot) getCapabilities() []string {
	for _, replica := range ss.Replicas {
		return replica.capabilities
	}

	return []string{}
}
