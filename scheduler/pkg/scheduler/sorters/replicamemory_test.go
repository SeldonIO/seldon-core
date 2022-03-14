package sorters

import (
	"sort"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
)

func TestReplicaMemorySort(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		replicas []*CandidateReplica
		ordering []int
	}

	model := store.NewModelVersion(
		nil,
		1,
		"server1",
		map[int]store.ReplicaStatus{3: {State: store.Loading}},
		false,
		store.ModelProgressing)
	tests := []test{
		{
			name: "ThreeReplicasDifferentMemory",
			replicas: []*CandidateReplica{
				{Model: model, Replica: store.NewServerReplica("", 8080, 5001, 1, nil, []string{}, 100, 100, map[store.ModelVersionID]bool{}, 100)},
				{Model: model, Replica: store.NewServerReplica("", 8080, 5001, 2, nil, []string{}, 100, 200, map[store.ModelVersionID]bool{}, 100)},
				{Model: model, Replica: store.NewServerReplica("", 8080, 5001, 3, nil, []string{}, 100, 150, map[store.ModelVersionID]bool{}, 100)},
			},
			ordering: []int{2, 3, 1},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sorter := AvailableMemorySorter{}
			sort.SliceStable(test.replicas, func(i, j int) bool { return sorter.IsLess(test.replicas[i], test.replicas[j]) })
			for idx, expected := range test.ordering {
				g.Expect(test.replicas[idx].Replica.GetReplicaIdx()).To(Equal(expected))
			}
		})
	}
}
