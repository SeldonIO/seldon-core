/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package filters

import (
	"fmt"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
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
