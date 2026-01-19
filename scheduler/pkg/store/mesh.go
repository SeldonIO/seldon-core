/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package store

import (
	"strings"

	pba "github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"
)

func NewServer(name string, shared bool) *db.Server {
	return &db.Server{
		Name:             name,
		Replicas:         make(map[int32]*db.ServerReplica),
		Shared:           shared,
		ExpectedReplicas: -1,
	}
}

func NewServerReplicaFromConfig(server *db.Server, replicaIdx int, loadedModels []*db.ModelVersionID, config *pba.ReplicaConfig, availableMemoryBytes uint64) *db.ServerReplica {
	return &db.ServerReplica{
		InferenceSvc:         config.GetInferenceSvc(),
		InferenceHttpPort:    config.GetInferenceHttpPort(),
		InferenceGrpcPort:    config.GetInferenceGrpcPort(),
		ServerName:           server.Name,
		ReplicaIdx:           int32(replicaIdx),
		Capabilities:         cleanCapabilities(config.GetCapabilities()),
		Memory:               config.GetMemoryBytes(),
		AvailableMemory:      availableMemoryBytes,
		LoadedModels:         loadedModels,
		LoadingModels:        make([]*db.ModelVersionID, 0),
		OverCommitPercentage: config.GetOverCommitPercentage(),
		UniqueLoadedModels:   toUniqueModels(loadedModels),
		IsDraining:           false,
	}
}

func NewDefaultModelVersion(model *pb.Model, version uint32) *db.ModelVersion {
	return &db.ModelVersion{
		Version:   version,
		ModelDefn: model,
		Replicas:  make(map[int32]*db.ReplicaStatus, 0),
		State: &db.ModelStatus{
			State:        db.ModelState_ModelStateUnknown,
			ModelGwState: db.ModelState_ModelCreate,
		},
	}
}

func cleanCapabilities(capabilities []string) []string {
	var cleaned []string
	for _, capability := range capabilities {
		cleaned = append(cleaned, strings.TrimSpace(capability))
	}
	return cleaned
}

func toUniqueModels(loadedModels []*db.ModelVersionID) map[string]bool {
	uniqueModels := make(map[string]bool)
	for _, key := range loadedModels {
		uniqueModels[key.Name] = true
	}
	return uniqueModels
}
