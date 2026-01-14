/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package util

import (
	"strings"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"
)

func NewTestServerReplica(inferenceSvc string,
	inferenceHttpPort int32,
	inferenceGrpcPort int32,
	replicaIdx int32,
	server *db.Server,
	capabilities []string,
	memory,
	availableMemory,
	reservedMemory uint64,
	loadedModels []*db.ModelVersionID,
	overCommitPercentage uint32,
) *db.ServerReplica {
	cleanCapabilities := func(capabilities []string) []string {
		var cleaned []string
		for _, capability := range capabilities {
			cleaned = append(cleaned, strings.TrimSpace(capability))
		}
		return cleaned
	}

	return &db.ServerReplica{
		InferenceSvc:         inferenceSvc,
		InferenceHttpPort:    inferenceHttpPort,
		InferenceGrpcPort:    inferenceGrpcPort,
		ServerName:           server.Name,
		ReplicaIdx:           replicaIdx,
		Capabilities:         cleanCapabilities(capabilities),
		Memory:               memory,
		AvailableMemory:      availableMemory,
		ReservedMemory:       reservedMemory,
		LoadedModels:         loadedModels,
		LoadingModels:        make([]*db.ModelVersionID, 0),
		OverCommitPercentage: overCommitPercentage,
		IsDraining:           false,
	}
}
