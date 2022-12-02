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

package store

import (
	"testing"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	. "github.com/onsi/gomega"
)

func TestReplicaStateToString(t *testing.T) {
	for _, state := range replicaStates {
		_ = state.String()
	}
}

func TestCleanCapabilities(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		in       []string
		expected []string
	}

	tests := []test{
		{
			name:     "misc",
			in:       []string{"mlserver", " foo ", " bar", "bar   "},
			expected: []string{"mlserver", "foo", "bar", "bar"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := cleanCapabilities(test.in)
			g.Expect(out).To(Equal(test.expected))
		})
	}
}

func TestCreateSnapshot(t *testing.T) {
	g := NewGomegaWithT(t)

	server := &Server{
		name: "test",
		replicas: map[int]*ServerReplica{
			0: {
				inferenceSvc: "svc",
				loadedModels: map[ModelVersionID]bool{
					{Name: "model1", Version: 1}: true,
					{Name: "model2", Version: 2}: true,
				},
			},
		},
		kubernetesMeta: &pb.KubernetesMeta{Namespace: "default"},
	}

	snapshot := server.CreateSnapshot(false, true)

	server.replicas[1] = &ServerReplica{
		inferenceSvc: "svc",
		loadedModels: map[ModelVersionID]bool{
			{Name: "model3", Version: 1}: true,
			{Name: "model4", Version: 2}: true,
		},
	}
	server.name = "foo"
	server.kubernetesMeta.Namespace = "test"

	g.Expect(snapshot.Name).To(Equal("test"))
	g.Expect(len(snapshot.Replicas)).To(Equal(1))
	g.Expect(snapshot.KubernetesMeta.Namespace).To(Equal("default"))

}
