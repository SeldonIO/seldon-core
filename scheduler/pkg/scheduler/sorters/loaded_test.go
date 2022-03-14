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
		1,
		"server1",
		map[int]store.ReplicaStatus{3: {State: store.Loading}},
		false,
		store.ModelProgressing)
	modelServer2 := store.NewModelVersion(
		nil,
		1,
		"server2",
		map[int]store.ReplicaStatus{3: {State: store.Loading}},
		false,
		store.ModelProgressing)
	server := store.NewServer("server1", true)

	tests := []test{
		{
			name: "OneLoadedModel",
			replicas: []*CandidateReplica{
				{Model: model, Server: &store.ServerSnapshot{Name: "server1"}, Replica: store.NewServerReplica("", 8080, 5001, 2, server, []string{}, 100, 100, map[store.ModelVersionID]bool{}, 100)},
				{Model: model, Server: &store.ServerSnapshot{Name: "server1"}, Replica: store.NewServerReplica("", 8080, 5001, 1, server, []string{}, 100, 100, map[store.ModelVersionID]bool{}, 100)},
				{Model: model, Server: &store.ServerSnapshot{Name: "server1"}, Replica: store.NewServerReplica("", 8080, 5001, 3, server, []string{}, 100, 100, map[store.ModelVersionID]bool{}, 100)},
			},
			ordering: []int{3, 2, 1},
		},
		{
			name: "LoadedDifferentServer",
			replicas: []*CandidateReplica{
				{Model: modelServer2, Server: &store.ServerSnapshot{Name: "server1"}, Replica: store.NewServerReplica("", 8080, 5001, 2, server, []string{}, 100, 100, map[store.ModelVersionID]bool{}, 100)},
				{Model: modelServer2, Server: &store.ServerSnapshot{Name: "server1"}, Replica: store.NewServerReplica("", 8080, 5001, 1, server, []string{}, 100, 100, map[store.ModelVersionID]bool{}, 100)},
				{Model: modelServer2, Server: &store.ServerSnapshot{Name: "server1"}, Replica: store.NewServerReplica("", 8080, 5001, 3, server, []string{}, 100, 100, map[store.ModelVersionID]bool{}, 100)},
			},
			ordering: []int{2, 1, 3},
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
