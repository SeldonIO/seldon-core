package filters

import (
	"fmt"

	"github.com/seldonio/seldon-core/scheduler/pkg/store"
)

const (
	alibiExplainerRequiredCapability = "alibi-explain"
)

type ExplainerFilter struct{}

func (s ExplainerFilter) Name() string {
	return "ExplainerFilter"
}

func (s ExplainerFilter) Filter(model *store.ModelVersion, replica *store.ServerReplica) bool {
	if model.GetModel().GetModelSpec().GetExplainer() != nil {
		for _, capability := range replica.GetCapabilities() {
			if alibiExplainerRequiredCapability == capability {
				return true
			}
		}
		return false
	}
	return true
}

func (s ExplainerFilter) Description(model *store.ModelVersion, replica *store.ServerReplica) string {
	return fmt.Sprintf("model is explainer %v replica capabilities %v", model.GetModel().GetModelSpec().GetExplainer() == nil, replica.GetCapabilities())
}
