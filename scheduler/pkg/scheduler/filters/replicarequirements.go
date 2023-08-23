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

type ServerRequirementFilter struct{}

func (s ServerRequirementFilter) Name() string {
	return "ServerRequirementsFilter"
}

func (s ServerRequirementFilter) Filter(model *store.ModelVersion, server *store.ServerSnapshot) bool {
	if len(server.Replicas) == 0 {
		// Capabilities are currently stored on replicas, so no replicas means no capabilities can be determined.
		return false
	}

	requirements := model.GetRequirements()
	capabilities := server.Replicas[0].GetCapabilities()

	for _, r := range requirements {
		if !contains(capabilities, r) {
			return false
		}
	}

	return true
}

func contains(capabilities []string, requirement string) bool {
	for _, c := range capabilities {
		if c == strings.TrimSpace(requirement) {
			return true
		}
	}

	return false
}

func (s ServerRequirementFilter) Description(model *store.ModelVersion, server *store.ServerSnapshot) string {
	requirements := model.GetRequirements()
	capabilities := []string{}

	replicas := server.Replicas
	if len(replicas) > 0 {
		capabilities = replicas[0].GetCapabilities()
	}

	return fmt.Sprintf("model requirements %v server capabilities %v", requirements, capabilities)
}
