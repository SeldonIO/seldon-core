package filter

import (
	"fmt"
	"slices"
	"strings"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/model"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/server"
)

type ServerRequirementFilter struct{}

func (s ServerRequirementFilter) Name() string {
	return "ServerRequirementsFilter"
}

func (s ServerRequirementFilter) Filter(model *model.VersionStatus, server *server.Snapshot) bool {
	if len(server.Replicas) == 0 {
		// Capabilities are currently stored on replicas, so no replicas means no capabilities can be determined.
		return false
	}

	requirements := model.ModelDefn.GetModelSpec().GetRequirements()

	capabilities := server.GetCapabilities()
	for _, req := range requirements {
		if slices.Contains(capabilities, strings.TrimSpace(req)) {
			return true // Server has at least one required capability
		}
	}

	return false // Server doesn't have any of the required capabilities
}

func (s ServerRequirementFilter) Description(model *model.VersionStatus, server *server.Snapshot) string {
	requirements := model.ModelDefn.GetModelSpec().GetRequirements()

	replicas := server.Replicas
	if len(replicas) == 0 {
		return fmt.Sprintf("model requirements %v, server capabilities unknown", requirements)
	}

	capabilities := server.GetCapabilities()
	return fmt.Sprintf("model requirements %v, server capabilities %v", requirements, capabilities)
}
