package server

import (
	"fmt"
	"sync"

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
	muReservedMemory  sync.RWMutex
	muLoadedModels    sync.RWMutex
	muDrainingState   sync.RWMutex
	inferenceSvc      string
	inferenceHttpPort int32
	inferenceGrpcPort int32
	serverName        string
	replicaIdx        int
	server            *Server
	capabilities      []string
	memory            uint64
	availableMemory   uint64
	// precomputed values to speed up ops on scheduler
	loadedModels map[ModelVersionID]bool
	// for marking models that are in process of load requested or loading on this server (to speed up ops)
	loadingModels        map[ModelVersionID]bool
	overCommitPercentage uint32
	// holding reserved memory on server replica while loading models, internal to scheduler
	reservedMemory uint64
	// precomputed values to speed up ops on scheduler
	uniqueLoadedModels map[string]bool
	isDraining         bool
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
