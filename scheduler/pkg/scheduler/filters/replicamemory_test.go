package filters

import (
	"testing"

	. "github.com/onsi/gomega"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
)

func getTestModelWithMemory(requiredmemory *uint64) *store.ModelVersion {
	return store.NewModelVersion(
		&pb.ModelDetails{
			Name:         "model1",
			Requirements: []string{},
			MemoryBytes:  requiredmemory,
		},
		"server",
		map[int]store.ReplicaStatus{3: {State: store.Loading}},
		false,
		store.ModelProgressing)
}

func getTestServerReplicaWithMemory(availableMemory uint64) *store.ServerReplica {
	return store.NewServerReplica("svc", 8080, 5001, 1, nil, []string{}, availableMemory, availableMemory, nil, true)
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
		{name: "EnoughMemory", model: getTestModelWithMemory(&memory), server: getTestServerReplicaWithMemory(100), expected: true},
		{name: "NoMemorySpecified", model: getTestModelWithMemory(nil), server: getTestServerReplicaWithMemory(200), expected: true},
		{name: "NotEnoughMemory", model: getTestModelWithMemory(&memory), server: getTestServerReplicaWithMemory(50), expected: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filter := AvailableMemoryFilter{}
			ok := filter.Filter(test.model, test.server)
			g.Expect(ok).To(Equal(test.expected))
		})
	}
}
