package filters

import (
	"testing"

	. "github.com/onsi/gomega"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
)

func getTestModelWithMemory(requiredmemory *uint64, serverName string, replicaId int) *store.ModelVersion {

	replicas := map[int]store.ReplicaStatus{}
	if replicaId >= 0 {
		replicas[replicaId] = store.ReplicaStatus{State: store.Loading}
	}
	return store.NewModelVersion(
		&pb.Model{ModelSpec: &pb.ModelSpec{MemoryBytes: requiredmemory}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
		1,
		serverName,
		replicas,
		false,
		store.ModelProgressing)
}

func getTestServerReplicaWithMemory(availableMemory uint64, serverName string, replicaId int) *store.ServerReplica {
	return store.NewServerReplica("svc", 8080, 5001, replicaId, store.NewServer(serverName, true), []string{}, availableMemory, availableMemory, nil, 100)
}

func TestReplicaMemoryFilter(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		model    *store.ModelVersion
		server   *store.ServerReplica
		expected bool
	}

	memory := uint64(100)
	tests := []test{
		{name: "EnoughMemory", model: getTestModelWithMemory(&memory, "", -1), server: getTestServerReplicaWithMemory(100, "server1", 0), expected: true},
		{name: "NoMemorySpecified", model: getTestModelWithMemory(nil, "", -1), server: getTestServerReplicaWithMemory(200, "server1", 0), expected: true},
		{name: "NotEnoughMemory", model: getTestModelWithMemory(&memory, "", -1), server: getTestServerReplicaWithMemory(50, "server1", 0), expected: false},
		{name: "ModelAlreadyLoaded", model: getTestModelWithMemory(&memory, "server1", 0), server: getTestServerReplicaWithMemory(0, "server1", 0), expected: true}, // note not enough memory on server replica
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filter := AvailableMemoryReplicaFilter{}
			ok := filter.Filter(test.model, test.server)
			g.Expect(ok).To(Equal(test.expected))
		})
	}
}
