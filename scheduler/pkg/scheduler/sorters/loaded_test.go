package sorters

import (
	"sort"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
)

func TestModelAlreadyLoadedSort(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		replicas []*CandidateReplica
		ordering []int
	}

	model := store.NewModelVersion(
		nil,
		"server1",
		map[int]store.ModelReplicaState{3: store.Loading},
		false,
		store.ModelProgressing)
	tests := []test{
		{
			name: "OneLoadedModel",
			replicas: []*CandidateReplica{
				{Model: model, Replica: store.NewServerReplica("", 8080, 5001, 2, nil, []string{}, 100, 100, map[string]bool{}, true)},
				{Model: model, Replica: store.NewServerReplica("", 8080, 5001, 1, nil, []string{}, 100, 100, map[string]bool{}, true)},
				{Model: model, Replica: store.NewServerReplica("", 8080, 5001, 3, nil, []string{}, 100, 100, map[string]bool{}, true)},
			},
			ordering: []int{3, 2, 1},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sorter := ModelAlreadyLoadedSorter{}
			sort.SliceStable(test.replicas, func(i, j int) bool { return sorter.IsLess(test.replicas[i], test.replicas[j]) })
			for idx, expected := range test.ordering {
				g.Expect(test.replicas[idx].Replica.GetReplicaIdx()).To(Equal(expected))
			}
		})
	}
}
