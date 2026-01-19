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
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

func TestModelVersion_GetAssignment(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name     string
		mv       *ModelVersion
		expected []int
	}{
		{
			name: "loaded and available replicas",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Loaded},
					1: {State: ModelReplicaState_Available},
					2: {State: ModelReplicaState_Unloaded},
				},
			},
			expected: []int{0, 1},
		},
		{
			name: "prefer non-draining over draining",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Loaded},
					1: {State: ModelReplicaState_Draining},
				},
			},
			expected: []int{0},
		},
		{
			name: "only draining replicas",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Draining},
					1: {State: ModelReplicaState_Draining},
				},
			},
			expected: []int{0, 1},
		},
		{
			name: "loaded unavailable replicas",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_LoadedUnavailable},
					1: {State: ModelReplicaState_Unloaded},
				},
			},
			expected: []int{0},
		},
		{
			name: "no available replicas",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Unloaded},
					1: {State: ModelReplicaState_LoadFailed},
				},
			},
			expected: nil,
		},
		{
			name:     "empty replicas",
			mv:       &ModelVersion{Replicas: map[int32]*ReplicaStatus{}},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mv.GetAssignment()
			g.Expect(result).To(ConsistOf(tt.expected))
		})
	}
}

func TestModelVersion_ReplicaState(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name     string
		mv       *ModelVersion
		expected map[int]*ReplicaStatus
	}{
		{
			name: "multiple replicas",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Loaded},
					1: {State: ModelReplicaState_Available},
				},
			},
			expected: map[int]*ReplicaStatus{
				0: {State: ModelReplicaState_Loaded},
				1: {State: ModelReplicaState_Available},
			},
		},
		{
			name:     "empty replicas",
			mv:       &ModelVersion{Replicas: map[int32]*ReplicaStatus{}},
			expected: map[int]*ReplicaStatus{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mv.ReplicaState()
			g.Expect(result).To(HaveLen(len(tt.expected)))
			for idx, status := range tt.expected {
				g.Expect(result[idx].State).To(Equal(status.State))
			}
		})
	}
}

func TestModelVersion_IsLoadingOrLoaded(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name        string
		mv          *ModelVersion
		server      string
		replicaIdx  int
		expected    bool
	}{
		{
			name: "loading on correct server",
			mv: &ModelVersion{
				Server: "server1",
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Loading},
				},
			},
			server:     "server1",
			replicaIdx: 0,
			expected:   true,
		},
		{
			name: "loaded on correct server",
			mv: &ModelVersion{
				Server: "server1",
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Loaded},
				},
			},
			server:     "server1",
			replicaIdx: 0,
			expected:   true,
		},
		{
			name: "wrong server",
			mv: &ModelVersion{
				Server: "server1",
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Loaded},
				},
			},
			server:     "server2",
			replicaIdx: 0,
			expected:   false,
		},
		{
			name: "wrong replica index",
			mv: &ModelVersion{
				Server: "server1",
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Loaded},
				},
			},
			server:     "server1",
			replicaIdx: 1,
			expected:   false,
		},
		{
			name: "unloaded state",
			mv: &ModelVersion{
				Server: "server1",
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Unloaded},
				},
			},
			server:     "server1",
			replicaIdx: 0,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mv.IsLoadingOrLoaded(tt.server, tt.replicaIdx)
			g.Expect(result).To(Equal(tt.expected))
		})
	}
}

func TestModelVersion_DesiredReplicas(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name     string
		mv       *ModelVersion
		expected int
	}{
		{
			name: "3 replicas",
			mv: &ModelVersion{
				ModelDefn: &pb.Model{
					DeploymentSpec: &pb.DeploymentSpec{
						Replicas: 3,
					},
				},
			},
			expected: 3,
		},
		{
			name: "0 replicas",
			mv: &ModelVersion{
				ModelDefn: &pb.Model{
					DeploymentSpec: &pb.DeploymentSpec{
						Replicas: 0,
					},
				},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mv.DesiredReplicas()
			g.Expect(result).To(Equal(tt.expected))
		})
	}
}

func TestModelVersion_ModelName(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name     string
		mv       *ModelVersion
		expected string
	}{
		{
			name: "valid model name",
			mv: &ModelVersion{
				ModelDefn: &pb.Model{
					Meta: &pb.MetaData{
						Name: "test-model",
					},
				},
			},
			expected: "test-model",
		},
		{
			name: "empty model name",
			mv: &ModelVersion{
				ModelDefn: &pb.Model{
					Meta: &pb.MetaData{
						Name: "",
					},
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mv.ModelName()
			g.Expect(result).To(Equal(tt.expected))
		})
	}
}

func TestModelVersion_IsLoadingOrLoadedOnServer(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name     string
		mv       *ModelVersion
		expected bool
	}{
		{
			name: "has loading replica",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Loading},
				},
			},
			expected: true,
		},
		{
			name: "has loaded replica",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Loaded},
				},
			},
			expected: true,
		},
		{
			name: "has available replica",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Available},
				},
			},
			expected: true,
		},
		{
			name: "has loaded unavailable replica",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_LoadedUnavailable},
				},
			},
			expected: true,
		},
		{
			name: "only unloaded replicas",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Unloaded},
				},
			},
			expected: false,
		},
		{
			name:     "empty replicas",
			mv:       &ModelVersion{Replicas: map[int32]*ReplicaStatus{}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mv.IsLoadingOrLoadedOnServer()
			g.Expect(result).To(Equal(tt.expected))
		})
	}
}

func TestModelVersion_GetReplicaForState(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name     string
		mv       *ModelVersion
		state    ModelReplicaState
		expected []int
	}{
		{
			name: "find loaded replicas",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Loaded},
					1: {State: ModelReplicaState_Available},
					2: {State: ModelReplicaState_Loaded},
				},
			},
			state:    ModelReplicaState_Loaded,
			expected: []int{0, 2},
		},
		{
			name: "no replicas in state",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Loaded},
					1: {State: ModelReplicaState_Available},
				},
			},
			state:    ModelReplicaState_Unloaded,
			expected: nil,
		},
		{
			name:     "empty replicas",
			mv:       &ModelVersion{Replicas: map[int32]*ReplicaStatus{}},
			state:    ModelReplicaState_Loaded,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mv.GetReplicaForState(tt.state)
			g.Expect(result).To(ConsistOf(tt.expected))
		})
	}
}

func TestModelVersion_HasLiveReplicas(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name     string
		mv       *ModelVersion
		expected bool
	}{
		{
			name: "has loaded replica",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Loaded},
				},
			},
			expected: true,
		},
		{
			name: "has available replica",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Available},
				},
			},
			expected: true,
		},
		{
			name: "has draining replica",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Draining},
				},
			},
			expected: true,
		},
		{
			name: "only unloaded replicas",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Unloaded},
				},
			},
			expected: false,
		},
		{
			name:     "empty replicas",
			mv:       &ModelVersion{Replicas: map[int32]*ReplicaStatus{}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mv.HasLiveReplicas()
			g.Expect(result).To(Equal(tt.expected))
		})
	}
}

func TestModelVersion_HasServer(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name     string
		mv       *ModelVersion
		expected bool
	}{
		{
			name:     "has server",
			mv:       &ModelVersion{Server: "server1"},
			expected: true,
		},
		{
			name:     "empty server",
			mv:       &ModelVersion{Server: ""},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mv.HasServer()
			g.Expect(result).To(Equal(tt.expected))
		})
	}
}

func TestModelVersion_GetRequiredMemory(t *testing.T) {
	g := NewWithT(t)

	uint64Ptr := func(v uint64) *uint64 { return &v }

	tests := []struct {
		name     string
		mv       *ModelVersion
		expected uint64
	}{
		{
			name: "basic memory without runtime info",
			mv: &ModelVersion{
				ModelDefn: &pb.Model{
					ModelSpec: &pb.ModelSpec{
						MemoryBytes: uint64Ptr(1000),
					},
				},
			},
			expected: 1000,
		},
		{
			name: "mlserver with parallel workers",
			mv: &ModelVersion{
				ModelDefn: &pb.Model{
					ModelSpec: &pb.ModelSpec{
						MemoryBytes: uint64Ptr(1000),
						ModelRuntimeInfo: &pb.ModelRuntimeInfo{
							ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{
								Mlserver: &pb.MLServerModelSettings{
									ParallelWorkers: 3,
								},
							},
						},
					},
				},
			},
			expected: 3000,
		},
		{
			name: "triton with instance count",
			mv: &ModelVersion{
				ModelDefn: &pb.Model{
					ModelSpec: &pb.ModelSpec{
						MemoryBytes: uint64Ptr(1000),
						ModelRuntimeInfo: &pb.ModelRuntimeInfo{
							ModelRuntimeInfo: &pb.ModelRuntimeInfo_Triton{
								Triton: &pb.TritonModelConfig{
									Cpu: []*pb.TritonCPU{
										{InstanceCount: 2},
									},
								},
							},
						},
					},
				},
			},
			expected: 2000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mv.GetRequiredMemory()
			g.Expect(result).To(Equal(tt.expected))
		})
	}
}

func TestModelVersion_SetReplicaState(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name        string
		mv          *ModelVersion
		replicaIdx  int
		state       ModelReplicaState
		reason      string
		expectedLen int
	}{
		{
			name:        "set state on empty replicas",
			mv:          &ModelVersion{},
			replicaIdx:  0,
			state:       ModelReplicaState_Loaded,
			reason:      "test reason",
			expectedLen: 1,
		},
		{
			name: "update existing replica state",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Loading},
				},
			},
			replicaIdx:  0,
			state:       ModelReplicaState_Loaded,
			reason:      "updated",
			expectedLen: 1,
		},
		{
			name: "add new replica state",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Loaded},
				},
			},
			replicaIdx:  1,
			state:       ModelReplicaState_Loading,
			reason:      "new replica",
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mv.SetReplicaState(tt.replicaIdx, tt.state, tt.reason)
			g.Expect(tt.mv.Replicas).To(HaveLen(tt.expectedLen))
			g.Expect(tt.mv.Replicas[int32(tt.replicaIdx)].State).To(Equal(tt.state))
			g.Expect(tt.mv.Replicas[int32(tt.replicaIdx)].Reason).To(Equal(tt.reason))
			g.Expect(tt.mv.Replicas[int32(tt.replicaIdx)].Timestamp).NotTo(BeNil())
		})
	}
}

func TestModelVersion_GetModelReplicaState(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name       string
		mv         *ModelVersion
		replicaIdx int
		expected   ModelReplicaState
	}{
		{
			name: "existing replica",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Loaded},
				},
			},
			replicaIdx: 0,
			expected:   ModelReplicaState_Loaded,
		},
		{
			name: "non-existing replica",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Loaded},
				},
			},
			replicaIdx: 1,
			expected:   ModelReplicaState_ModelReplicaStateUnknown,
		},
		{
			name:       "empty replicas",
			mv:         &ModelVersion{},
			replicaIdx: 0,
			expected:   ModelReplicaState_ModelReplicaStateUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mv.GetModelReplicaState(tt.replicaIdx)
			g.Expect(result).To(Equal(tt.expected))
		})
	}
}

func TestModelVersion_Inactive(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name     string
		mv       *ModelVersion
		expected bool
	}{
		{
			name: "all replicas inactive",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Unloaded},
					1: {State: ModelReplicaState_UnloadFailed},
				},
			},
			expected: true,
		},
		{
			name: "has active replica",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Loaded},
					1: {State: ModelReplicaState_Unloaded},
				},
			},
			expected: false,
		},
		{
			name:     "empty replicas",
			mv:       &ModelVersion{Replicas: map[int32]*ReplicaStatus{}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mv.Inactive()
			g.Expect(result).To(Equal(tt.expected))
		})
	}
}

func TestModelVersion_DeleteReplica(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name        string
		mv          *ModelVersion
		replicaIdx  int
		expectedLen int
	}{
		{
			name: "delete existing replica",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Loaded},
					1: {State: ModelReplicaState_Available},
				},
			},
			replicaIdx:  0,
			expectedLen: 1,
		},
		{
			name: "delete non-existing replica",
			mv: &ModelVersion{
				Replicas: map[int32]*ReplicaStatus{
					0: {State: ModelReplicaState_Loaded},
				},
			},
			replicaIdx:  1,
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mv.DeleteReplica(tt.replicaIdx)
			g.Expect(tt.mv.Replicas).To(HaveLen(tt.expectedLen))
			_, exists := tt.mv.Replicas[int32(tt.replicaIdx)]
			g.Expect(exists).To(BeFalse())
		})
	}
}

func TestModel_Latest(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name     string
		model    *Model
		expected *ModelVersion
	}{
		{
			name: "has versions",
			model: &Model{
				Versions: []*ModelVersion{
					{Version: 1},
					{Version: 2},
					{Version: 3},
				},
			},
			expected: &ModelVersion{Version: 3},
		},
		{
			name:     "no versions",
			model:    &Model{Versions: []*ModelVersion{}},
			expected: nil,
		},
		{
			name:     "nil versions",
			model:    &Model{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.model.Latest()
			if tt.expected == nil {
				g.Expect(result).To(BeNil())
			} else {
				g.Expect(result.Version).To(Equal(tt.expected.Version))
			}
		})
	}
}

func TestModel_HasLatest(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name     string
		model    *Model
		expected bool
	}{
		{
			name: "has versions",
			model: &Model{
				Versions: []*ModelVersion{
					{Version: 1},
				},
			},
			expected: true,
		},
		{
			name:     "no versions",
			model:    &Model{Versions: []*ModelVersion{}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.model.HasLatest()
			g.Expect(result).To(Equal(tt.expected))
		})
	}
}

func TestModel_GetVersion(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name     string
		model    *Model
		version  uint32
		expected *ModelVersion
	}{
		{
			name: "version exists",
			model: &Model{
				Versions: []*ModelVersion{
					{Version: 1},
					{Version: 2},
					{Version: 3},
				},
			},
			version:  2,
			expected: &ModelVersion{Version: 2},
		},
		{
			name: "version does not exist",
			model: &Model{
				Versions: []*ModelVersion{
					{Version: 1},
					{Version: 2},
				},
			},
			version:  5,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.model.GetVersion(tt.version)
			if tt.expected == nil {
				g.Expect(result).To(BeNil())
			} else {
				g.Expect(result.Version).To(Equal(tt.expected.Version))
			}
		})
	}
}

func TestModel_GetLastAvailableModel(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name     string
		model    *Model
		expected *ModelVersion
	}{
		{
			name: "has available version",
			model: &Model{
				Versions: []*ModelVersion{
					{
						Version: 1,
						State:   &ModelStatus{State: ModelState_ModelAvailable},
					},
					{
						Version: 2,
						State:   &ModelStatus{State: ModelState_ModelProgressing},
					},
					{
						Version: 3,
						State:   &ModelStatus{State: ModelState_ModelAvailable},
					},
				},
			},
			expected: &ModelVersion{Version: 3},
		},
		{
			name: "no available version",
			model: &Model{
				Versions: []*ModelVersion{
					{
						Version: 1,
						State:   &ModelStatus{State: ModelState_ModelProgressing},
					},
				},
			},
			expected: nil,
		},
		{
			name:     "nil model",
			model:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.model.GetLastAvailableModel()
			if tt.expected == nil {
				g.Expect(result).To(BeNil())
			} else {
				g.Expect(result.Version).To(Equal(tt.expected.Version))
			}
		})
	}
}

func TestModel_CanReceiveTraffic(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name     string
		model    *Model
		expected bool
	}{
		{
			name: "has available model",
			model: &Model{
				Versions: []*ModelVersion{
					{
						Version: 1,
						State:   &ModelStatus{State: ModelState_ModelAvailable},
					},
				},
			},
			expected: true,
		},
		{
			name: "latest has live replicas",
			model: &Model{
				Versions: []*ModelVersion{
					{
						Version: 1,
						State:   &ModelStatus{State: ModelState_ModelProgressing},
						Replicas: map[int32]*ReplicaStatus{
							0: {State: ModelReplicaState_Loaded},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "no traffic-ready versions",
			model: &Model{
				Versions: []*ModelVersion{
					{
						Version: 1,
						State:   &ModelStatus{State: ModelState_ModelProgressing},
						Replicas: map[int32]*ReplicaStatus{
							0: {State: ModelReplicaState_Unloaded},
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.model.CanReceiveTraffic()
			g.Expect(result).To(Equal(tt.expected))
		})
	}
}

func TestModel_GetVersionsBeforeLastAvailable(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name     string
		model    *Model
		expected []*ModelVersion
	}{
		{
			name: "has versions before last available",
			model: &Model{
				Versions: []*ModelVersion{
					{Version: 1, State: &ModelStatus{State: ModelState_ModelAvailable}},
					{Version: 2, State: &ModelStatus{State: ModelState_ModelProgressing}},
					{Version: 3, State: &ModelStatus{State: ModelState_ModelAvailable}},
					{Version: 4, State: &ModelStatus{State: ModelState_ModelProgressing}},
				},
			},
			expected: []*ModelVersion{
				{Version: 1},
				{Version: 2},
			},
		},
		{
			name: "no available version",
			model: &Model{
				Versions: []*ModelVersion{
					{Version: 1, State: &ModelStatus{State: ModelState_ModelProgressing}},
				},
			},
			expected: nil,
		},
		{
			name: "first version is available",
			model: &Model{
				Versions: []*ModelVersion{
					{Version: 1, State: &ModelStatus{State: ModelState_ModelAvailable}},
					{Version: 2, State: &ModelStatus{State: ModelState_ModelProgressing}},
				},
			},
			expected: []*ModelVersion{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.model.GetVersionsBeforeLastAvailable()
			if tt.expected == nil {
				g.Expect(result).To(BeNil())
			} else {
				g.Expect(result).To(HaveLen(len(tt.expected)))
			}
		})
	}
}

func TestModel_GetVersionsBeforeLastModelGwAvailable(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name     string
		model    *Model
		expected []*ModelVersion
	}{
		{
			name: "has versions before last modelgw available",
			model: &Model{
				Versions: []*ModelVersion{
					{Version: 1, State: &ModelStatus{ModelGwState: ModelState_ModelAvailable}},
					{Version: 2, State: &ModelStatus{ModelGwState: ModelState_ModelProgressing}},
					{Version: 3, State: &ModelStatus{ModelGwState: ModelState_ModelAvailable}},
				},
			},
			expected: []*ModelVersion{
				{Version: 1},
				{Version: 2},
			},
		},
		{
			name: "no modelgw available version",
			model: &Model{
				Versions: []*ModelVersion{
					{Version: 1, State: &ModelStatus{ModelGwState: ModelState_ModelProgressing}},
				},
			},
			expected: nil,
		},
		{
			name:     "nil model",
			model:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.model.GetVersionsBeforeLastModelGwAvailable()
			if tt.expected == nil {
				g.Expect(result).To(BeNil())
			} else {
				g.Expect(result).To(HaveLen(len(tt.expected)))
			}
		})
	}
}

func TestModel_Inactive(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name     string
		model    *Model
		expected bool
	}{
		{
			name: "latest version inactive",
			model: &Model{
				Versions: []*ModelVersion{
					{
						Version: 1,
						Replicas: map[int32]*ReplicaStatus{
							0: {State: ModelReplicaState_Unloaded},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "latest version active",
			model: &Model{
				Versions: []*ModelVersion{
					{
						Version: 1,
						Replicas: map[int32]*ReplicaStatus{
							0: {State: ModelReplicaState_Loaded},
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.model.Inactive()
			g.Expect(result).To(Equal(tt.expected))
		})
	}
}

func TestModel_GetLastAvailableModelVersion(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name     string
		model    *Model
		expected *ModelVersion
	}{
		{
			name: "has available version",
			model: &Model{
				Versions: []*ModelVersion{
					{Version: 1, State: &ModelStatus{State: ModelState_ModelProgressing}},
					{Version: 2, State: &ModelStatus{State: ModelState_ModelAvailable}},
					{Version: 3, State: &ModelStatus{State: ModelState_ModelProgressing}},
				},
			},
			expected: &ModelVersion{Version: 2},
		},
		{
			name: "no available version",
			model: &Model{
				Versions: []*ModelVersion{
					{Version: 1, State: &ModelStatus{State: ModelState_ModelProgressing}},
				},
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.model.GetLastAvailableModelVersion()
			if tt.expected == nil {
				g.Expect(result).To(BeNil())
			} else {
				g.Expect(result.Version).To(Equal(tt.expected.Version))
			}
		})
	}
}

func TestModelReplicaState_CanReceiveTraffic(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name     string
		state    ModelReplicaState
		expected bool
	}{
		{name: "Loaded", state: ModelReplicaState_Loaded, expected: true},
		{name: "Available", state: ModelReplicaState_Available, expected: true},
		{name: "LoadedUnavailable", state: ModelReplicaState_LoadedUnavailable, expected: true},
		{name: "Draining", state: ModelReplicaState_Draining, expected: true},
		{name: "Loading", state: ModelReplicaState_Loading, expected: false},
		{name: "Unloaded", state: ModelReplicaState_Unloaded, expected: false},
		{name: "Unloading", state: ModelReplicaState_Unloading, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.state.CanReceiveTraffic()
			g.Expect(result).To(Equal(tt.expected))
		})
	}
}

func TestModelReplicaState_AlreadyLoadingOrLoaded(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name     string
		state    ModelReplicaState
		expected bool
	}{
		{name: "Loading", state: ModelReplicaState_Loading, expected: true},
		{name: "Loaded", state: ModelReplicaState_Loaded, expected: true},
		{name: "Available", state: ModelReplicaState_Available, expected: true},
		{name: "LoadedUnavailable", state: ModelReplicaState_LoadedUnavailable, expected: true},
		{name: "LoadRequested", state: ModelReplicaState_LoadRequested, expected: false},
		{name: "Unloaded", state: ModelReplicaState_Unloaded, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.state.AlreadyLoadingOrLoaded()
			g.Expect(result).To(Equal(tt.expected))
		})
	}
}

func TestModelReplicaState_UnloadingOrUnloaded(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name     string
		state    ModelReplicaState
		expected bool
	}{
		{name: "UnloadEnvoyRequested", state: ModelReplicaState_UnloadEnvoyRequested, expected: true},
		{name: "UnloadRequested", state: ModelReplicaState_UnloadRequested, expected: true},
		{name: "Unloading", state: ModelReplicaState_Unloading, expected: true},
		{name: "Unloaded", state: ModelReplicaState_Unloaded, expected: true},
		{name: "ModelReplicaStateUnknown", state: ModelReplicaState_ModelReplicaStateUnknown, expected: true},
		{name: "Loaded", state: ModelReplicaState_Loaded, expected: false},
		{name: "Available", state: ModelReplicaState_Available, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.state.UnloadingOrUnloaded()
			g.Expect(result).To(Equal(tt.expected))
		})
	}
}

func TestModelReplicaState_Inactive(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name     string
		state    ModelReplicaState
		expected bool
	}{
		{name: "Unloaded", state: ModelReplicaState_Unloaded, expected: true},
		{name: "UnloadFailed", state: ModelReplicaState_UnloadFailed, expected: true},
		{name: "ModelReplicaStateUnknown", state: ModelReplicaState_ModelReplicaStateUnknown, expected: true},
		{name: "LoadFailed", state: ModelReplicaState_LoadFailed, expected: true},
		{name: "Loaded", state: ModelReplicaState_Loaded, expected: false},
		{name: "Available", state: ModelReplicaState_Available, expected: false},
		{name: "Loading", state: ModelReplicaState_Loading, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.state.Inactive()
			g.Expect(result).To(Equal(tt.expected))
		})
	}
}

func TestModelReplicaState_IsLoadingOrLoaded(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name     string
		state    ModelReplicaState
		expected bool
	}{
		{name: "Loaded", state: ModelReplicaState_Loaded, expected: true},
		{name: "LoadRequested", state: ModelReplicaState_LoadRequested, expected: true},
		{name: "Loading", state: ModelReplicaState_Loading, expected: true},
		{name: "Available", state: ModelReplicaState_Available, expected: true},
		{name: "LoadedUnavailable", state: ModelReplicaState_LoadedUnavailable, expected: true},
		{name: "Unloaded", state: ModelReplicaState_Unloaded, expected: false},
		{name: "Unloading", state: ModelReplicaState_Unloading, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.state.IsLoadingOrLoaded()
			g.Expect(result).To(Equal(tt.expected))
		})
	}
}