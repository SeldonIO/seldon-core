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
				loadingModels: map[ModelVersionID]bool{
					{Name: "model10", Version: 1}: true,
					{Name: "model20", Version: 2}: true,
				},
			},
		},
		kubernetesMeta: &pb.KubernetesMeta{Namespace: "default"},
	}

	snapshot := server.CreateSnapshot(false, true)

	g.Expect(snapshot.Replicas[0].loadedModels).To(Equal(
		map[ModelVersionID]bool{
			{Name: "model1", Version: 1}: true,
			{Name: "model2", Version: 2}: true,
		},
	))

	g.Expect(snapshot.Replicas[0].loadingModels).To(Equal(
		map[ModelVersionID]bool{
			{Name: "model10", Version: 1}: true,
			{Name: "model20", Version: 2}: true,
		},
	))

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

func TestLoadedModel(t *testing.T) {
	g := NewGomegaWithT(t)

	const (
		add int = iota
		remove
	)

	type test struct {
		name                       string
		op                         int
		model                      string
		version                    int
		state                      ModelReplicaState
		loadedModels               map[ModelVersionID]bool
		uniqueLoadedModels         map[string]bool
		loadingModels              map[ModelVersionID]bool
		expectedLoadedModels       map[ModelVersionID]bool
		expectedLoadingModels      map[ModelVersionID]bool
		expectedUniqueLoadedModels map[string]bool
	}

	tests := []test{
		{
			name:         "add loading",
			op:           add,
			model:        "dummy",
			version:      1,
			state:        Loading,
			loadedModels: map[ModelVersionID]bool{},
			loadingModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 1}: true, // we should already have an entry from an earlier load request
			},
			uniqueLoadedModels:         map[string]bool{},
			expectedLoadedModels:       map[ModelVersionID]bool{},
			expectedUniqueLoadedModels: map[string]bool{},
			expectedLoadingModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 1}: true,
			},
		},
		{
			name:         "add loaded",
			op:           add,
			model:        "dummy",
			version:      1,
			state:        Loaded,
			loadedModels: map[ModelVersionID]bool{},
			loadingModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 1}: true, // we should transition from loading to loaded
			},
			uniqueLoadedModels: map[string]bool{},
			expectedLoadedModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 1}: true,
			},
			expectedUniqueLoadedModels: map[string]bool{
				"dummy": true,
			},
			expectedLoadingModels: map[ModelVersionID]bool{},
		},
		{
			name:    "add loading - new version",
			op:      add,
			model:   "dummy",
			version: 2,
			state:   Loading,
			loadedModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 1}: true,
			},
			loadingModels: map[ModelVersionID]bool{},
			uniqueLoadedModels: map[string]bool{
				"dummy": true,
			},
			expectedLoadedModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 1}: true,
			},
			expectedUniqueLoadedModels: map[string]bool{
				"dummy": true,
			},
			expectedLoadingModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 2}: true,
			},
		},
		{
			name:    "add loaded- new version",
			op:      add,
			model:   "dummy",
			version: 2,
			state:   Loaded,
			loadedModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 1}: true,
			},
			loadingModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 2}: true, // we should transition from loading to loaded
			},
			uniqueLoadedModels: map[string]bool{
				"dummy": true,
			},
			expectedLoadedModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 1}: true,
				{Name: "dummy", Version: 2}: true,
			},
			expectedUniqueLoadedModels: map[string]bool{
				"dummy": true,
			},
			expectedLoadingModels: map[ModelVersionID]bool{},
		},
		{
			name:    "remove with early version",
			op:      remove,
			model:   "dummy",
			version: 2,
			loadedModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 1}: true,
				{Name: "dummy", Version: 2}: true,
			},
			loadingModels: map[ModelVersionID]bool{},
			uniqueLoadedModels: map[string]bool{
				"dummy": true,
			},
			expectedLoadedModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 1}: true,
			},
			expectedUniqueLoadedModels: map[string]bool{
				"dummy": true,
			},
			expectedLoadingModels: map[ModelVersionID]bool{},
		},
		{
			name:    "remove",
			op:      remove,
			model:   "dummy",
			version: 1,
			loadedModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 1}: true,
			},
			loadingModels: map[ModelVersionID]bool{},
			uniqueLoadedModels: map[string]bool{
				"dummy": true,
			},
			expectedLoadedModels:       map[ModelVersionID]bool{},
			expectedUniqueLoadedModels: map[string]bool{},
			expectedLoadingModels:      map[ModelVersionID]bool{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := ServerReplica{
				inferenceSvc:       "svc",
				loadedModels:       test.loadedModels,
				loadingModels:      test.loadingModels,
				uniqueLoadedModels: test.uniqueLoadedModels,
			}

			if test.op == add {
				server.addModelVersion(test.model, uint32(test.version), test.state)
			} else {
				server.deleteModelVersion(test.model, uint32(test.version))
			}
			g.Expect(server.loadedModels).To(Equal(test.expectedLoadedModels))
			g.Expect(server.loadingModels).To(Equal(test.expectedLoadingModels))
			g.Expect(server.uniqueLoadedModels).To(Equal(test.expectedUniqueLoadedModels))
		})
	}
}
