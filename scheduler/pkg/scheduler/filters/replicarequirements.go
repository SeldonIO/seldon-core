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
	"strings"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
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

func (s RequirementsReplicaFilter) Description(model *store.ModelVersion, replica *store.ServerReplica) string {
	return fmt.Sprintf("model requirements %v replica capabilities %v", model.GetRequirements(), replica.GetCapabilities())
}
