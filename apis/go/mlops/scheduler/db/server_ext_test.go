/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package db

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestServer_AddReplica(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name        string
		server      *Server
		replicaID   int32
		replica     *ServerReplica
		expectedLen int
	}{
		{
			name:        "add replica to empty server",
			server:      &Server{},
			replicaID:   0,
			replica:     &ServerReplica{ServerName: "test-server"},
			expectedLen: 1,
		},
		{
			name: "add replica to server with existing replicas",
			server: &Server{
				Replicas: map[int32]*ServerReplica{
					0: {ServerName: "server1"},
				},
			},
			replicaID:   1,
			replica:     &ServerReplica{ServerName: "server2"},
			expectedLen: 2,
		},
		{
			name: "overwrite existing replica",
			server: &Server{
				Replicas: map[int32]*ServerReplica{
					0: {ServerName: "old-server"},
				},
			},
			replicaID:   0,
			replica:     &ServerReplica{ServerName: "new-server"},
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.server.AddReplica(tt.replicaID, tt.replica)
			g.Expect(tt.server.Replicas).To(HaveLen(tt.expectedLen))
			g.Expect(tt.server.Replicas[tt.replicaID]).To(Equal(tt.replica))
		})
	}
}

func TestServerReplica_GetLoadedOrLoadingModelVersions(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name         string
		replica      *ServerReplica
		expectedLen  int
		expectedIDs  []*ModelVersionID
	}{
		{
			name: "both loaded and loading models",
			replica: &ServerReplica{
				LoadedModels: []*ModelVersionID{
					{Name: "model1", Version: 1},
					{Name: "model2", Version: 1},
				},
				LoadingModels: []*ModelVersionID{
					{Name: "model3", Version: 1},
				},
			},
			expectedLen: 3,
			expectedIDs: []*ModelVersionID{
				{Name: "model1", Version: 1},
				{Name: "model2", Version: 1},
				{Name: "model3", Version: 1},
			},
		},
		{
			name: "only loaded models",
			replica: &ServerReplica{
				LoadedModels: []*ModelVersionID{
					{Name: "model1", Version: 1},
				},
				LoadingModels: []*ModelVersionID{},
			},
			expectedLen: 1,
			expectedIDs: []*ModelVersionID{
				{Name: "model1", Version: 1},
			},
		},
		{
			name: "only loading models",
			replica: &ServerReplica{
				LoadedModels: []*ModelVersionID{},
				LoadingModels: []*ModelVersionID{
					{Name: "model1", Version: 1},
				},
			},
			expectedLen: 1,
			expectedIDs: []*ModelVersionID{
				{Name: "model1", Version: 1},
			},
		},
		{
			name: "no models",
			replica: &ServerReplica{
				LoadedModels:  []*ModelVersionID{},
				LoadingModels: []*ModelVersionID{},
			},
			expectedLen: 0,
			expectedIDs: []*ModelVersionID{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.replica.GetLoadedOrLoadingModelVersions()
			g.Expect(result).To(HaveLen(tt.expectedLen))
			for i, expected := range tt.expectedIDs {
				g.Expect(result[i].Name).To(Equal(expected.Name))
				g.Expect(result[i].Version).To(Equal(expected.Version))
			}
		})
	}
}

func TestServerReplica_UpdateReservedMemory(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name             string
		replica          *ServerReplica
		memBytes         uint64
		isAdd            bool
		expectedReserved uint64
	}{
		{
			name: "add memory",
			replica: &ServerReplica{
				ReservedMemory: 1000,
			},
			memBytes:         500,
			isAdd:            true,
			expectedReserved: 1500,
		},
		{
			name: "subtract memory",
			replica: &ServerReplica{
				ReservedMemory: 1000,
			},
			memBytes:         300,
			isAdd:            false,
			expectedReserved: 700,
		},
		{
			name: "subtract more than reserved - clamp to 0",
			replica: &ServerReplica{
				ReservedMemory: 500,
			},
			memBytes:         1000,
			isAdd:            false,
			expectedReserved: 0,
		},
		{
			name: "add to zero",
			replica: &ServerReplica{
				ReservedMemory: 0,
			},
			memBytes:         500,
			isAdd:            true,
			expectedReserved: 500,
		},
		{
			name: "subtract exact amount",
			replica: &ServerReplica{
				ReservedMemory: 500,
			},
			memBytes:         500,
			isAdd:            false,
			expectedReserved: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.replica.UpdateReservedMemory(tt.memBytes, tt.isAdd)
			g.Expect(tt.replica.ReservedMemory).To(Equal(tt.expectedReserved))
		})
	}
}

func TestServerReplica_GetNumLoadedModels(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name     string
		replica  *ServerReplica
		expected int
	}{
		{
			name: "multiple unique models",
			replica: &ServerReplica{
				UniqueLoadedModels: map[string]bool{
					"model1": true,
					"model2": true,
					"model3": true,
				},
			},
			expected: 3,
		},
		{
			name: "single model",
			replica: &ServerReplica{
				UniqueLoadedModels: map[string]bool{
					"model1": true,
				},
			},
			expected: 1,
		},
		{
			name: "no models",
			replica: &ServerReplica{
				UniqueLoadedModels: map[string]bool{},
			},
			expected: 0,
		},
		{
			name: "nil map",
			replica: &ServerReplica{
				UniqueLoadedModels: nil,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.replica.GetNumLoadedModels()
			g.Expect(result).To(Equal(tt.expected))
		})
	}
}

func TestServerReplica_AddModelVersion(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name                   string
		replica                *ServerReplica
		modelName              string
		modelVersion           uint32
		replicaState           ModelReplicaState
		expectedLoadingLen     int
		expectedLoadedLen      int
		expectedUniqueModels   map[string]bool
	}{
		{
			name:                 "add loading model to empty replica",
			replica:              &ServerReplica{},
			modelName:            "model1",
			modelVersion:         1,
			replicaState:         ModelReplicaState_Loading,
			expectedLoadingLen:   1,
			expectedLoadedLen:    0,
			expectedUniqueModels: nil,
		},
		{
			name: "add loaded model moves from loading to loaded",
			replica: &ServerReplica{
				LoadingModels: []*ModelVersionID{
					{Name: "model1", Version: 1},
				},
			},
			modelName:          "model1",
			modelVersion:       1,
			replicaState:       ModelReplicaState_Loaded,
			expectedLoadingLen: 0,
			expectedLoadedLen:  1,
			expectedUniqueModels: map[string]bool{
				"model1": true,
			},
		},
		{
			name: "add loaded model to existing loaded models",
			replica: &ServerReplica{
				LoadedModels: []*ModelVersionID{
					{Name: "model1", Version: 1},
				},
				UniqueLoadedModels: map[string]bool{
					"model1": true,
				},
			},
			modelName:          "model2",
			modelVersion:       1,
			replicaState:       ModelReplicaState_Loaded,
			expectedLoadingLen: 0,
			expectedLoadedLen:  2,
			expectedUniqueModels: map[string]bool{
				"model1": true,
				"model2": true,
			},
		},
		{
			name: "add multiple versions of same model",
			replica: &ServerReplica{
				LoadedModels: []*ModelVersionID{
					{Name: "model1", Version: 1},
				},
				UniqueLoadedModels: map[string]bool{
					"model1": true,
				},
			},
			modelName:          "model1",
			modelVersion:       2,
			replicaState:       ModelReplicaState_Loaded,
			expectedLoadingLen: 0,
			expectedLoadedLen:  2,
			expectedUniqueModels: map[string]bool{
				"model1": true,
			},
		},
		{
			name: "add duplicate loading model - should not duplicate",
			replica: &ServerReplica{
				LoadingModels: []*ModelVersionID{
					{Name: "model1", Version: 1},
				},
			},
			modelName:            "model1",
			modelVersion:         1,
			replicaState:         ModelReplicaState_Loading,
			expectedLoadingLen:   1,
			expectedLoadedLen:    0,
			expectedUniqueModels: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.replica.AddModelVersion(tt.modelName, tt.modelVersion, tt.replicaState)
			g.Expect(tt.replica.LoadingModels).To(HaveLen(tt.expectedLoadingLen))
			g.Expect(tt.replica.LoadedModels).To(HaveLen(tt.expectedLoadedLen))
			if tt.expectedUniqueModels != nil {
				g.Expect(tt.replica.UniqueLoadedModels).To(Equal(tt.expectedUniqueModels))
			}
		})
	}
}

func TestServerReplica_DeleteModelVersion(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name                 string
		replica              *ServerReplica
		modelName            string
		modelVersion         uint32
		expectedLoadingLen   int
		expectedLoadedLen    int
		expectedUniqueModels map[string]bool
	}{
		{
			name: "delete loading model",
			replica: &ServerReplica{
				LoadingModels: []*ModelVersionID{
					{Name: "model1", Version: 1},
					{Name: "model2", Version: 1},
				},
			},
			modelName:            "model1",
			modelVersion:         1,
			expectedLoadingLen:   1,
			expectedLoadedLen:    0,
			expectedUniqueModels: nil,
		},
		{
			name: "delete loaded model",
			replica: &ServerReplica{
				LoadedModels: []*ModelVersionID{
					{Name: "model1", Version: 1},
					{Name: "model2", Version: 1},
				},
				UniqueLoadedModels: map[string]bool{
					"model1": true,
					"model2": true,
				},
			},
			modelName:          "model1",
			modelVersion:       1,
			expectedLoadingLen: 0,
			expectedLoadedLen:  1,
			expectedUniqueModels: map[string]bool{
				"model2": true,
			},
		},
		{
			name: "delete last version of model removes from unique",
			replica: &ServerReplica{
				LoadedModels: []*ModelVersionID{
					{Name: "model1", Version: 1},
				},
				UniqueLoadedModels: map[string]bool{
					"model1": true,
				},
			},
			modelName:            "model1",
			modelVersion:         1,
			expectedLoadingLen:   0,
			expectedLoadedLen:    0,
			expectedUniqueModels: map[string]bool{},
		},
		{
			name: "delete one version keeps model in unique if other versions exist",
			replica: &ServerReplica{
				LoadedModels: []*ModelVersionID{
					{Name: "model1", Version: 1},
					{Name: "model1", Version: 2},
				},
				UniqueLoadedModels: map[string]bool{
					"model1": true,
				},
			},
			modelName:          "model1",
			modelVersion:       1,
			expectedLoadingLen: 0,
			expectedLoadedLen:  1,
			expectedUniqueModels: map[string]bool{
				"model1": true,
			},
		},
		{
			name: "delete from both loading and loaded",
			replica: &ServerReplica{
				LoadingModels: []*ModelVersionID{
					{Name: "model1", Version: 1},
				},
				LoadedModels: []*ModelVersionID{
					{Name: "model1", Version: 1},
				},
				UniqueLoadedModels: map[string]bool{
					"model1": true,
				},
			},
			modelName:            "model1",
			modelVersion:         1,
			expectedLoadingLen:   0,
			expectedLoadedLen:    0,
			expectedUniqueModels: map[string]bool{},
		},
		{
			name: "delete non-existent model",
			replica: &ServerReplica{
				LoadedModels: []*ModelVersionID{
					{Name: "model1", Version: 1},
				},
				UniqueLoadedModels: map[string]bool{
					"model1": true,
				},
			},
			modelName:          "model2",
			modelVersion:       1,
			expectedLoadingLen: 0,
			expectedLoadedLen:  1,
			expectedUniqueModels: map[string]bool{
				"model1": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.replica.DeleteModelVersion(tt.modelName, tt.modelVersion)
			g.Expect(tt.replica.LoadingModels).To(HaveLen(tt.expectedLoadingLen))
			g.Expect(tt.replica.LoadedModels).To(HaveLen(tt.expectedLoadedLen))
			if tt.expectedUniqueModels != nil {
				g.Expect(tt.replica.UniqueLoadedModels).To(Equal(tt.expectedUniqueModels))
			}
		})
	}
}

func TestServerReplica_AddModelVersion_Integration(t *testing.T) {
	g := NewWithT(t)

	t.Run("full lifecycle: loading -> loaded -> delete", func(t *testing.T) {
		replica := &ServerReplica{}

		// Add model as loading
		replica.AddModelVersion("model1", 1, ModelReplicaState_Loading)
		g.Expect(replica.LoadingModels).To(HaveLen(1))
		g.Expect(replica.LoadedModels).To(HaveLen(0))
		g.Expect(replica.GetNumLoadedModels()).To(Equal(0))

		// Move to loaded
		replica.AddModelVersion("model1", 1, ModelReplicaState_Loaded)
		g.Expect(replica.LoadingModels).To(HaveLen(0))
		g.Expect(replica.LoadedModels).To(HaveLen(1))
		g.Expect(replica.GetNumLoadedModels()).To(Equal(1))

		// Verify it's in loaded models list
		allModels := replica.GetLoadedOrLoadingModelVersions()
		g.Expect(allModels).To(HaveLen(1))
		g.Expect(allModels[0].Name).To(Equal("model1"))
		g.Expect(allModels[0].Version).To(Equal(uint32(1)))

		// Delete model
		replica.DeleteModelVersion("model1", 1)
		g.Expect(replica.LoadedModels).To(HaveLen(0))
		g.Expect(replica.GetNumLoadedModels()).To(Equal(0))
	})

	t.Run("multiple models", func(t *testing.T) {
		replica := &ServerReplica{}

		// Add three models
		replica.AddModelVersion("model1", 1, ModelReplicaState_Loaded)
		replica.AddModelVersion("model2", 1, ModelReplicaState_Loaded)
		replica.AddModelVersion("model3", 1, ModelReplicaState_Loading)

		g.Expect(replica.LoadedModels).To(HaveLen(2))
		g.Expect(replica.LoadingModels).To(HaveLen(1))
		g.Expect(replica.GetNumLoadedModels()).To(Equal(2))

		allModels := replica.GetLoadedOrLoadingModelVersions()
		g.Expect(allModels).To(HaveLen(3))

		// Delete one loaded model
		replica.DeleteModelVersion("model1", 1)
		g.Expect(replica.GetNumLoadedModels()).To(Equal(1))
		g.Expect(replica.UniqueLoadedModels).To(HaveKey("model2"))
		g.Expect(replica.UniqueLoadedModels).NotTo(HaveKey("model1"))
	})
}

func TestServerReplica_UpdateReservedMemory_Integration(t *testing.T) {
	g := NewWithT(t)

	t.Run("reserve and release memory", func(t *testing.T) {
		replica := &ServerReplica{
			Memory:          10000,
			ReservedMemory:  0,
			AvailableMemory: 10000,
		}

		// Reserve memory for model 1
		replica.UpdateReservedMemory(3000, true)
		g.Expect(replica.ReservedMemory).To(Equal(uint64(3000)))

		// Reserve memory for model 2
		replica.UpdateReservedMemory(2000, true)
		g.Expect(replica.ReservedMemory).To(Equal(uint64(5000)))

		// Release model 1
		replica.UpdateReservedMemory(3000, false)
		g.Expect(replica.ReservedMemory).To(Equal(uint64(2000)))

		// Release model 2
		replica.UpdateReservedMemory(2000, false)
		g.Expect(replica.ReservedMemory).To(Equal(uint64(0)))
	})
}