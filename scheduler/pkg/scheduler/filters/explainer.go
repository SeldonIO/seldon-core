/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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
