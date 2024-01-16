/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
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
			&pb.Model{
				ModelSpec: &pb.ModelSpec{
					Requirements: requirements,
				},
				DeploymentSpec: &pb.DeploymentSpec{
					Replicas: 1,
				},
			},
			1,
			"server",
			map[int]store.ReplicaStatus{
				3: {State: store.Loading},
			},
			false,
			store.ModelProgressing,
		)
	}

	makeServerReplica := func(server *store.Server, capabilities []string) *store.ServerReplica {
		return store.NewServerReplica("svc", 8080, 5001, 1, store.NewServer("server", true), capabilities, 100, 100, 0, nil, 100)
	}

	makeServer := func(replicas int, capabilities []string, startIdx int) *store.ServerSnapshot {
		server := store.NewServer("server", true)
		snapshot := server.CreateSnapshot(false, false)

		for i := 0; i < replicas; i++ {
			replica := makeServerReplica(server, capabilities)
			snapshot.Replicas[i+startIdx] = replica
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
		{
			name:     "NoReplicas",
			model:    makeModel([]string{"sklearn"}),
			server:   makeServer(0, []string{}, 0),
			expected: false,
		},
		{
			name:     "Match",
			model:    makeModel([]string{"sklearn"}),
			server:   makeServer(1, []string{"sklearn"}, 0),
			expected: true,
		},
		{
			name:     "MatchNonZeroReplicaIdx",
			model:    makeModel([]string{"sklearn"}),
			server:   makeServer(1, []string{"sklearn"}, 10),
			expected: true,
		},
		{
			name:     "Mismatch",
			model:    makeModel([]string{"sklearn"}),
			server:   makeServer(1, []string{"xgboost"}, 0),
			expected: false,
		},
		{
			name:     "PartialMatch",
			model:    makeModel([]string{"sklearn", "xgboost"}),
			server:   makeServer(1, []string{"xgboost"}, 0),
			expected: false,
		},
		{
			name:     "MultiMatch",
			model:    makeModel([]string{"sklearn", "xgboost"}),
			server:   makeServer(1, []string{"xgboost", "sklearn", "tensorflow"}, 0),
			expected: true,
		},
		{
			name:     "Duplicates",
			model:    makeModel([]string{"sklearn", "xgboost", "sklearn"}),
			server:   makeServer(1, []string{"xgboost", "sklearn", "tensorflow"}, 0),
			expected: true,
		},
		{
			name:     "MultipleReplicasMatch",
			model:    makeModel([]string{"sklearn"}),
			server:   makeServer(2, []string{"xgboost", "sklearn", "tensorflow"}, 0),
			expected: true,
		},
		{
			name:     "MultipleReplicasMismatch",
			model:    makeModel([]string{"sklearn"}),
			server:   makeServer(2, []string{"xgboost"}, 0),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filter := ServerRequirementFilter{}
			ok := filter.Filter(test.model, test.server)
			g.Expect(ok).To(Equal(test.expected))
		})
	}
}
