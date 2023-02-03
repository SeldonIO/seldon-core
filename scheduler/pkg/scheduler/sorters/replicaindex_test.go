/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sorters

import (
	"sort"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

func TestReplicaIndexSorter(t *testing.T) {
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
			name: "OrderByIndex",
			replicas: []*CandidateReplica{
				{Model: model, Replica: store.NewServerReplica("", 8080, 5001, 20, store.NewServer("dummy", true), []string{}, 100, 200, map[store.ModelVersionID]bool{}, 100)},
				{Model: model, Replica: store.NewServerReplica("", 8080, 5001, 10, store.NewServer("dummy", true), []string{}, 100, 100, map[store.ModelVersionID]bool{}, 100)},
				{Model: model, Replica: store.NewServerReplica("", 8080, 5001, 30, store.NewServer("dummy", true), []string{}, 100, 150, map[store.ModelVersionID]bool{}, 100)},
			},
			ordering: []int{10, 20, 30},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sorter := ReplicaIndexSorter{}
			sort.SliceStable(test.replicas, func(i, j int) bool { return sorter.IsLess(test.replicas[i], test.replicas[j]) })
			for idx, expected := range test.ordering {
				g.Expect(test.replicas[idx].Replica.GetReplicaIdx()).To(Equal(expected))
			}
		})
	}
}
