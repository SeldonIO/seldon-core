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
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
)

func TestSharingFilter(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		model    *store.ModelVersion
		server   *store.ServerSnapshot
		expected bool
	}
	serverName := "server1"
	modelExplicitServer := store.NewModelVersion(
		&pb.Model{ModelSpec: &pb.ModelSpec{Server: &serverName}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
		1,
		serverName,
		map[int]store.ReplicaStatus{3: {State: store.Loading}},
		false,
		store.ModelProgressing)
	modelSharedServer := store.NewModelVersion(
		&pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
		1,
		serverName,
		map[int]store.ReplicaStatus{3: {State: store.Loading}},
		false,
		store.ModelProgressing)
	tests := []test{
		{name: "ModelAndServerMatchNotShared", model: modelExplicitServer, server: &store.ServerSnapshot{Name: serverName, Shared: false}, expected: true},
		{name: "ModelAndServerMatchShared", model: modelExplicitServer, server: &store.ServerSnapshot{Name: serverName, Shared: true}, expected: true},
		{name: "ModelAndServerDontMatch", model: modelExplicitServer, server: &store.ServerSnapshot{Name: "foo", Shared: true}, expected: false},
		{name: "SharedModelAnyServer", model: modelSharedServer, server: &store.ServerSnapshot{Name: "foo", Shared: true}, expected: true},
		{name: "SharedModelNotSharedServer", model: modelSharedServer, server: &store.ServerSnapshot{Name: "foo", Shared: false}, expected: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filter := SharingServerFilter{}
			ok := filter.Filter(test.model, test.server)
			g.Expect(ok).To(Equal(test.expected))
		})
	}
}
