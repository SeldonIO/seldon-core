/*
Copyright 2023 Seldon Technologies Ltd.

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

func TestServerReplicasFilter(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		model    *store.ModelVersion
		server   *store.ServerSnapshot
		expected bool
	}
	serverName := "server1"
	model := store.NewModelVersion(
		&pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
		1,
		serverName,
		map[int]store.ReplicaStatus{3: {State: store.Loading}},
		false,
		store.ModelProgressing)
	tests := []test{
		{
			name:  "No Replicas",
			model: model,
			server: &store.ServerSnapshot{Name: serverName,
				Shared:           true,
				ExpectedReplicas: 0,
			},
			expected: false,
		},
		{
			name:  "Replicas",
			model: model,
			server: &store.ServerSnapshot{Name: serverName,
				Shared:           true,
				ExpectedReplicas: 0,
				Replicas: map[int]*store.ServerReplica{
					0: &store.ServerReplica{},
				},
			},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filter := ServerReplicaFilter{}
			ok := filter.Filter(test.model, test.server)
			g.Expect(ok).To(Equal(test.expected))
		})
	}
}
