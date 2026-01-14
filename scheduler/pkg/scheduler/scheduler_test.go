/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package scheduler

import (
	"errors"
	"maps"
	"slices"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gotidy/ptr"
	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	ptr2 "k8s.io/utils/ptr"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	pbs "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/mock"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/synchroniser"
	mock2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/synchroniser/mock"
)

func TestScheduler(t *testing.T) {
	logger := log.New()
	g := NewGomegaWithT(t)

	newTestModel := func(name string, requiredMemory uint64, requirements []string, replicas, minReplicas uint32, maxReplicas uint32, loadedModels []int, deleted bool, scheduledServer string, drainedModels []int) *store.ModelSnapshot {
		config := &pb.Model{Meta: &pb.MetaData{Name: t.Name()}, ModelSpec: &pb.ModelSpec{MemoryBytes: &requiredMemory, Requirements: requirements}, DeploymentSpec: &pb.DeploymentSpec{Replicas: replicas, MinReplicas: minReplicas, MaxReplicas: maxReplicas}}
		rmap := make(map[int]store.ReplicaStatus)
		for _, ridx := range loadedModels {
			rmap[ridx] = store.ReplicaStatus{State: store.Loaded}
		}
		for _, ridx := range drainedModels {
			rmap[ridx] = store.ReplicaStatus{State: store.Draining}
		}
		return &store.ModelSnapshot{
			Name:     name,
			Versions: []*store.ModelVersion{store.NewModelVersion(config, 1, scheduledServer, rmap, false, store.ModelProgressing)},
			Deleted:  deleted,
		}
	}

	gsr := func(replicaIdx int, availableMemory uint64, capabilities []string, serverName string, shared, isDraining bool) *store.ServerReplica {
		replica := store.NewServerReplica("svc", 8080, 5001, replicaIdx, store.NewServer(serverName, shared), capabilities, availableMemory, availableMemory, 0, nil, 100)
		if isDraining {
			replica.SetIsDraining()
		}
		return replica
	}

	type test struct {
		name                 string
		model                *store.ModelSnapshot
		modelName            string
		servers              []*store.ServerSnapshot
		scheduled            bool
		checkServerEvents    bool
		expectedServerEvents int
		setupMock            func(m *mock.MockModelServerAPI)
	}

	tests := []test{
		{
			name:      "SmokeTest",
			modelName: "model1",
			scheduled: true,
			setupMock: func(m *mock.MockModelServerAPI) {
				m.EXPECT().LockModel("model1")
				m.EXPECT().GetModel("model1").Return(&db.Model{Name: "model1", Versions: []*db.ModelVersion{
					util.NewTestModelVersion(
						&pbs.Model{
							Meta: &pbs.MetaData{Name: "model1"},
							ModelSpec: &pbs.ModelSpec{
								Uri:              "",
								ArtifactVersion:  nil,
								StorageConfig:    nil,
								Requirements:     []string{"sklearn"},
								MemoryBytes:      ptr2.To(uint64(200)),
								Server:           nil,
								Parameters:       nil,
								ModelRuntimeInfo: nil,
								ModelSpec:        nil,
							},
							DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1, MinReplicas: 0, MaxReplicas: 1},
						},
						1, "server1",
						nil, db.ModelState_ModelProgressing),
				}}, nil).MinTimes(1)
				m.EXPECT().UnlockModel("model1")
				servers := []*db.Server{
					{
						Name: "server1",
						Replicas: map[int32]*db.ServerReplica{
							1: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server1", false), []string{"sklearn"}, 0, 200, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: 1,
						MinReplicas:      1,
						MaxReplicas:      1,
						KubernetesMeta:   nil,
					},
				}
				m.EXPECT().GetServers().Return(
					servers,
					nil,
				)
				m.EXPECT().UpdateLoadedModels("model1", uint32(1),
					"server1", slices.Collect(maps.Values(servers[0].Replicas))).Return(nil)
			},
		},
		{
			name:      "ReplicasTwo",
			model:     newTestModel("model1", 100, []string{"sklearn"}, 2, 0, 2, []int{}, false, "", nil),
			modelName: "model1",
			servers: []*store.ServerSnapshot{
				{
					Name:             "server1",
					Replicas:         map[int]*store.ServerReplica{0: gsr(0, 200, []string{"sklearn"}, "server1", true, false)},
					Shared:           true,
					ExpectedReplicas: -1,
				},
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}, "server2", true, false), // expect schedule here
						1: gsr(1, 200, []string{"sklearn"}, "server2", true, false), // expect schedule here
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled: true,
			setupMock: func(m *mock.MockModelServerAPI) {
				m.EXPECT().LockModel("model1")
				m.EXPECT().GetModel("model1").Return(&db.Model{Name: "model1", Versions: []*db.ModelVersion{
					util.NewTestModelVersion(
						&pbs.Model{
							Meta: &pbs.MetaData{Name: "model1"},
							ModelSpec: &pbs.ModelSpec{
								Uri:              "",
								ArtifactVersion:  nil,
								StorageConfig:    nil,
								Requirements:     []string{"sklearn"},
								MemoryBytes:      ptr2.To(uint64(200)),
								Server:           nil,
								Parameters:       nil,
								ModelRuntimeInfo: nil,
								ModelSpec:        nil,
							},
							DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1, MinReplicas: 0, MaxReplicas: 1},
						},
						1, "server1",
						nil, db.ModelState_ModelProgressing),
				}}, nil).MinTimes(1)
				m.EXPECT().UnlockModel("model1")
				servers := []*db.Server{
					{
						Name: "server1",
						Replicas: map[int32]*db.ServerReplica{
							1: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server1", false), []string{"sklearn"}, 0, 200, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
					{
						Name: "server2",
						Replicas: map[int32]*db.ServerReplica{
							1: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server2", false), []string{"sklearn"}, 0, 200, 0, nil, 100),
							2: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server2", false), []string{"sklearn"}, 0, 200, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
				}

				expectedUpdatedServers := []*db.ServerReplica{
					util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server1", false), []string{"sklearn"}, 0, 200, 0, nil, 100),
				}
				m.EXPECT().GetServers().Return(
					servers,
					nil,
				)
				m.EXPECT().UpdateLoadedModels("model1", uint32(1),
					"server1", expectedUpdatedServers).Return(nil)
			},
		},
		{
			name:      "NotEnoughReplicas",
			model:     newTestModel("model1", 100, []string{"sklearn"}, 2, 0, 2, []int{}, false, "", nil),
			modelName: "model1",
			servers: []*store.ServerSnapshot{
				{
					Name:             "server1",
					Replicas:         map[int]*store.ServerReplica{0: gsr(0, 200, []string{"sklearn"}, "server1", true, false)},
					Shared:           true,
					ExpectedReplicas: -1,
				},
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}, "server2", true, false),
						1: gsr(1, 0, []string{"sklearn"}, "server2", true, false),
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled: false,
			setupMock: func(m *mock.MockModelServerAPI) {
				m.EXPECT().LockModel("model1")
				m.EXPECT().GetModel("model1").Return(&db.Model{Name: "model1", Versions: []*db.ModelVersion{
					util.NewTestModelVersion(
						&pbs.Model{
							Meta: &pbs.MetaData{Name: "model1"},
							ModelSpec: &pbs.ModelSpec{
								Uri:              "",
								ArtifactVersion:  nil,
								StorageConfig:    nil,
								Requirements:     []string{"sklearn"},
								MemoryBytes:      ptr2.To(uint64(100)),
								Server:           nil,
								Parameters:       nil,
								ModelRuntimeInfo: nil,
								ModelSpec:        nil,
							},
							DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MinReplicas: 0, MaxReplicas: 2},
						},
						1, "server1",
						nil, db.ModelState_ModelProgressing),
				}}, nil).MinTimes(1)
				m.EXPECT().UnlockModel("model1")
				servers := []*db.Server{
					{
						Name: "server1",
						Replicas: map[int32]*db.ServerReplica{
							1: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server1", false), []string{"sklearn"}, 0, 200, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
					{
						Name: "server2",
						Replicas: map[int32]*db.ServerReplica{
							1: util.NewTestServerReplica("host1", 8080, 5000, 0, store.NewServer("server2", false), []string{"sklearn"}, 0, 200, 0, nil, 100),
							2: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server2", false), []string{"sklearn"}, 0, 0, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
				}
				m.EXPECT().GetServers().Return(
					servers,
					nil,
				)
				m.EXPECT().FailedScheduling("model1", gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		{
			name:      "NotEnoughReplicas - schedule min replicas",
			model:     newTestModel("model1", 100, []string{"sklearn"}, 3, 2, 3, []int{}, false, "server2", nil),
			modelName: "model1",
			servers: []*store.ServerSnapshot{
				{
					Name:             "server1",
					Replicas:         map[int]*store.ServerReplica{0: gsr(0, 200, []string{"sklearn"}, "server1", true, false)},
					Shared:           true,
					ExpectedReplicas: -1,
				},
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}, "server2", true, false), // expect schedule here
						1: gsr(1, 200, []string{"sklearn"}, "server2", true, false), // expect schedule here
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled: true, // not here that we still trying to mark the model as Available
			setupMock: func(m *mock.MockModelServerAPI) {
				m.EXPECT().LockModel("model1")
				m.EXPECT().GetModel("model1").Return(&db.Model{Name: "model1", Versions: []*db.ModelVersion{
					util.NewTestModelVersion(
						&pbs.Model{
							Meta: &pbs.MetaData{Name: "model1"},
							ModelSpec: &pbs.ModelSpec{
								Uri:              "",
								ArtifactVersion:  nil,
								StorageConfig:    nil,
								Requirements:     []string{"sklearn"},
								MemoryBytes:      ptr2.To(uint64(200)),
								Server:           nil,
								Parameters:       nil,
								ModelRuntimeInfo: nil,
								ModelSpec:        nil,
							},
							DeploymentSpec: &pbs.DeploymentSpec{Replicas: 3, MinReplicas: 2, MaxReplicas: 3},
						},
						1, "server2",
						nil, db.ModelState_ModelProgressing),
				}}, nil).MinTimes(1)
				m.EXPECT().UnlockModel("model1")
				servers := []*db.Server{
					{
						Name: "server1",
						Replicas: map[int32]*db.ServerReplica{
							1: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server1", false), []string{"sklearn"}, 0, 200, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
					{
						Name: "server2",
						Replicas: map[int32]*db.ServerReplica{
							1: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server2", false), []string{"sklearn"}, 0, 200, 0, nil, 100),
							2: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server2", false), []string{"sklearn"}, 0, 200, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
				}
				expectedUpdatedServers := []*db.ServerReplica{
					util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server2", false), []string{"sklearn"}, 0, 200, 0, nil, 100),
					util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server2", false), []string{"sklearn"}, 0, 200, 0, nil, 100),
				}
				m.EXPECT().GetServers().Return(
					servers,
					nil,
				)
				m.EXPECT().UpdateLoadedModels("model1", uint32(1),
					"server2", expectedUpdatedServers).Return(nil)
			},
			expectedServerEvents: 1,
		},
		{
			name:      "MemoryOneServer",
			model:     newTestModel("model1", 100, []string{"sklearn"}, 1, 0, 1, []int{}, false, "", nil),
			modelName: "model1",
			servers: []*store.ServerSnapshot{
				{
					Name:             "server1",
					Replicas:         map[int]*store.ServerReplica{0: gsr(0, 50, []string{"sklearn"}, "server1", true, false)},
					Shared:           true,
					ExpectedReplicas: -1,
				},
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}, "server2", true, false), // expect schedule here
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled: true,
			setupMock: func(m *mock.MockModelServerAPI) {
				m.EXPECT().LockModel("model1")
				m.EXPECT().GetModel("model1").Return(&db.Model{Name: "model1", Versions: []*db.ModelVersion{
					util.NewTestModelVersion(
						&pbs.Model{
							Meta: &pbs.MetaData{Name: "model1"},
							ModelSpec: &pbs.ModelSpec{
								Uri:              "",
								ArtifactVersion:  nil,
								StorageConfig:    nil,
								Requirements:     []string{"sklearn"},
								MemoryBytes:      ptr2.To(uint64(200)),
								Server:           nil,
								Parameters:       nil,
								ModelRuntimeInfo: nil,
								ModelSpec:        nil,
							},
							DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1, MinReplicas: 0, MaxReplicas: 1},
						},
						1, "server2",
						nil, db.ModelState_ModelProgressing),
				}}, nil).MinTimes(1)
				m.EXPECT().UnlockModel("model1")
				servers := []*db.Server{
					{
						Name: "server1",
						Replicas: map[int32]*db.ServerReplica{
							1: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server1", false), []string{"sklearn"}, 0, 50, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
					{
						Name: "server2",
						Replicas: map[int32]*db.ServerReplica{
							1: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server2", false), []string{"sklearn"}, 0, 200, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
				}
				expectedUpdatedServers := []*db.ServerReplica{
					util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server2", false), []string{"sklearn"}, 0, 200, 0, nil, 100),
				}
				m.EXPECT().GetServers().Return(
					servers,
					nil,
				)
				m.EXPECT().UpdateLoadedModels("model1", uint32(1),
					"server2", expectedUpdatedServers).Return(nil)
			},
		},
		{
			name:      "ModelsLoaded",
			model:     newTestModel("model1", 100, []string{"sklearn"}, 2, 0, 2, []int{1}, false, "", nil),
			modelName: "model1",
			servers: []*store.ServerSnapshot{
				{
					Name:             "server1",
					Replicas:         map[int]*store.ServerReplica{0: gsr(0, 50, []string{"sklearn"}, "server1", true, false)},
					Shared:           true,
					ExpectedReplicas: -1,
				},
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}, "server2", true, false), // expect schedule here
						1: gsr(1, 200, []string{"sklearn"}, "server2", true, false), // expect schedule here
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled: true,
			setupMock: func(m *mock.MockModelServerAPI) {
				m.EXPECT().LockModel("model1")
				m.EXPECT().GetModel("model1").Return(&db.Model{Name: "model1", Versions: []*db.ModelVersion{
					util.NewTestModelVersion(
						&pbs.Model{
							Meta: &pbs.MetaData{Name: "model1"},
							ModelSpec: &pbs.ModelSpec{
								Uri:              "",
								ArtifactVersion:  nil,
								StorageConfig:    nil,
								Requirements:     []string{"sklearn"},
								MemoryBytes:      ptr2.To(uint64(200)),
								Server:           nil,
								Parameters:       nil,
								ModelRuntimeInfo: nil,
								ModelSpec:        nil,
							},
							DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1, MinReplicas: 0, MaxReplicas: 1},
						},
						1, "server2",
						nil, db.ModelState_ModelProgressing),
				}}, nil).MinTimes(1)
				m.EXPECT().UnlockModel("model1")
				servers := []*db.Server{
					{
						Name: "server1",
						Replicas: map[int32]*db.ServerReplica{
							1: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server1", true), []string{"sklearn"}, 0, 50, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
					{
						Name: "server2",
						Replicas: map[int32]*db.ServerReplica{
							1: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server2", true), []string{"sklearn"}, 0, 200, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
				}
				expectedUpdatedServers := []*db.ServerReplica{
					util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server2", true), []string{"sklearn"}, 0, 200, 0, nil, 100),
				}
				m.EXPECT().GetServers().Return(
					servers,
					nil,
				)
				m.EXPECT().UpdateLoadedModels("model1", uint32(1),
					"server2", expectedUpdatedServers).Return(nil)
			},
		},
		{
			name:      "ModelUnLoaded",
			model:     newTestModel("model1", 100, []string{"sklearn"}, 2, 0, 2, []int{1}, true, "server2", nil),
			modelName: "model1",
			servers: []*store.ServerSnapshot{
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}, "server2", true, false),
						1: gsr(1, 200, []string{"sklearn"}, "server2", true, false),
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled: true,
			setupMock: func(m *mock.MockModelServerAPI) {
				m.EXPECT().LockModel("model1")
				m.EXPECT().GetModel("model1").Return(&db.Model{Name: "model1", Versions: []*db.ModelVersion{
					util.NewTestModelVersion(
						&pbs.Model{
							Meta: &pbs.MetaData{Name: "model1"},
							ModelSpec: &pbs.ModelSpec{
								Uri:              "",
								ArtifactVersion:  nil,
								StorageConfig:    nil,
								Requirements:     []string{"sklearn"},
								MemoryBytes:      ptr2.To(uint64(200)),
								Server:           nil,
								Parameters:       nil,
								ModelRuntimeInfo: nil,
								ModelSpec:        nil,
							},
							DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1, MinReplicas: 0, MaxReplicas: 1},
						},
						1, "server2",
						map[int32]*db.ReplicaStatus{
							1: {
								State:     db.ModelReplicaState_Loaded,
								Reason:    "",
								Timestamp: nil,
							},
						},
						db.ModelState_ModelProgressing),
				}}, nil).MinTimes(1)
				m.EXPECT().UnlockModel("model1")
				servers := []*db.Server{
					{
						Name: "server2",
						Replicas: map[int32]*db.ServerReplica{
							0: util.NewTestServerReplica("host1", 8080, 5000, 0, store.NewServer("server2", true), []string{"sklearn"}, 0, 200, 0, nil, 100),
							1: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server2", true), []string{"sklearn"}, 0, 200, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
				}
				expectedUpdatedServers := []*db.ServerReplica{
					util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server2", true), []string{"sklearn"}, 0, 200, 0, nil, 100),
				}
				m.EXPECT().GetServers().Return(
					servers,
					nil,
				)
				m.EXPECT().UpdateLoadedModels("model1", uint32(1),
					"server2", expectedUpdatedServers).Return(nil)
			},
		},
		{
			name:      "DeletedServer",
			model:     newTestModel("model1", 100, []string{"sklearn"}, 1, 0, 1, []int{}, false, "", nil),
			modelName: "model1",
			servers: []*store.ServerSnapshot{
				{
					Name:             "server1",
					Replicas:         map[int]*store.ServerReplica{0: gsr(0, 200, []string{"sklearn"}, "server1", true, false)},
					Shared:           true,
					ExpectedReplicas: 0,
				},
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}, "server2", true, false), // expect schedule here
						1: gsr(1, 200, []string{"sklearn"}, "server2", true, false),
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled: true,
			setupMock: func(m *mock.MockModelServerAPI) {
				m.EXPECT().LockModel("model1")
				m.EXPECT().GetModel("model1").Return(&db.Model{Name: "model1", Versions: []*db.ModelVersion{
					util.NewTestModelVersion(
						&pbs.Model{
							Meta: &pbs.MetaData{Name: "model1"},
							ModelSpec: &pbs.ModelSpec{
								Uri:              "",
								ArtifactVersion:  nil,
								StorageConfig:    nil,
								Requirements:     []string{"sklearn"},
								MemoryBytes:      ptr2.To(uint64(200)),
								Server:           nil,
								Parameters:       nil,
								ModelRuntimeInfo: nil,
								ModelSpec:        nil,
							},
							DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1, MinReplicas: 0, MaxReplicas: 1},
						},
						1, "server2",
						nil, db.ModelState_ModelProgressing),
				}}, nil).MinTimes(1)
				m.EXPECT().UnlockModel("model1")
				servers := []*db.Server{
					{
						Name: "server1",
						Replicas: map[int32]*db.ServerReplica{
							1: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server1", true), []string{"sklearn"}, 0, 0, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
					{
						Name: "server2",
						Replicas: map[int32]*db.ServerReplica{
							1: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server2", true), []string{"sklearn"}, 0, 200, 0, nil, 100),
							2: util.NewTestServerReplica("host1", 8080, 5000, 2, store.NewServer("server2", true), []string{"sklearn"}, 0, 200, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
				}

				expectedUpdatedServers := []*db.ServerReplica{
					util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server2", false), []string{"sklearn"}, 0, 200, 0, nil, 100),
				}

				m.EXPECT().GetServers().Return(
					servers,
					nil,
				)
				m.EXPECT().UpdateLoadedModels("model1", uint32(1),
					"server2", expectedUpdatedServers).Return(nil)
			},
		},
		{
			name:      "Reschedule",
			model:     newTestModel("model1", 100, []string{"sklearn"}, 1, 0, 1, []int{0}, false, "server1", nil),
			modelName: "model1",
			servers: []*store.ServerSnapshot{
				{
					Name:             "server1",
					Replicas:         map[int]*store.ServerReplica{0: gsr(0, 200, []string{"sklearn"}, "server1", true, false)},
					Shared:           true,
					ExpectedReplicas: 0,
				},
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}, "server2", true, false), // expect schedule here
						1: gsr(1, 200, []string{"sklearn"}, "server2", true, false),
					},
					Shared:           true,
					ExpectedReplicas: 1,
				},
			},
			scheduled: true,
			setupMock: func(m *mock.MockModelServerAPI) {
				m.EXPECT().LockModel("model1")
				m.EXPECT().GetModel("model1").Return(&db.Model{Name: "model1", Versions: []*db.ModelVersion{
					util.NewTestModelVersion(
						&pbs.Model{
							Meta: &pbs.MetaData{Name: "model1"},
							ModelSpec: &pbs.ModelSpec{
								Uri:              "",
								ArtifactVersion:  nil,
								StorageConfig:    nil,
								Requirements:     []string{"sklearn"},
								MemoryBytes:      ptr2.To(uint64(200)),
								Parameters:       nil,
								ModelRuntimeInfo: nil,
								ModelSpec:        nil,
							},
							DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1, MinReplicas: 0, MaxReplicas: 1},
						},
						1, "server1",
						map[int32]*db.ReplicaStatus{
							0: {
								State:     db.ModelReplicaState_Loaded,
								Reason:    "",
								Timestamp: nil,
							},
						},
						db.ModelState_ModelProgressing),
				}}, nil).MinTimes(1)
				m.EXPECT().UnlockModel("model1")
				servers := []*db.Server{
					{
						Name: "server1",
						Replicas: map[int32]*db.ServerReplica{
							1: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server1", true), []string{"sklearn"}, 0, 200, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: 0,
						KubernetesMeta:   nil,
					},
					{
						Name: "server2",
						Replicas: map[int32]*db.ServerReplica{
							0: util.NewTestServerReplica("host1", 8080, 5000, 0, store.NewServer("server2", true), []string{"sklearn"}, 0, 200, 0, nil, 100),
							1: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server2", true), []string{"sklearn"}, 0, 200, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
				}

				expectedUpdatedServers := []*db.ServerReplica{
					util.NewTestServerReplica("host1", 8080, 5000, 0, store.NewServer("server2", true), []string{"sklearn"}, 0, 200, 0, nil, 100),
				}

				m.EXPECT().GetServers().Return(
					servers,
					nil,
				)
				m.EXPECT().UpdateLoadedModels("model1", uint32(1),
					"server2", expectedUpdatedServers).Return(nil)
			},
		},
		{
			name:      "DeletedServerFail",
			model:     newTestModel("model1", 100, []string{"sklearn"}, 1, 0, 1, []int{1}, false, "", nil),
			modelName: "model1",
			servers: []*store.ServerSnapshot{
				{
					Name:             "server1",
					Replicas:         map[int]*store.ServerReplica{0: gsr(0, 200, []string{"sklearn"}, "server1", true, false)},
					Shared:           true,
					ExpectedReplicas: 0,
				},
			},
			scheduled: false,
			setupMock: func(m *mock.MockModelServerAPI) {
				m.EXPECT().LockModel("model1")
				m.EXPECT().GetModel("model1").Return(&db.Model{Name: "model1", Versions: []*db.ModelVersion{
					util.NewTestModelVersion(
						&pbs.Model{
							Meta: &pbs.MetaData{Name: "model1"},
							ModelSpec: &pbs.ModelSpec{
								Uri:              "",
								ArtifactVersion:  nil,
								StorageConfig:    nil,
								Requirements:     []string{"sklearn"},
								MemoryBytes:      ptr2.To(uint64(100)),
								Server:           nil,
								Parameters:       nil,
								ModelRuntimeInfo: nil,
								ModelSpec:        nil,
							},
							DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1, MinReplicas: 0, MaxReplicas: 1},
						},
						1, "",
						map[int32]*db.ReplicaStatus{
							0: {
								State:     db.ModelReplicaState_Loaded,
								Reason:    "",
								Timestamp: nil,
							},
						},
						db.ModelState_ModelProgressing),
				}}, nil).MinTimes(1)
				m.EXPECT().UnlockModel("model1")
				servers := []*db.Server{
					{
						Name: "server1",
						Replicas: map[int32]*db.ServerReplica{
							0: util.NewTestServerReplica("host1", 8080, 5000, 0, store.NewServer("server1", false), []string{"sklearn"}, 0, 50, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
				}
				m.EXPECT().GetServers().Return(
					servers,
					nil,
				)
				m.EXPECT().FailedScheduling("model1", gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		{
			name:      "Available memory sorting",
			model:     newTestModel("model1", 100, []string{"sklearn"}, 1, 0, 1, []int{1}, false, "", nil),
			modelName: "model1",
			servers: []*store.ServerSnapshot{
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 150, []string{"sklearn"}, "server2", true, false),
						1: gsr(1, 200, []string{"sklearn"}, "server2", true, false), // expect schedule here
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled: true,
			setupMock: func(m *mock.MockModelServerAPI) {
				m.EXPECT().LockModel("model1")
				m.EXPECT().GetModel("model1").Return(&db.Model{Name: "model1", Versions: []*db.ModelVersion{
					util.NewTestModelVersion(
						&pbs.Model{
							Meta: &pbs.MetaData{Name: "model1"},
							ModelSpec: &pbs.ModelSpec{
								Uri:              "",
								ArtifactVersion:  nil,
								StorageConfig:    nil,
								Requirements:     []string{"sklearn"},
								MemoryBytes:      ptr2.To(uint64(100)),
								Server:           nil,
								Parameters:       nil,
								ModelRuntimeInfo: nil,
								ModelSpec:        nil,
							},
							DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1, MinReplicas: 0, MaxReplicas: 1},
						},
						1, "server2",
						nil, db.ModelState_ModelProgressing),
				}}, nil).MinTimes(1)
				m.EXPECT().UnlockModel("model1")
				servers := []*db.Server{
					{
						Name: "server2",
						Replicas: map[int32]*db.ServerReplica{
							0: util.NewTestServerReplica("host1", 8080, 5000, 0, store.NewServer("server2", true), []string{"sklearn"}, 0, 150, 0, nil, 100),
							1: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server2", true), []string{"sklearn"}, 0, 200, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
				}

				expectedUpdatedServers := []*db.ServerReplica{
					util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server2", true), []string{"sklearn"}, 0, 200, 0, nil, 100),
				}

				m.EXPECT().GetServers().Return(
					servers,
					nil,
				)
				m.EXPECT().UpdateLoadedModels("model1", uint32(1),
					"server2", expectedUpdatedServers).Return(nil)
			},
		},
		{
			name:      "Available memory sorting with multiple replicas",
			model:     newTestModel("model1", 100, []string{"sklearn"}, 2, 0, 1, []int{1}, false, "", nil),
			modelName: "model1",
			servers: []*store.ServerSnapshot{
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 150, []string{"sklearn"}, "server2", true, false),
						1: gsr(1, 200, []string{"sklearn"}, "server2", true, false), // expect schedule here
						2: gsr(2, 175, []string{"sklearn"}, "server2", true, false), // expect schedule here
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled: true,
			setupMock: func(m *mock.MockModelServerAPI) {
				m.EXPECT().LockModel("model1")
				m.EXPECT().GetModel("model1").Return(&db.Model{Name: "model1", Versions: []*db.ModelVersion{
					util.NewTestModelVersion(
						&pbs.Model{
							Meta: &pbs.MetaData{Name: "model1"},
							ModelSpec: &pbs.ModelSpec{
								Uri:              "",
								ArtifactVersion:  nil,
								StorageConfig:    nil,
								Requirements:     []string{"sklearn"},
								MemoryBytes:      ptr2.To(uint64(100)),
								Server:           nil,
								Parameters:       nil,
								ModelRuntimeInfo: nil,
								ModelSpec:        nil,
							},
							DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MinReplicas: 0, MaxReplicas: 1},
						},
						1, "server2",
						nil, db.ModelState_ModelProgressing),
				}}, nil).MinTimes(1)
				m.EXPECT().UnlockModel("model1")
				servers := []*db.Server{
					{
						Name: "server2",
						Replicas: map[int32]*db.ServerReplica{
							0: util.NewTestServerReplica("host1", 8080, 5000, 0, store.NewServer("server2", true), []string{"sklearn"}, 0, 150, 0, nil, 100),
							1: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server2", true), []string{"sklearn"}, 0, 200, 0, nil, 100),
							2: util.NewTestServerReplica("host1", 8080, 5000, 2, store.NewServer("server2", true), []string{"sklearn"}, 0, 175, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
				}

				expectedUpdatedServers := []*db.ServerReplica{
					util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server2", true), []string{"sklearn"}, 0, 200, 0, nil, 100),
					util.NewTestServerReplica("host1", 8080, 5000, 2, store.NewServer("server2", true), []string{"sklearn"}, 0, 175, 0, nil, 100),
				}

				m.EXPECT().GetServers().Return(
					servers,
					nil,
				)
				m.EXPECT().UpdateLoadedModels("model1", uint32(1),
					"server2", expectedUpdatedServers).Return(nil)
			},
		},
		{
			name:      "Scale up",
			model:     newTestModel("model1", 100, []string{"sklearn"}, 3, 0, 3, []int{1, 2}, false, "server1", nil),
			modelName: "model1",
			servers: []*store.ServerSnapshot{
				{
					Name: "server1",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 50, []string{"sklearn"}, "server1", true, false),
						1: gsr(1, 200, []string{"sklearn"}, "server1", true, false), // expect schedule here - nop
						2: gsr(2, 175, []string{"sklearn"}, "server1", true, false), // expect schedule here - nop
						3: gsr(3, 100, []string{"sklearn"}, "server1", true, false), // expect schedule here
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled: true,
			setupMock: func(m *mock.MockModelServerAPI) {
				m.EXPECT().LockModel("model1")
				m.EXPECT().GetModel("model1").Return(&db.Model{Name: "model1", Versions: []*db.ModelVersion{
					util.NewTestModelVersion(
						&pbs.Model{
							Meta: &pbs.MetaData{Name: "model1"},
							ModelSpec: &pbs.ModelSpec{
								Uri:              "",
								ArtifactVersion:  nil,
								StorageConfig:    nil,
								Requirements:     []string{"sklearn"},
								MemoryBytes:      ptr2.To(uint64(100)),
								Server:           nil,
								Parameters:       nil,
								ModelRuntimeInfo: nil,
								ModelSpec:        nil,
							},
							DeploymentSpec: &pbs.DeploymentSpec{Replicas: 3, MinReplicas: 0, MaxReplicas: 3},
						},
						1, "server1",
						map[int32]*db.ReplicaStatus{

							1: {
								State:     db.ModelReplicaState_Loaded,
								Reason:    "",
								Timestamp: nil,
							},
							2: {
								State:     db.ModelReplicaState_Loaded,
								Reason:    "",
								Timestamp: nil,
							},
						},
						db.ModelState_ModelProgressing),
				}}, nil).MinTimes(1)
				m.EXPECT().UnlockModel("model1")
				servers := []*db.Server{
					{
						Name: "server1",
						Replicas: map[int32]*db.ServerReplica{
							0: util.NewTestServerReplica("host1", 8080, 5000, 0, store.NewServer("server1", true), []string{"sklearn"}, 0, 50, 0, nil, 100),
							1: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server1", true), []string{"sklearn"}, 0, 200, 0, nil, 100),
							2: util.NewTestServerReplica("host1", 8080, 5000, 2, store.NewServer("server1", true), []string{"sklearn"}, 0, 175, 0, nil, 100),
							3: util.NewTestServerReplica("host1", 8080, 5000, 3, store.NewServer("server1", true), []string{"sklearn"}, 0, 100, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
				}
				expectedUpdatedServers := []*db.ServerReplica{
					util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server1", true), []string{"sklearn"}, 0, 200, 0, nil, 100),
					util.NewTestServerReplica("host1", 8080, 5000, 2, store.NewServer("server1", true), []string{"sklearn"}, 0, 175, 0, nil, 100),
					util.NewTestServerReplica("host1", 8080, 5000, 3, store.NewServer("server1", true), []string{"sklearn"}, 0, 100, 0, nil, 100),
				}
				m.EXPECT().GetServers().Return(
					servers,
					nil,
				)
				m.EXPECT().UpdateLoadedModels("model1", uint32(1),
					"server1", expectedUpdatedServers).Return(nil)
			},
		},
		{
			name:      "Scale down",
			model:     newTestModel("model1", 100, []string{"sklearn"}, 1, 0, 1, []int{1, 2}, false, "server1", nil),
			modelName: "model1",
			servers: []*store.ServerSnapshot{
				{
					Name: "server1",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 50, []string{"sklearn"}, "server1", true, false),
						1: gsr(1, 200, []string{"sklearn"}, "server1", true, false), // expect schedule here - nop
						2: gsr(2, 175, []string{"sklearn"}, "server1", true, false), // expect schedule here - nop
						3: gsr(3, 100, []string{"sklearn"}, "server1", true, false),
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled: true,
			setupMock: func(m *mock.MockModelServerAPI) {
				m.EXPECT().LockModel("model1")
				m.EXPECT().GetModel("model1").Return(&db.Model{Name: "model1", Versions: []*db.ModelVersion{
					util.NewTestModelVersion(
						&pbs.Model{
							Meta: &pbs.MetaData{Name: "model1"},
							ModelSpec: &pbs.ModelSpec{
								Uri:              "",
								ArtifactVersion:  nil,
								StorageConfig:    nil,
								Requirements:     []string{"sklearn"},
								MemoryBytes:      ptr2.To(uint64(100)),
								Server:           nil,
								Parameters:       nil,
								ModelRuntimeInfo: nil,
								ModelSpec:        nil,
							},
							DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1, MinReplicas: 0, MaxReplicas: 1},
						},
						1, "server1",
						map[int32]*db.ReplicaStatus{

							1: {
								State:     db.ModelReplicaState_Loaded,
								Reason:    "",
								Timestamp: nil,
							},
							2: {
								State:     db.ModelReplicaState_Loaded,
								Reason:    "",
								Timestamp: nil,
							},
						},
						db.ModelState_ModelProgressing),
				}}, nil).MinTimes(1)
				m.EXPECT().UnlockModel("model1")
				servers := []*db.Server{
					{
						Name: "server1",
						Replicas: map[int32]*db.ServerReplica{
							0: util.NewTestServerReplica("host1", 8080, 5000, 0, store.NewServer("server1", true), []string{"sklearn"}, 0, 50, 0, nil, 100),
							1: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server1", true), []string{"sklearn"}, 0, 200, 0, nil, 100),
							2: util.NewTestServerReplica("host1", 8080, 5000, 2, store.NewServer("server1", true), []string{"sklearn"}, 0, 175, 0, nil, 100),
							3: util.NewTestServerReplica("host1", 8080, 5000, 3, store.NewServer("server1", true), []string{"sklearn"}, 0, 100, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
				}
				expectedUpdatedServers := []*db.ServerReplica{
					util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server1", true), []string{"sklearn"}, 0, 200, 0, nil, 100),
				}
				m.EXPECT().GetServers().Return(
					servers,
					nil,
				)
				m.EXPECT().UpdateLoadedModels("model1", uint32(1),
					"server1", expectedUpdatedServers).Return(nil)
			},
		},
		{
			name:      "Scale up - not enough replicas use max of the server",
			model:     newTestModel("model1", 100, []string{"sklearn"}, 5, 3, 5, []int{1, 2}, false, "server1", nil),
			modelName: "model1",
			servers: []*store.ServerSnapshot{
				{
					Name: "server1",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 100, []string{"sklearn"}, "server1", true, false), // expect schedule here
						1: gsr(1, 100, []string{"sklearn"}, "server1", true, false), // expect schedule here - nop
						2: gsr(2, 100, []string{"sklearn"}, "server1", true, false), // expect schedule here - nop
						3: gsr(3, 100, []string{"sklearn"}, "server1", true, false), // expect schedule here
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled:            true, // note that we are still trying to make the model as Available
			expectedServerEvents: 1,
			setupMock: func(m *mock.MockModelServerAPI) {
				m.EXPECT().LockModel("model1")
				m.EXPECT().GetModel("model1").Return(&db.Model{Name: "model1", Versions: []*db.ModelVersion{
					util.NewTestModelVersion(
						&pbs.Model{
							Meta: &pbs.MetaData{Name: "model1"},
							ModelSpec: &pbs.ModelSpec{
								Uri:              "",
								ArtifactVersion:  nil,
								StorageConfig:    nil,
								Requirements:     []string{"sklearn"},
								MemoryBytes:      ptr2.To(uint64(100)),
								Server:           nil,
								Parameters:       nil,
								ModelRuntimeInfo: nil,
								ModelSpec:        nil,
							},
							DeploymentSpec: &pbs.DeploymentSpec{Replicas: 5, MinReplicas: 3, MaxReplicas: 5},
						},
						1, "server1",
						map[int32]*db.ReplicaStatus{

							1: {
								State:     db.ModelReplicaState_Loaded,
								Reason:    "",
								Timestamp: nil,
							},
							2: {
								State:     db.ModelReplicaState_Loaded,
								Reason:    "",
								Timestamp: nil,
							},
						},
						db.ModelState_ModelProgressing),
				}}, nil).MinTimes(1)
				m.EXPECT().UnlockModel("model1")
				servers := []*db.Server{
					{
						Name: "server1",
						Replicas: map[int32]*db.ServerReplica{
							0: util.NewTestServerReplica("host1", 8080, 5000, 0, store.NewServer("server1", true), []string{"sklearn"}, 0, 100, 0, nil, 100),
							1: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server1", true), []string{"sklearn"}, 0, 100, 0, nil, 100),
							2: util.NewTestServerReplica("host1", 8080, 5000, 2, store.NewServer("server1", true), []string{"sklearn"}, 0, 100, 0, nil, 100),
							3: util.NewTestServerReplica("host1", 8080, 5000, 3, store.NewServer("server1", true), []string{"sklearn"}, 0, 100, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
				}

				expectedUpdatedServers := []*db.ServerReplica{
					util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server1", true), []string{"sklearn"}, 0, 100, 0, nil, 100),
					util.NewTestServerReplica("host1", 8080, 5000, 2, store.NewServer("server1", true), []string{"sklearn"}, 0, 100, 0, nil, 100),
					util.NewTestServerReplica("host1", 8080, 5000, 0, store.NewServer("server1", true), []string{"sklearn"}, 0, 100, 0, nil, 100),
					util.NewTestServerReplica("host1", 8080, 5000, 3, store.NewServer("server1", true), []string{"sklearn"}, 0, 100, 0, nil, 100),
				}
				m.EXPECT().GetServers().Return(
					servers,
					nil,
				)
				m.EXPECT().UpdateLoadedModels("model1", uint32(1),
					"server1", expectedUpdatedServers).Return(nil)
			},
		},
		{
			name:      "Scale up - no capacity on loaded replica servers, should still go there",
			model:     newTestModel("model1", 100, []string{"sklearn"}, 3, 0, 3, []int{1, 2}, false, "server1", nil),
			modelName: "model1",
			servers: []*store.ServerSnapshot{
				{
					Name: "server1",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 50, []string{"sklearn"}, "server1", true, false),
						1: gsr(1, 0, []string{"sklearn"}, "server1", true, false),   // expect schedule here - nop
						2: gsr(2, 0, []string{"sklearn"}, "server1", true, false),   // expect schedule here - nop
						3: gsr(3, 100, []string{"sklearn"}, "server1", true, false), // expect schedule here
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled: true,
			setupMock: func(m *mock.MockModelServerAPI) {
				m.EXPECT().LockModel("model1")
				m.EXPECT().GetModel("model1").Return(&db.Model{Name: "model1", Versions: []*db.ModelVersion{
					util.NewTestModelVersion(
						&pbs.Model{
							Meta: &pbs.MetaData{Name: "model1"},
							ModelSpec: &pbs.ModelSpec{
								Uri:              "",
								ArtifactVersion:  nil,
								StorageConfig:    nil,
								Requirements:     []string{"sklearn"},
								MemoryBytes:      ptr2.To(uint64(100)),
								Server:           nil,
								Parameters:       nil,
								ModelRuntimeInfo: nil,
								ModelSpec:        nil,
							},
							DeploymentSpec: &pbs.DeploymentSpec{Replicas: 3, MinReplicas: 0, MaxReplicas: 3},
						},
						1, "server1",
						map[int32]*db.ReplicaStatus{

							1: {
								State:     db.ModelReplicaState_Loaded,
								Reason:    "",
								Timestamp: nil,
							},
							2: {
								State:     db.ModelReplicaState_Loaded,
								Reason:    "",
								Timestamp: nil,
							},
						},
						db.ModelState_ModelProgressing),
				}}, nil).MinTimes(1)
				m.EXPECT().UnlockModel("model1")
				servers := []*db.Server{
					{
						Name: "server1",
						Replicas: map[int32]*db.ServerReplica{
							0: util.NewTestServerReplica("host1", 8080, 5000, 0, store.NewServer("server1", true), []string{"sklearn"}, 0, 50, 0, nil, 100),
							1: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server1", true), []string{"sklearn"}, 0, 0, 0, nil, 100),
							2: util.NewTestServerReplica("host1", 8080, 5000, 2, store.NewServer("server1", true), []string{"sklearn"}, 0, 0, 0, nil, 100),
							3: util.NewTestServerReplica("host1", 8080, 5000, 3, store.NewServer("server1", true), []string{"sklearn"}, 0, 100, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
				}

				expectedUpdatedServers := []*db.ServerReplica{
					util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server1", true), []string{"sklearn"}, 0, 0, 0, nil, 100),
					util.NewTestServerReplica("host1", 8080, 5000, 2, store.NewServer("server1", true), []string{"sklearn"}, 0, 0, 0, nil, 100),
					util.NewTestServerReplica("host1", 8080, 5000, 3, store.NewServer("server1", true), []string{"sklearn"}, 0, 100, 0, nil, 100),
				}
				m.EXPECT().GetServers().Return(
					servers,
					nil,
				)
				m.EXPECT().UpdateLoadedModels("model1", uint32(1),
					"server1", expectedUpdatedServers).Return(nil)
			},
		},
		{
			name:      "Scale down - no capacity on loaded replica servers, should still go there",
			model:     newTestModel("model1", 100, []string{"sklearn"}, 1, 0, 1, []int{1, 2}, false, "server1", nil),
			modelName: "model1",
			servers: []*store.ServerSnapshot{
				{
					Name: "server1",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 50, []string{"sklearn"}, "server1", true, false),
						1: gsr(1, 0, []string{"sklearn"}, "server1", true, false), // expect schedule here - nop
						2: gsr(2, 0, []string{"sklearn"}, "server1", true, false), // expect schedule here - nop
						3: gsr(3, 100, []string{"sklearn"}, "server1", true, false),
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled: true,
			setupMock: func(m *mock.MockModelServerAPI) {
				m.EXPECT().LockModel("model1")
				m.EXPECT().GetModel("model1").Return(&db.Model{Name: "model1", Versions: []*db.ModelVersion{
					util.NewTestModelVersion(
						&pbs.Model{
							Meta: &pbs.MetaData{Name: "model1"},
							ModelSpec: &pbs.ModelSpec{
								Uri:              "",
								ArtifactVersion:  nil,
								StorageConfig:    nil,
								Requirements:     []string{"sklearn"},
								MemoryBytes:      ptr2.To(uint64(100)),
								Server:           nil,
								Parameters:       nil,
								ModelRuntimeInfo: nil,
								ModelSpec:        nil,
							},
							DeploymentSpec: &pbs.DeploymentSpec{Replicas: 3, MinReplicas: 0, MaxReplicas: 3},
						},
						1, "server1",
						map[int32]*db.ReplicaStatus{

							1: {
								State:     db.ModelReplicaState_Loaded,
								Reason:    "",
								Timestamp: nil,
							},
							2: {
								State:     db.ModelReplicaState_Loaded,
								Reason:    "",
								Timestamp: nil,
							},
						},
						db.ModelState_ModelProgressing),
				}}, nil).MinTimes(1)
				m.EXPECT().UnlockModel("model1")
				servers := []*db.Server{
					{
						Name: "server1",
						Replicas: map[int32]*db.ServerReplica{
							0: util.NewTestServerReplica("host1", 8080, 5000, 0, store.NewServer("server1", true), []string{"sklearn"}, 0, 50, 0, nil, 100),
							1: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server1", true), []string{"sklearn"}, 0, 0, 0, nil, 100),
							2: util.NewTestServerReplica("host1", 8080, 5000, 2, store.NewServer("server1", true), []string{"sklearn"}, 0, 0, 0, nil, 100),
							3: util.NewTestServerReplica("host1", 8080, 5000, 3, store.NewServer("server1", true), []string{"sklearn"}, 0, 100, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
				}

				expectedUpdatedServers := []*db.ServerReplica{
					util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server1", true), []string{"sklearn"}, 0, 0, 0, nil, 100),
					util.NewTestServerReplica("host1", 8080, 5000, 2, store.NewServer("server1", true), []string{"sklearn"}, 0, 0, 0, nil, 100),
					util.NewTestServerReplica("host1", 8080, 5000, 3, store.NewServer("server1", true), []string{"sklearn"}, 0, 100, 0, nil, 100),
				}
				m.EXPECT().GetServers().Return(
					servers,
					nil,
				)
				m.EXPECT().UpdateLoadedModels("model1", uint32(1),
					"server1", expectedUpdatedServers).Return(nil)
			},
		},
		{
			name:      "Drain",
			model:     newTestModel("model1", 100, []string{"sklearn"}, 2, 0, 2, []int{1}, false, "server1", []int{2}),
			modelName: "model1",
			servers: []*store.ServerSnapshot{
				{
					Name: "server1",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 50, []string{"sklearn"}, "server1", true, false),
						1: gsr(1, 200, []string{"sklearn"}, "server1", true, false), // expect schedule here - nop
						2: gsr(2, 175, []string{"sklearn"}, "server1", true, true),  // drain - should not be returned
						3: gsr(3, 100, []string{"sklearn"}, "server1", true, false), // expect schedule here new replica
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled: true,
			setupMock: func(m *mock.MockModelServerAPI) {
				m.EXPECT().LockModel("model1")
				m.EXPECT().GetModel("model1").Return(&db.Model{Name: "model1", Versions: []*db.ModelVersion{
					util.NewTestModelVersion(
						&pbs.Model{
							Meta: &pbs.MetaData{Name: "model1"},
							ModelSpec: &pbs.ModelSpec{
								Uri:              "",
								ArtifactVersion:  nil,
								StorageConfig:    nil,
								Requirements:     []string{"sklearn"},
								MemoryBytes:      ptr2.To(uint64(100)),
								Server:           nil,
								Parameters:       nil,
								ModelRuntimeInfo: nil,
								ModelSpec:        nil,
							},
							DeploymentSpec: &pbs.DeploymentSpec{Replicas: 3, MinReplicas: 0, MaxReplicas: 3},
						},
						1, "server1",
						map[int32]*db.ReplicaStatus{

							1: {
								State:     db.ModelReplicaState_Loaded,
								Reason:    "",
								Timestamp: nil,
							},
							2: {
								State:     db.ModelReplicaState_Loaded,
								Reason:    "",
								Timestamp: nil,
							},
						},
						db.ModelState_ModelProgressing),
				}}, nil).MinTimes(1)
				m.EXPECT().UnlockModel("model1")
				servers := []*db.Server{
					{
						Name: "server1",
						Replicas: map[int32]*db.ServerReplica{
							0: util.NewTestServerReplica("host1", 8080, 5000, 0, store.NewServer("server1", true), []string{"sklearn"}, 0, 50, 0, nil, 100),
							1: util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server1", true), []string{"sklearn"}, 0, 200, 0, nil, 100),
							2: util.NewTestServerReplica("host1", 8080, 5000, 2, store.NewServer("server1", true), []string{"sklearn"}, 0, 175, 0, nil, 100),
							3: util.NewTestServerReplica("host1", 8080, 5000, 3, store.NewServer("server1", true), []string{"sklearn"}, 0, 100, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
				}

				expectedUpdatedServers := []*db.ServerReplica{
					util.NewTestServerReplica("host1", 8080, 5000, 1, store.NewServer("server1", true), []string{"sklearn"}, 0, 200, 0, nil, 100),
					util.NewTestServerReplica("host1", 8080, 5000, 2, store.NewServer("server1", true), []string{"sklearn"}, 0, 175, 0, nil, 100),
					util.NewTestServerReplica("host1", 8080, 5000, 3, store.NewServer("server1", true), []string{"sklearn"}, 0, 100, 0, nil, 100),
				}
				m.EXPECT().GetServers().Return(
					servers,
					nil,
				)
				m.EXPECT().UpdateLoadedModels("model1", uint32(1),
					"server1", expectedUpdatedServers).Return(nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			eventHub, _ := coordinator.NewEventHub(logger)

			ctrl := gomock.NewController(t)
			mockModelServerAPI := mock.NewMockModelServerAPI(ctrl)
			test.setupMock(mockModelServerAPI)

			serverEvents := int64(0)
			eventHub.RegisterServerEventHandler(
				"handler-server",
				10,
				logger,
				func(event coordinator.ServerEventMsg) { atomic.AddInt64(&serverEvents, 1) },
			)

			scheduler := NewSimpleScheduler(logger, mockModelServerAPI, DefaultSchedulerConfig(mockModelServerAPI), synchroniser.NewSimpleSynchroniser(time.Duration(10*time.Millisecond)), eventHub)
			err := scheduler.Schedule(test.modelName)
			if test.scheduled {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).ToNot(BeNil())
			}
			if test.expectedServerEvents > 0 { // wait for event
				time.Sleep(500 * time.Millisecond)
			}

			if test.checkServerEvents {
				g.Expect(atomic.LoadInt64(&serverEvents)).To(Equal(int64(test.expectedServerEvents)))
			}
		})
	}
}

func TestFailedModels(t *testing.T) {
	logger := log.New()
	g := NewGomegaWithT(t)

	type test struct {
		name                 string
		models               []*db.Model
		setupMock            func(m *mock.MockModelServerAPI)
		expectedFailedModels []string
	}

	tests := []test{
		{
			name: "SmokeTest",
			setupMock: func(m *mock.MockModelServerAPI) {

				model3 := util.NewTestModelVersion(
					&pbs.Model{
						Meta: &pbs.MetaData{Name: "model3"},
						ModelSpec: &pbs.ModelSpec{
							Uri:              "",
							ArtifactVersion:  nil,
							StorageConfig:    nil,
							Requirements:     []string{"sklearn"},
							MemoryBytes:      ptr2.To(uint64(100)),
							Server:           nil,
							Parameters:       nil,
							ModelRuntimeInfo: nil,
							ModelSpec:        nil,
						},
						DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1, MinReplicas: 0, MaxReplicas: 1},
					},
					1, "server1",
					map[int32]*db.ReplicaStatus{
						1: {
							State:     db.ModelReplicaState_Loaded,
							Reason:    "",
							Timestamp: nil,
						},
					},
					db.ModelState_ModelAvailable)

				// set available replicas
				model3.State.AvailableReplicas = 1

				model4 := util.NewTestModelVersion(
					&pbs.Model{
						Meta: &pbs.MetaData{Name: "model4"},
						ModelSpec: &pbs.ModelSpec{
							Uri:              "",
							ArtifactVersion:  nil,
							StorageConfig:    nil,
							Requirements:     []string{"sklearn"},
							MemoryBytes:      ptr2.To(uint64(100)),
							Server:           nil,
							Parameters:       nil,
							ModelRuntimeInfo: nil,
							ModelSpec:        nil,
						},
						DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MinReplicas: 1, MaxReplicas: 2},
					},
					1, "server1",
					map[int32]*db.ReplicaStatus{
						1: {
							State:     db.ModelReplicaState_Loaded,
							Reason:    "",
							Timestamp: nil,
						},
					},
					db.ModelState_ModelAvailable)

				// set available replicas
				model4.State.AvailableReplicas = 1

				m.EXPECT().GetModels().Return(
					[]*db.Model{
						{
							Name: "model1",
							Versions: []*db.ModelVersion{util.NewTestModelVersion(
								&pbs.Model{
									Meta: &pbs.MetaData{Name: "model1"},
									ModelSpec: &pbs.ModelSpec{
										Uri:              "",
										ArtifactVersion:  nil,
										StorageConfig:    nil,
										Requirements:     []string{"sklearn"},
										MemoryBytes:      ptr2.To(uint64(100)),
										Server:           nil,
										Parameters:       nil,
										ModelRuntimeInfo: nil,
										ModelSpec:        nil,
									},
									DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1, MinReplicas: 0, MaxReplicas: 1},
								},
								1, "server1",
								nil,
								db.ModelState_ScheduleFailed),
							},
						},
						{
							Name: "model2",
							Versions: []*db.ModelVersion{util.NewTestModelVersion(
								&pbs.Model{
									Meta: &pbs.MetaData{Name: "model2"},
									ModelSpec: &pbs.ModelSpec{
										Uri:              "",
										ArtifactVersion:  nil,
										StorageConfig:    nil,
										Requirements:     []string{"sklearn"},
										MemoryBytes:      ptr2.To(uint64(100)),
										Server:           nil,
										Parameters:       nil,
										ModelRuntimeInfo: nil,
										ModelSpec:        nil,
									},
									DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1, MinReplicas: 0, MaxReplicas: 1},
								},
								1, "server1",
								nil,
								db.ModelState_ModelFailed),
							},
						},
						{
							Name: "model3",
							Versions: []*db.ModelVersion{
								model3,
							},
						},
						{
							Name: "model4",
							Versions: []*db.ModelVersion{
								model4,
							},
						}}, nil)

			},
			expectedFailedModels: []string{"model1", "model2", "model4"},
		},
		{
			name: "SmokeTest",
			setupMock: func(m *mock.MockModelServerAPI) {

				model3 := util.NewTestModelVersion(
					&pbs.Model{
						Meta: &pbs.MetaData{Name: "model3"},
						ModelSpec: &pbs.ModelSpec{
							Uri:              "",
							ArtifactVersion:  nil,
							StorageConfig:    nil,
							Requirements:     []string{"sklearn"},
							MemoryBytes:      ptr2.To(uint64(100)),
							Server:           nil,
							Parameters:       nil,
							ModelRuntimeInfo: nil,
							ModelSpec:        nil,
						},
						DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1, MinReplicas: 0, MaxReplicas: 1},
					},
					1, "server1",
					map[int32]*db.ReplicaStatus{
						1: {
							State:     db.ModelReplicaState_Loaded,
							Reason:    "",
							Timestamp: nil,
						},
					},
					db.ModelState_ModelAvailable)

				// set available replicas
				model3.State.AvailableReplicas = 1

				m.EXPECT().GetModels().Return(
					[]*db.Model{

						{
							Name: "model3",
							Versions: []*db.ModelVersion{
								model3,
							},
						}}, nil)

			},
			expectedFailedModels: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			eventHub, _ := coordinator.NewEventHub(logger)

			ctrl := gomock.NewController(t)
			mockModelServerAPI := mock.NewMockModelServerAPI(ctrl)
			test.setupMock(mockModelServerAPI)

			scheduler := NewSimpleScheduler(logger, mockModelServerAPI, DefaultSchedulerConfig(mockModelServerAPI), synchroniser.NewSimpleSynchroniser(10*time.Millisecond), eventHub)
			failedModels, err := scheduler.getFailedModels()
			g.Expect(err).To(BeNil())
			sort.Strings(failedModels)
			sort.Strings(test.expectedFailedModels)
			g.Expect(failedModels).To(Equal(test.expectedFailedModels))
		})
	}
}

func TestRemoveAllVersions(t *testing.T) {
	logger := log.New()
	g := NewGomegaWithT(t)

	newTestModel := func(name string, requirements []string, loadedModels []int, scheduledServer string, numVersions int) *store.ModelSnapshot {
		config := &pb.Model{Meta: &pb.MetaData{Name: t.Name()}, ModelSpec: &pb.ModelSpec{Requirements: requirements}}
		rmap := make(map[int]store.ReplicaStatus)
		for _, ridx := range loadedModels {
			rmap[ridx] = store.ReplicaStatus{State: store.Loaded}
		}

		versions := []*store.ModelVersion{}
		for i := 1; i <= numVersions; i++ {
			versions = append(versions, store.NewModelVersion(config, uint32(i), scheduledServer, rmap, false, store.ModelAvailable))
		}
		// load a bad version - this should not get unloaded by the test
		versions = append(versions, store.NewModelVersion(config, uint32(numVersions+1), scheduledServer, map[int]store.ReplicaStatus{}, false, store.ScheduleFailed))

		return &store.ModelSnapshot{
			Name:     name,
			Versions: versions,
			Deleted:  true,
		}
	}

	gsr := func(replicaIdx int, availableMemory uint64, capabilities []string, serverName string) *store.ServerReplica {
		replica := store.NewServerReplica("svc", 8080, 5001, replicaIdx, store.NewServer(serverName, true), capabilities, availableMemory, availableMemory, 0, nil, 100)
		return replica
	}

	//newMockStore := func(model *store.ModelSnapshot, servers []*store.ServerSnapshot) *mockStore {
	//	modelMap := make(map[string]*store.ModelSnapshot)
	//	modelMap[model.Name] = model
	//	return &mockStore{
	//		models:         modelMap,
	//		servers:        servers,
	//		unloadedModels: make(map[string]uint32),
	//	}
	//}

	type test struct {
		name        string
		model       *store.ModelSnapshot
		servers     []*store.ServerSnapshot
		numVersions int
	}

	tests := []test{
		{
			name:  "Allversions - 1",
			model: newTestModel("model1", []string{"sklearn"}, []int{0, 1}, "server", 1),
			servers: []*store.ServerSnapshot{
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}, "server"),
						1: gsr(1, 200, []string{"sklearn"}, "server"),
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			numVersions: 1,
		},
		{
			name:  "Allversions - > 1",
			model: newTestModel("model1", []string{"sklearn"}, []int{0, 1}, "server", 10),
			servers: []*store.ServerSnapshot{
				{
					Name: "server",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}, "server"),
						1: gsr(1, 200, []string{"sklearn"}, "server"),
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			numVersions: 10,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _ = coordinator.NewEventHub(logger)

			ctrl := gomock.NewController(t)
			mockModelServerAPI := mock.NewMockModelServerAPI(ctrl)
			//test.setupMock(mockModelServerAPI)

			//mockStore := newMockStore(test.model, test.servers)
			scheduler := NewSimpleScheduler(logger, mockModelServerAPI, DefaultSchedulerConfig(mockModelServerAPI), synchroniser.NewSimpleSynchroniser(time.Duration(10*time.Millisecond)), nil)
			err := scheduler.Schedule(test.model.Name)
			g.Expect(err).To(BeNil())

			//g.Expect(mockStore.unloadedModels[test.model.Name]).To(Equal(uint32(test.numVersions)))
		})
	}
}

func TestScheduleFailedModels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupMocks     func(*mock.MockModelServerAPI, *mock2.MockSynchroniser)
		expectedModels []string
		expectError    bool
		errorContains  string
	}{
		{
			name: "success - schedules single failed model",
			setupMocks: func(ms *mock.MockModelServerAPI, sync *mock2.MockSynchroniser) {
				sync.EXPECT().IsReady().Return(true)

				model1 := &db.Model{
					Name: "model1",
					Versions: []*db.ModelVersion{
						util.NewTestModelVersion(
							&pbs.Model{
								Meta: &pbs.MetaData{Name: "model1"},
								ModelSpec: &pbs.ModelSpec{
									Uri:              "",
									ArtifactVersion:  nil,
									StorageConfig:    nil,
									Requirements:     nil,
									MemoryBytes:      nil,
									Server:           ptr.String("server1"),
									Parameters:       nil,
									ModelRuntimeInfo: nil,
									ModelSpec:        nil,
								},
								DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1, MinReplicas: 0, MaxReplicas: 0},
							},
							1, "server1",
							nil,
							db.ModelState_ScheduleFailed)},
				}

				ms.EXPECT().GetModels().Return([]*db.Model{model1}, nil)

				ms.EXPECT().LockModel("model1")
				ms.EXPECT().UnlockModel("model1")
				ms.EXPECT().GetModel("model1").Return(model1, nil)

				servers := []*db.Server{
					{
						Name: "server1",
						Replicas: map[int32]*db.ServerReplica{
							0: util.NewTestServerReplica("host1", 8080, 5000, 0, store.NewServer("server1", true), []string{"sklearn"}, 0, 16000, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
				}

				expectedServerUpdate := []*db.ServerReplica{
					util.NewTestServerReplica("host1", 8080, 5000, 0, store.NewServer("server1", true), []string{"sklearn"}, 0, 16000, 0, nil, 100),
				}

				ms.EXPECT().GetServers().Return(servers, nil)
				ms.EXPECT().UpdateLoadedModels("model1", uint32(1),
					"server1", expectedServerUpdate).Return(nil)
			},
			expectedModels: []string{"model1"},
			expectError:    false,
		},
		{
			name: "success - schedules 2 failed models",
			setupMocks: func(ms *mock.MockModelServerAPI, sync *mock2.MockSynchroniser) {
				sync.EXPECT().IsReady().Return(true)

				model1 := &db.Model{
					Name: "model1",
					Versions: []*db.ModelVersion{
						util.NewTestModelVersion(
							&pbs.Model{
								Meta: &pbs.MetaData{Name: "model1"},
								ModelSpec: &pbs.ModelSpec{
									Uri:              "",
									ArtifactVersion:  nil,
									StorageConfig:    nil,
									Requirements:     nil,
									MemoryBytes:      nil,
									Server:           ptr.String("server1"),
									Parameters:       nil,
									ModelRuntimeInfo: nil,
									ModelSpec:        nil,
								},
								DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1, MinReplicas: 0, MaxReplicas: 0},
							},
							1, "server1",
							nil,
							db.ModelState_ScheduleFailed)},
				}

				model2 := &db.Model{
					Name: "model2",
					Versions: []*db.ModelVersion{
						util.NewTestModelVersion(
							&pbs.Model{
								Meta: &pbs.MetaData{Name: "model2"},
								ModelSpec: &pbs.ModelSpec{
									Uri:              "",
									ArtifactVersion:  nil,
									StorageConfig:    nil,
									Requirements:     nil,
									MemoryBytes:      nil,
									Server:           ptr.String("server1"),
									Parameters:       nil,
									ModelRuntimeInfo: nil,
									ModelSpec:        nil,
								},
								DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1, MinReplicas: 0, MaxReplicas: 0},
							},
							1, "server1",
							nil,
							db.ModelState_ScheduleFailed)},
				}

				ms.EXPECT().GetModels().Return([]*db.Model{model1, model2}, nil)

				// model1
				ms.EXPECT().LockModel("model1")
				ms.EXPECT().UnlockModel("model1")
				ms.EXPECT().GetModel("model1").Return(model1, nil)

				servers := []*db.Server{
					{
						Name: "server1",
						Replicas: map[int32]*db.ServerReplica{
							0: util.NewTestServerReplica("host1", 8080, 5000, 0, store.NewServer("server1", true), []string{"sklearn"}, 0, 16000, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
				}

				expectedUpdatedServers := []*db.ServerReplica{
					util.NewTestServerReplica("host1", 8080, 5000, 0, store.NewServer("server1", true), []string{"sklearn"}, 0, 16000, 0, nil, 100),
				}
				ms.EXPECT().GetServers().Return(servers, nil)
				ms.EXPECT().UpdateLoadedModels("model1", uint32(1),
					"server1", expectedUpdatedServers).Return(nil)

				// model2

				ms.EXPECT().LockModel("model2")
				ms.EXPECT().UnlockModel("model2")
				ms.EXPECT().GetModel("model2").Return(model2, nil)

				ms.EXPECT().GetServers().Return(servers, nil)
				ms.EXPECT().UpdateLoadedModels("model2", uint32(1),
					"server1", expectedUpdatedServers).Return(nil)
			},
			expectedModels: []string{"model1", "model2"},
			expectError:    false,
		},
		{
			name: "failure - unable to schedule model on desired replicas or min replicas",
			setupMocks: func(ms *mock.MockModelServerAPI, sync *mock2.MockSynchroniser) {
				sync.EXPECT().IsReady().Return(true)

				model11 := &db.Model{
					Name: "model1",
					Versions: []*db.ModelVersion{
						util.NewTestModelVersion(
							&pbs.Model{
								Meta: &pbs.MetaData{Name: "model1"},
								ModelSpec: &pbs.ModelSpec{
									Uri:              "",
									ArtifactVersion:  nil,
									StorageConfig:    nil,
									Requirements:     nil,
									MemoryBytes:      nil,
									Server:           ptr.String("server1"),
									Parameters:       nil,
									ModelRuntimeInfo: nil,
									ModelSpec:        nil,
								},
								DeploymentSpec: &pbs.DeploymentSpec{Replicas: 3, MinReplicas: 2, MaxReplicas: 0},
							},
							1, "server1",
							nil,
							db.ModelState_ScheduleFailed)},
				}

				ms.EXPECT().GetModels().Return([]*db.Model{model11}, nil)

				ms.EXPECT().LockModel("model1")
				ms.EXPECT().UnlockModel("model1")
				ms.EXPECT().GetModel("model1").Return(model11, nil)

				serverss := []*db.Server{
					{
						Name: "server1",
						Replicas: map[int32]*db.ServerReplica{
							0: util.NewTestServerReplica("host1", 8080, 5000, 0, store.NewServer("server1", true), []string{"sklearn"}, 0, 16000, 0, nil, 100),
						},
						Shared:           true,
						ExpectedReplicas: -1,
						KubernetesMeta:   nil,
					},
				}
				ms.EXPECT().GetServers().Return(serverss, nil)
				ms.EXPECT().FailedScheduling("model1", uint32(1),
					"Failed to schedule model as no matching server had enough suitable replicas", true).Return(nil)
			},
			expectedModels: []string{},
		},
		{
			name: "failure - failed getting models",
			setupMocks: func(ms *mock.MockModelServerAPI, sync *mock2.MockSynchroniser) {
				sync.EXPECT().IsReady().Return(true)
				ms.EXPECT().GetModels().Return(nil, errors.New("some error"))
			},
			expectedModels: []string{},
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockModelServerAPI := mock.NewMockModelServerAPI(ctrl)
			mockSync := mock2.NewMockSynchroniser(ctrl)

			tt.setupMocks(mockModelServerAPI, mockSync)

			eventHub, err := coordinator.NewEventHub(log.New())
			require.NoError(t, err)

			scheduler := NewSimpleScheduler(
				log.New(),
				mockModelServerAPI,
				DefaultSchedulerConfig(mockModelServerAPI),
				mockSync,
				eventHub)

			updatedModels, err := scheduler.ScheduleFailedModels()

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			assert.ElementsMatch(t, tt.expectedModels, updatedModels)
		})
	}
}
