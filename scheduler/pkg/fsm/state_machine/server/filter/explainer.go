package filter

import (
	"fmt"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/model"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/server"
)

const (
	alibiExplainerRequiredCapability = "alibi-explain"
)

type ExplainerFilter struct{}

func (s ExplainerFilter) Name() string {
	return "ExplainerFilter"
}

func (s ExplainerFilter) Filter(modelVersion *model.VersionStatus, replica *server.Replica) bool {
	if modelVersion.ModelDefn.ModelSpec.GetExplainer() != nil {
		for _, capability := range replica.Capabilities {
			if alibiExplainerRequiredCapability == capability {
				return true
			}
		}
		return false
	}
	return true
}

func (s ExplainerFilter) Description(modelVersion *model.VersionStatus, replica *server.Replica) string {
	return fmt.Sprintf("model is explainer %v replica capabilities %v", modelVersion.ModelDefn.ModelSpec.GetExplainer() == nil, replica.Capabilities)
}
