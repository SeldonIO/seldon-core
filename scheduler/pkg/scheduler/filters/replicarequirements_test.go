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

package filters

import (
	"testing"

	. "github.com/onsi/gomega"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

func TestServerRequirementFilter(t *testing.T) {
	g := NewGomegaWithT(t)

	makeModel := func(requirements []string) *store.ModelVersion {
		return store.NewModelVersion(
			&pb.Model{ModelSpec: &pb.ModelSpec{Requirements: requirements}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
			1,
			"server",
			map[int]store.ReplicaStatus{3: {State: store.Loading}},
			false,
			store.ModelProgressing)
	}

	makeServerReplica := func(server *store.Server, capabilities []string) *store.ServerReplica {
		return store.NewServerReplica("svc", 8080, 5001, 1, store.NewServer("server", true), capabilities, 100, 100, 0, nil, 100)
	}

	makeServer := func(replicas int, capabilities []string) *store.ServerSnapshot {
		server := store.NewServer("server", true)
		snapshot := server.CreateSnapshot(false, false)

		for i := 0; i < replicas; i++ {
			replica := makeServerReplica(server, capabilities)
			snapshot.Replicas[i] = replica
		}

		return snapshot
	}

	type test struct {
		name     string
		model    *store.ModelVersion
		server   *store.ServerSnapshot
		expected bool
	}

	tests := []test{
		{name: "Match", model: makeModel([]string{"sklearn"}), server: makeServer(1, []string{"sklearn"}), expected: true},
		{name: "Mismatch", model: makeModel([]string{"sklearn"}), server: makeServer(1, []string{"xgboost"}), expected: false},
		{name: "PartialMatch", model: makeModel([]string{"sklearn", "xgboost"}), server: makeServer(1, []string{"xgboost"}), expected: false},
		{name: "MultiMatch", model: makeModel([]string{"sklearn", "xgboost"}), server: makeServer(1, []string{"xgboost", "sklearn", "tensorflow"}), expected: true},
		{name: "Duplicates", model: makeModel([]string{"sklearn", "xgboost", "sklearn"}), server: makeServer(1, []string{"xgboost", "sklearn", "tensorflow"}), expected: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filter := ServerRequirementFilter{}
			ok := filter.Filter(test.model, test.server)
			g.Expect(ok).To(Equal(test.expected))
		})
	}
}
