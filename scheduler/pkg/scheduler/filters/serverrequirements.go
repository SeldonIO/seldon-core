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

	// get the capabilities of the first available replica
	capabilities := getFirstAvailableReplicaCapabilities(server.Replicas)

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

func getFirstAvailableReplicaCapabilities(replicas map[int]*store.ServerReplica) []string {
	for _, replica := range replicas {
		return replica.GetCapabilities()
	}
	return []string{}
}

func (s ServerRequirementFilter) Description(model *store.ModelVersion, server *store.ServerSnapshot) string {
	requirements := model.GetRequirements()

	replicas := server.Replicas
	if len(replicas) == 0 {
		return fmt.Sprintf("model requirements %v, server capabilities unknown", requirements)
	}

	capabilities := getFirstAvailableReplicaCapabilities(replicas)
	return fmt.Sprintf("model requirements %v, server capabilities %v", requirements, capabilities)
}
