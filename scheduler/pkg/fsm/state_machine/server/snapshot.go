package server

import (
	"fmt"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

type Snapshot struct {
	Name             string
	Replicas         map[int]*Replica
	Shared           bool
	ExpectedReplicas int
	MinReplicas      int
	MaxReplicas      int
	KubernetesMeta   *pb.KubernetesMeta
	Stats            *Stats
}

type Replica struct {
	InferenceSvc      string
	InferenceHttpPort int32
	InferenceGrpcPort int32
	ServerName        string
	ReplicaIdx        int
	Server            *Server
	Capabilities      []string
	Memory            uint64
	AvailableMemory   uint64
	// precomputed values to speed up ops on scheduler
	LoadedModels map[ModelVersionID]bool
	// for marking models that are in process of load requested or loading on this server (to speed up ops)
	LoadingModels        map[ModelVersionID]bool
	OverCommitPercentage uint32
	// holding reserved memory on server replica while loading models, internal to scheduler
	ReservedMemory uint64
	// precomputed values to speed up ops on scheduler
	UniqueLoadedModels map[string]bool
	IsDraining         bool
}

type Stats struct {
	NumEmptyReplicas          uint32
	MaxNumReplicaHostedModels uint32
}

type ModelVersionID struct {
	Name    string
	Version uint32
}

func (mv *ModelVersionID) String() string {
	return fmt.Sprintf("%s:%d", mv.Name, mv.Version)
}

type Server struct {
	name             string
	replicas         map[int]*Replica
	shared           bool
	expectedReplicas int
	minReplicas      int
	maxReplicas      int
	kubernetesMeta   *pb.KubernetesMeta
}

func (ss *Snapshot) GetCapabilities() []string {
	for _, replica := range ss.Replicas {
		return replica.Capabilities
	}

	return []string{}
}
