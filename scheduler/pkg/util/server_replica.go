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
