/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package store

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
)

func TestUpdateModel(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name            string
		store           *LocalSchedulerStore
		loadModelReq    *pb.LoadModelRequest
		expectedVersion uint32
		err             error
	}

	tests := []test{
		{
			name:  "simple",
			store: NewLocalSchedulerStore(),
			loadModelReq: &pb.LoadModelRequest{
				Model: &pb.Model{
					Meta: &pb.MetaData{
						Name: "model",
					},
				},
			},
			expectedVersion: 1,
		},
		{
			name: "VersionAlreadyExists",
			store: &LocalSchedulerStore{
				models: map[string]*Model{
					"model": {
						versions: []*ModelVersion{
							{
								version: 1,
								modelDefn: &pb.Model{
									Meta: &pb.MetaData{
										Name: "model",
									},
								},
							},
						},
					},
				},
			},
			loadModelReq: &pb.LoadModelRequest{
				Model: &pb.Model{
					Meta: &pb.MetaData{
						Name: "model",
					},
				},
			},
			expectedVersion: 1,
		},
		{
			name: "Meta data is changed - no new version created assuming same name of model",
			store: &LocalSchedulerStore{
				models: map[string]*Model{
					"model": {
						versions: []*ModelVersion{
							{
								version: 1,
								modelDefn: &pb.Model{
									Meta: &pb.MetaData{
										Name: "model",
									},
								},
							},
						},
					},
				},
			},
			loadModelReq: &pb.LoadModelRequest{
				Model: &pb.Model{
					Meta: &pb.MetaData{
						Name: "model",
						KubernetesMeta: &pb.KubernetesMeta{
							Generation: 2,
						},
					},
				},
			},
			expectedVersion: 1,
		},
		{
			name: "DeploymentSpecDiffers",
			store: &LocalSchedulerStore{
				models: map[string]*Model{
					"model": {
						versions: []*ModelVersion{
							{
								version: 1,
								modelDefn: &pb.Model{
									Meta: &pb.MetaData{
										Name: "model",
									},
									ModelSpec: &pb.ModelSpec{
										Uri: "gs:/models/iris",
									},
									DeploymentSpec: &pb.DeploymentSpec{
										Replicas: 2,
									},
								},
							},
						},
					},
				},
			},
			loadModelReq: &pb.LoadModelRequest{
				Model: &pb.Model{
					Meta: &pb.MetaData{
						Name: "model",
					},
					ModelSpec: &pb.ModelSpec{
						Uri: "gs:/models/iris",
					},
					DeploymentSpec: &pb.DeploymentSpec{
						Replicas: 4,
					},
				},
			},
			expectedVersion: 1,
		},
		{
			name: "ModelSpecDiffers",
			store: &LocalSchedulerStore{
				models: map[string]*Model{
					"model": {
						versions: []*ModelVersion{
							{
								version: 1,
								modelDefn: &pb.Model{
									Meta: &pb.MetaData{
										Name: "model",
									},
									ModelSpec: &pb.ModelSpec{
										Uri: "gs:/models/iris",
									},
									DeploymentSpec: &pb.DeploymentSpec{
										Replicas: 2,
									},
								},
							},
						},
					},
				},
			},
			loadModelReq: &pb.LoadModelRequest{
				Model: &pb.Model{
					Meta: &pb.MetaData{
						Name: "model",
					},
					ModelSpec: &pb.ModelSpec{
						Uri: "gs:/models/iris2",
					},
					DeploymentSpec: &pb.DeploymentSpec{
						Replicas: 4,
					},
				},
			},
			expectedVersion: 2,
		},
		{
			name:  "ModelNameIsNotValid",
			store: NewLocalSchedulerStore(),
			loadModelReq: &pb.LoadModelRequest{
				Model: &pb.Model{
					Meta: &pb.MetaData{
						Name: "this.Name",
					},
				},
			},
			expectedVersion: 1,
			err:             errors.New("Model this.Name does not have a valid name - it must be alphanumeric and not contains dots (.)"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())
			ms := NewMemoryStore(logger, test.store, eventHub)
			err = ms.UpdateModel(test.loadModelReq)
			if test.err != nil {
				g.Expect(err.Error()).To(BeIdenticalTo(test.err.Error()))
			} else {
				g.Expect(err).To(BeNil())
				m := test.store.models[test.loadModelReq.GetModel().GetMeta().GetName()]
				latest := m.Latest()
				g.Expect(latest.modelDefn).To(Equal(test.loadModelReq.Model))
				g.Expect(latest.GetVersion()).To(Equal(test.expectedVersion))
			}
		})
	}
}

func TestGetModel(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		store    *LocalSchedulerStore
		key      string
		versions int
		err      error
	}

	tests := []test{
		{
			name:     "NoModel",
			store:    NewLocalSchedulerStore(),
			key:      "model",
			versions: 0,
			err:      nil,
		},
		{
			name: "VersionAlreadyExists",
			store: &LocalSchedulerStore{
				models: map[string]*Model{
					"model": {
						versions: []*ModelVersion{
							{
								modelDefn: &pb.Model{
									Meta: &pb.MetaData{
										Name: "model",
									},
								},
							},
						},
					},
				},
			},
			key:      "model",
			versions: 1,
			err:      nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())
			ms := NewMemoryStore(logger, test.store, eventHub)
			model, err := ms.GetModel(test.key)
			if test.err == nil {
				g.Expect(err).To(BeNil())
				g.Expect(model.Name).To(Equal(test.key))
				g.Expect(len(model.Versions)).To(Equal(test.versions))
			} else {
				g.Expect(err).ToNot(BeNil())
			}
		})
	}
}

func TestRemoveModel(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name  string
		store *LocalSchedulerStore
		key   string
		err   bool
	}

	tests := []test{
		{
			name:  "NoModel",
			store: NewLocalSchedulerStore(),
			key:   "model",
			err:   true,
		},
		{
			name: "VersionAlreadyExists",
			store: &LocalSchedulerStore{
				models: map[string]*Model{
					"model": {
						versions: []*ModelVersion{
							{
								modelDefn: &pb.Model{
									Meta: &pb.MetaData{
										Name: "model",
									},
								},
							},
						},
					},
				},
			},
			key: "model",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())
			ms := NewMemoryStore(logger, test.store, eventHub)
			err = ms.RemoveModel(&pb.UnloadModelRequest{Model: &pb.ModelReference{Name: test.key}})
			if !test.err {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).ToNot(BeNil())
			}
		})
	}
}

func TestUpdateLoadedModels(t *testing.T) {
	g := NewGomegaWithT(t)
	memBytes := uint64(1)

	type test struct {
		name               string
		store              *LocalSchedulerStore
		modelKey           string
		version            uint32
		serverKey          string
		replicas           []*ServerReplica
		expectedStates     map[int]ReplicaStatus
		err                bool
		isModelDeleted     bool
		expectedModelState *ModelStatus
	}

	tests := []test{
		{
			name: "ModelVersionNotLatest",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"model": {
					versions: []*ModelVersion{
						{
							modelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
							server:    "server",
							version:   1,
							replicas:  map[int]ReplicaStatus{},
						},
						{
							modelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
							server:    "server",
							version:   2,
							replicas:  map[int]ReplicaStatus{},
						},
					},
				}},
				servers: map[string]*Server{
					"server": {
						name: "server",
					},
				},
			},
			modelKey:  "model",
			version:   1,
			serverKey: "server",
			replicas:  nil,
			err:       true,
		},
		{
			name: "UpdatedVersionsOK",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"model": {
					versions: []*ModelVersion{
						{
							modelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
							server:    "server",
							version:   1,
							replicas:  map[int]ReplicaStatus{},
						},
					},
				}},
				servers: map[string]*Server{
					"server": {
						name: "server",
						replicas: map[int]*ServerReplica{
							0: {},
							1: {},
						},
					},
				},
			},
			modelKey:  "model",
			version:   1,
			serverKey: "server",
			replicas: []*ServerReplica{
				{replicaIdx: 0}, {replicaIdx: 1},
			},
			expectedStates: map[int]ReplicaStatus{0: {State: LoadRequested}, 1: {State: LoadRequested}},
		},
		{
			name: "WithAlreadyLoadedModels",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"model": {
					versions: []*ModelVersion{
						{
							modelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
							server:    "server",
							version:   1,
							replicas: map[int]ReplicaStatus{
								0: {State: Loaded},
							},
						},
					},
				}},
				servers: map[string]*Server{
					"server": {
						name: "server",
						replicas: map[int]*ServerReplica{
							0: {},
							1: {},
						},
					},
				},
			},
			modelKey:  "model",
			version:   1,
			serverKey: "server",
			replicas: []*ServerReplica{
				{replicaIdx: 0}, {replicaIdx: 1},
			},
			expectedStates: map[int]ReplicaStatus{0: {State: Loaded}, 1: {State: LoadRequested}},
		},
		{
			name: "UnloadModelsNotSelected",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"model": {
					versions: []*ModelVersion{
						{
							modelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
							server:    "server",
							version:   1,
							replicas: map[int]ReplicaStatus{
								0: {State: Loaded},
							},
						},
					},
				}},
				servers: map[string]*Server{
					"server": {
						name: "server",
						replicas: map[int]*ServerReplica{
							0: {},
							1: {},
						},
					},
				},
			},
			modelKey:  "model",
			version:   1,
			serverKey: "server",
			replicas: []*ServerReplica{
				{replicaIdx: 1},
			},
			expectedStates: map[int]ReplicaStatus{0: {State: UnloadEnvoyRequested}, 1: {State: LoadRequested}},
		},
		{
			name: "DeletedModel",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"model": {
					versions: []*ModelVersion{
						{
							modelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
							server:    "server",
							version:   1,
							replicas: map[int]ReplicaStatus{
								0: {State: Loaded},
								1: {State: Loading},
							},
						},
					},
				}},
				servers: map[string]*Server{
					"server": {
						name: "server",
						replicas: map[int]*ServerReplica{
							0: {},
							1: {},
						},
					},
				},
			},
			modelKey:       "model",
			version:        1,
			serverKey:      "server",
			replicas:       []*ServerReplica{},
			isModelDeleted: true,
			expectedStates: map[int]ReplicaStatus{0: {State: UnloadEnvoyRequested}, 1: {State: UnloadEnvoyRequested}},
		},
		{
			name: "DeletedModelNoReplicas",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"model": {
					versions: []*ModelVersion{
						{
							modelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
							server:    "server",
							version:   1,
							replicas: map[int]ReplicaStatus{
								0: {State: Unloaded},
							},
						},
					},
				}},
				servers: map[string]*Server{
					"server": {
						name: "server",
						replicas: map[int]*ServerReplica{
							0: {},
							1: {},
						},
					},
				},
			},
			modelKey:       "model",
			version:        1,
			serverKey:      "server",
			replicas:       []*ServerReplica{},
			isModelDeleted: true,
			expectedStates: map[int]ReplicaStatus{0: {State: Unloaded}},
		},
		{
			name: "ServerChanged",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"model": {
					versions: []*ModelVersion{
						{
							modelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
							server:    "server1",
							version:   1,
							replicas:  map[int]ReplicaStatus{},
						},
					},
				}},
				servers: map[string]*Server{
					"server1": {
						name: "server1",
						replicas: map[int]*ServerReplica{
							0: {},
							1: {},
						},
					},
					"server2": {
						name: "server2",
						replicas: map[int]*ServerReplica{
							0: {},
							1: {},
						},
					},
				},
			},
			modelKey:  "model",
			version:   1,
			serverKey: "server2",
			replicas: []*ServerReplica{
				{replicaIdx: 0}, {replicaIdx: 1},
			},
			expectedStates: map[int]ReplicaStatus{0: {State: LoadRequested}, 1: {State: LoadRequested}},
		},
		{
			name: "WithDrainingServerReplicaSameServer",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"model": {
					versions: []*ModelVersion{
						{
							modelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
							server:    "server",
							version:   1,
							replicas: map[int]ReplicaStatus{
								0: {State: Draining},
							},
						},
					},
				}},
				servers: map[string]*Server{
					"server": {
						name: "server",
						replicas: map[int]*ServerReplica{
							0: {isDraining: true},
							1: {},
						},
					},
				},
			},
			modelKey:  "model",
			version:   1,
			serverKey: "server",
			replicas: []*ServerReplica{
				{replicaIdx: 1},
			},
			expectedStates: map[int]ReplicaStatus{0: {State: Draining}, 1: {State: LoadRequested}},
		},
		{
			name: "WithDrainingServerReplicaNewServer",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"model": {
					versions: []*ModelVersion{
						{
							modelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
							server:    "server1",
							version:   1,
							replicas: map[int]ReplicaStatus{
								0: {State: Draining},
							},
						},
					},
				}},
				servers: map[string]*Server{
					"server1": {
						name: "server1",
						replicas: map[int]*ServerReplica{
							0: {isDraining: true},
						},
					},
					"server2": {
						name: "server2",
						replicas: map[int]*ServerReplica{
							0: {},
						},
					},
				},
			},
			modelKey:  "model",
			version:   1,
			serverKey: "server2",
			replicas: []*ServerReplica{
				{replicaIdx: 0},
			},
			expectedStates: map[int]ReplicaStatus{0: {State: LoadRequested}},
		},
		{
			name: "DeleteFailedScheduleModel",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"model": {
					versions: []*ModelVersion{
						{
							modelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
							server:    "",
							version:   1,
							replicas:  map[int]ReplicaStatus{},
						},
					},
				}},
				servers: map[string]*Server{},
			},
			modelKey:       "model",
			version:        1,
			serverKey:      "",
			replicas:       []*ServerReplica{},
			isModelDeleted: true,
			expectedStates: map[int]ReplicaStatus{},
		},
		{
			name: "ProgressModelLoading",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"model": {
					versions: []*ModelVersion{
						{
							modelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
							server:    "server",
							version:   1,
							replicas: map[int]ReplicaStatus{
								0: {State: Available},
								1: {State: Unloaded},
							},
							state: ModelStatus{State: ModelProgressing},
						},
					},
				}},
				servers: map[string]*Server{
					"server": {
						name: "server",
						replicas: map[int]*ServerReplica{
							0: {},
							1: {},
						},
					},
				},
			},
			modelKey:  "model",
			version:   1,
			serverKey: "server",
			replicas: []*ServerReplica{
				{replicaIdx: 0},
			},
			expectedStates:     map[int]ReplicaStatus{0: {State: Available}, 1: {State: Unloaded}},
			expectedModelState: &ModelStatus{State: ModelAvailable},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())
			if test.isModelDeleted {
				test.store.models[test.modelKey].SetDeleted()
			}
			ms := NewMemoryStore(logger, test.store, eventHub)
			msg, err := ms.updateLoadedModelsImpl(test.modelKey, test.version, test.serverKey, test.replicas)
			if !test.err {
				g.Expect(err).To(BeNil())
				g.Expect(msg).ToNot(BeNil())
				mv := test.store.models[test.modelKey].Latest()
				for replicaIdx, state := range test.expectedStates {
					g.Expect(mv).ToNot(BeNil())
					g.Expect(mv.GetModelReplicaState(replicaIdx)).To(Equal(state.State))
					ss, _ := ms.GetServer(test.serverKey, false, true)
					if state.State == LoadRequested {
						g.Expect(ss.Replicas[replicaIdx].GetReservedMemory()).To(Equal(memBytes))
					} else {
						g.Expect(ss.Replicas[replicaIdx].GetReservedMemory()).To(Equal(uint64(0)))
					}
				}
				if test.expectedModelState != nil {
					g.Expect(mv.state.State).To(Equal(test.expectedModelState.State))
				}
			} else {
				g.Expect(err).ToNot(BeNil())
				g.Expect(msg).To(BeNil())
			}
		})
	}
}

func TestUpdateModelState(t *testing.T) {
	g := NewGomegaWithT(t)
	memBytes := uint64(1)

	type test struct {
		name                   string
		store                  *LocalSchedulerStore
		modelKey               string
		version                uint32
		serverKey              string
		replicaIdx             int
		expectedState          ModelReplicaState
		desiredState           ModelReplicaState
		availableMemory        uint64
		numModelVersionsLoaded int
		modelVersionLoaded     bool
		deleted                bool
		err                    bool
	}

	tests := []test{
		{
			name: "LoadedModel",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"model": {
					versions: []*ModelVersion{
						{
							modelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
							version:   1,
							replicas:  map[int]ReplicaStatus{},
						},
					},
				}},
				servers: map[string]*Server{
					"server": {
						name: "server",
						replicas: map[int]*ServerReplica{
							0: {loadedModels: map[ModelVersionID]bool{}, reservedMemory: memBytes, uniqueLoadedModels: map[string]bool{}},
							1: {loadedModels: map[ModelVersionID]bool{}, reservedMemory: memBytes, uniqueLoadedModels: map[string]bool{}},
						},
					},
				},
			},
			modelKey:               "model",
			version:                1,
			serverKey:              "server",
			replicaIdx:             0,
			expectedState:          ModelReplicaStateUnknown,
			desiredState:           Loaded,
			numModelVersionsLoaded: 1,
			modelVersionLoaded:     true,
			availableMemory:        20,
		},
		{
			name: "UnloadedModel",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"model": {
					versions: []*ModelVersion{
						{
							modelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
							version:   1,
							replicas:  map[int]ReplicaStatus{},
						},
					},
				}},
				servers: map[string]*Server{
					"server": {
						name: "server",
						replicas: map[int]*ServerReplica{
							0: {loadedModels: map[ModelVersionID]bool{}, reservedMemory: memBytes, uniqueLoadedModels: map[string]bool{}},
							1: {loadedModels: map[ModelVersionID]bool{}, reservedMemory: memBytes, uniqueLoadedModels: map[string]bool{}},
						},
					},
				},
			},
			modelKey:               "model",
			version:                1,
			serverKey:              "server",
			replicaIdx:             0,
			expectedState:          ModelReplicaStateUnknown,
			desiredState:           Unloaded,
			numModelVersionsLoaded: 0,
			modelVersionLoaded:     false,
			availableMemory:        20,
		},
		{
			name: "Unloaded model but not matching expected state",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"model": {
					versions: []*ModelVersion{
						{
							modelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
							version:   1,
							replicas: map[int]ReplicaStatus{
								0: {State: LoadRequested},
							},
						},
					},
				}},
				servers: map[string]*Server{
					"server": {
						name: "server",
						replicas: map[int]*ServerReplica{
							0: {loadedModels: map[ModelVersionID]bool{}, reservedMemory: memBytes, uniqueLoadedModels: map[string]bool{}},
							1: {loadedModels: map[ModelVersionID]bool{}, reservedMemory: memBytes, uniqueLoadedModels: map[string]bool{}},
						},
					},
				},
			},
			modelKey:               "model",
			version:                1,
			serverKey:              "server",
			replicaIdx:             0,
			expectedState:          Unloading,
			desiredState:           Unloaded,
			numModelVersionsLoaded: 0,
			modelVersionLoaded:     false,
			availableMemory:        20,
			err:                    true,
		},
		{
			name: "DeletedModel",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"model": {
					versions: []*ModelVersion{
						{
							modelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
							version:   1,
							replicas:  map[int]ReplicaStatus{},
						},
					},
				}},
				servers: map[string]*Server{
					"server": {
						name: "server",
						replicas: map[int]*ServerReplica{
							0: {loadedModels: map[ModelVersionID]bool{}, reservedMemory: memBytes, uniqueLoadedModels: map[string]bool{}},
							1: {loadedModels: map[ModelVersionID]bool{}, reservedMemory: memBytes, uniqueLoadedModels: map[string]bool{}},
						},
					},
				},
			},
			modelKey:               "model",
			version:                1,
			serverKey:              "server",
			replicaIdx:             0,
			expectedState:          ModelReplicaStateUnknown,
			desiredState:           Unloaded,
			numModelVersionsLoaded: 0,
			modelVersionLoaded:     false,
			availableMemory:        20,
			deleted:                true,
		},
		{
			name: "Model updated but not latest on replica which is loaded",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"foo": {
					versions: []*ModelVersion{
						{
							modelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
							version:   1,
							replicas: map[int]ReplicaStatus{
								0: {State: Unloading},
							},
						},
						{
							modelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
							version:   2,
							replicas: map[int]ReplicaStatus{
								0: {State: Loaded},
							},
						},
					},
				}},
				servers: map[string]*Server{
					"server": {
						name: "server",
						replicas: map[int]*ServerReplica{
							0: {loadedModels: map[ModelVersionID]bool{{Name: "foo", Version: 2}: true, {Name: "foo", Version: 1}: true}, reservedMemory: memBytes, uniqueLoadedModels: map[string]bool{"foo": true}},
							1: {loadedModels: map[ModelVersionID]bool{}, reservedMemory: memBytes, uniqueLoadedModels: map[string]bool{}},
						},
					},
				},
			},
			modelKey:               "foo",
			version:                1,
			serverKey:              "server",
			replicaIdx:             0,
			expectedState:          Unloading,
			desiredState:           Unloaded,
			numModelVersionsLoaded: 1,
			modelVersionLoaded:     false,
			availableMemory:        20,
			err:                    false,
		},
		{
			name: "Model updated but not latest on replica which is Available",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"foo": {
					versions: []*ModelVersion{
						{
							modelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
							version:   1,
							replicas: map[int]ReplicaStatus{
								0: {State: Unloading},
							},
						},
						{
							modelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
							version:   2,
							replicas: map[int]ReplicaStatus{
								0: {State: Available},
							},
						},
					},
				}},
				servers: map[string]*Server{
					"server": {
						name: "server",
						replicas: map[int]*ServerReplica{
							0: {loadedModels: map[ModelVersionID]bool{{Name: "foo", Version: 2}: true, {Name: "foo", Version: 1}: true}, reservedMemory: memBytes, uniqueLoadedModels: map[string]bool{"foo": true}},
							1: {loadedModels: map[ModelVersionID]bool{}, reservedMemory: memBytes, uniqueLoadedModels: map[string]bool{}},
						},
					},
				},
			},
			modelKey:               "foo",
			version:                1,
			serverKey:              "server",
			replicaIdx:             0,
			expectedState:          Unloading,
			desiredState:           Unloaded,
			numModelVersionsLoaded: 1,
			modelVersionLoaded:     false,
			availableMemory:        20,
			err:                    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())
			if test.deleted {
				test.store.models[test.modelKey].SetDeleted()
			}
			ms := NewMemoryStore(logger, test.store, eventHub)
			err = ms.UpdateModelState(test.modelKey, test.version, test.serverKey, test.replicaIdx, &test.availableMemory, test.expectedState, test.desiredState, "")
			if !test.err {
				g.Expect(err).To(BeNil())
				if !test.deleted {
					g.Expect(test.store.models[test.modelKey].GetVersion(test.version).GetModelReplicaState(test.replicaIdx)).To(Equal(test.desiredState))
					g.Expect(test.store.servers[test.serverKey].replicas[test.replicaIdx].loadedModels[ModelVersionID{Name: test.modelKey, Version: test.version}]).To(Equal(test.modelVersionLoaded))
					g.Expect(test.store.servers[test.serverKey].replicas[test.replicaIdx].GetNumLoadedModels()).To(Equal(test.numModelVersionsLoaded))
				} else {
					// g.Expect(test.store.models[test.modelKey]).To(BeNil())
					g.Expect(test.store.models[test.modelKey].Latest().state.State).To(Equal(ModelTerminated))
				}

			} else {
				g.Expect(err).ToNot(BeNil())
			}
			if test.desiredState == Loaded || test.desiredState == LoadFailed {
				g.Expect(test.store.servers[test.serverKey].replicas[test.replicaIdx].GetReservedMemory()).To(Equal(uint64(0)))
			} else {
				g.Expect(test.store.servers[test.serverKey].replicas[test.replicaIdx].GetReservedMemory()).To(Equal(memBytes))
			}

			uniqueLoadedModels := toUniqueModels(test.store.servers[test.serverKey].replicas[test.replicaIdx].loadedModels)
			g.Expect(uniqueLoadedModels).To(Equal(test.store.servers[test.serverKey].replicas[test.replicaIdx].uniqueLoadedModels))
		})
	}
}

func TestUpdateModelStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                      string
		deleted                   bool
		modelVersion              *ModelVersion
		prevAvailableModelVersion *ModelVersion
		expectedState             ModelState
		expectedReason            string
		expectedAvailableReplicas uint32
		expectedTimestamp         time.Time
	}
	d1 := time.Date(2021, 1, 1, 12, 0, 0, 0, time.UTC)
	r1 := "reason1"
	d2 := time.Date(2021, 1, 2, 12, 0, 0, 0, time.UTC)
	r2 := "reason2"
	tests := []test{
		{
			name:    "Available",
			deleted: false,
			modelVersion: NewModelVersion(
				&pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
				1,
				"server",
				map[int]ReplicaStatus{
					0: {State: Available, Reason: "", Timestamp: d1},
				},
				false,
				ModelProgressing),
			prevAvailableModelVersion: nil,
			expectedState:             ModelAvailable,
			expectedAvailableReplicas: 1,
			expectedReason:            "",
			expectedTimestamp:         d1,
		},
		{
			name:    "Progressing",
			deleted: false,
			modelVersion: NewModelVersion(
				&pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 2}},
				1,
				"server",
				map[int]ReplicaStatus{
					0: {State: Available, Reason: "", Timestamp: d1},
					1: {State: Loading, Reason: "", Timestamp: d1},
				},
				false,
				ModelProgressing),
			prevAvailableModelVersion: nil,
			expectedState:             ModelProgressing,
			expectedAvailableReplicas: 1,
			expectedReason:            "",
			expectedTimestamp:         d1,
		},
		{
			name:    "Failed",
			deleted: false,
			modelVersion: NewModelVersion(
				&pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
				1,
				"server",
				map[int]ReplicaStatus{
					0: {State: LoadFailed, Reason: r1, Timestamp: d1},
				},
				false,
				ModelProgressing),
			prevAvailableModelVersion: nil,
			expectedState:             ModelFailed,
			expectedAvailableReplicas: 0,
			expectedReason:            r1,
			expectedTimestamp:         d1,
		},
		{
			name:    "AvailableAndFailed",
			deleted: false,
			modelVersion: NewModelVersion(
				&pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 2}},
				1,
				"server",
				map[int]ReplicaStatus{
					0: {State: Loaded, Reason: "", Timestamp: d1},
					1: {State: LoadFailed, Reason: r1, Timestamp: d2},
				},
				false,
				ModelProgressing),
			prevAvailableModelVersion: nil,
			expectedState:             ModelFailed,
			expectedAvailableReplicas: 0,
			expectedReason:            r1,
			expectedTimestamp:         d2,
		},
		{
			name:    "TwoFailed",
			deleted: false,
			modelVersion: NewModelVersion(
				&pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 2}},
				1,
				"server",
				map[int]ReplicaStatus{
					0: {State: LoadFailed, Reason: r1, Timestamp: d1},
					1: {State: LoadFailed, Reason: r2, Timestamp: d2},
				},
				false,
				ModelProgressing),
			prevAvailableModelVersion: nil,
			expectedState:             ModelFailed,
			expectedAvailableReplicas: 0,
			expectedReason:            r2,
			expectedTimestamp:         d2,
		},
		{
			name:    "AvailableV2",
			deleted: false,
			modelVersion: NewModelVersion(
				&pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
				1,
				"server",
				map[int]ReplicaStatus{
					0: {State: Loading, Reason: "", Timestamp: d1},
					1: {State: Available, Reason: "", Timestamp: d2},
				},
				false,
				ModelProgressing),
			prevAvailableModelVersion: NewModelVersion(
				&pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
				1,
				"server",
				map[int]ReplicaStatus{
					0: {State: Available, Reason: "", Timestamp: d1},
				},
				false,
				ModelAvailable),
			expectedState:             ModelAvailable,
			expectedAvailableReplicas: 1,
			expectedReason:            "",
			expectedTimestamp:         d2,
		},
		{
			name:    "Terminating",
			deleted: true,
			modelVersion: NewModelVersion(
				&pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 2}},
				1,
				"server",
				map[int]ReplicaStatus{
					0: {State: Unloading, Reason: "", Timestamp: d1},
					1: {State: Unloading, Reason: "", Timestamp: d2},
				},
				true,
				ModelProgressing),
			expectedState:             ModelTerminating,
			expectedAvailableReplicas: 0,
			expectedReason:            "",
			expectedTimestamp:         d2,
		},
		{
			name:    "TerminatingLoadingReplicas",
			deleted: true,
			modelVersion: NewModelVersion(
				&pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 2}},
				2,
				"server",
				map[int]ReplicaStatus{
					0: {State: Loading, Reason: "", Timestamp: d1},
					1: {State: Loading, Reason: "", Timestamp: d2},
				},
				true,
				ModelProgressing),
			expectedState:             ModelTerminating,
			expectedAvailableReplicas: 0,
			expectedReason:            "",
			expectedTimestamp:         d2,
		},
		{
			name:    "Terminated",
			deleted: true,
			modelVersion: NewModelVersion(
				&pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 2}},
				1,
				"server",
				map[int]ReplicaStatus{
					0: {State: Unloaded, Reason: "", Timestamp: d1},
					1: {State: Unloaded, Reason: "", Timestamp: d2},
				},
				true,
				ModelProgressing),
			expectedState:             ModelTerminated,
			expectedAvailableReplicas: 0,
			expectedReason:            "",
			expectedTimestamp:         d2,
		},
		{
			name:    "TerminateFailed",
			deleted: true,
			modelVersion: NewModelVersion(
				&pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 2}},
				1,
				"server",
				map[int]ReplicaStatus{
					0: {State: UnloadFailed, Reason: r1, Timestamp: d1},
					1: {State: Unloaded, Reason: "", Timestamp: d2},
				},
				true,
				ModelProgressing),
			expectedState:             ModelTerminateFailed,
			expectedAvailableReplicas: 0,
			expectedReason:            r1,
			expectedTimestamp:         d1,
		},
		{
			name:    "AvailableV2PrevTerminated",
			deleted: false,
			modelVersion: NewModelVersion(
				&pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 2}},
				2,
				"server",
				map[int]ReplicaStatus{
					0: {State: Available, Reason: "", Timestamp: d1},
					1: {State: Available, Reason: "", Timestamp: d2},
				},
				false,
				ModelProgressing),
			prevAvailableModelVersion: NewModelVersion(
				&pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 2}},
				1,
				"server",
				map[int]ReplicaStatus{
					0: {State: Available, Reason: "", Timestamp: d1},
					1: {State: Available, Reason: "", Timestamp: d2},
				},
				false,
				ModelTerminating),
			expectedState:             ModelAvailable,
			expectedAvailableReplicas: 2,
			expectedReason:            "",
			expectedTimestamp:         d2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())
			ms := NewMemoryStore(logger, &LocalSchedulerStore{}, eventHub)
			ms.updateModelStatus(true, test.deleted, test.modelVersion, test.prevAvailableModelVersion)
			g.Expect(test.modelVersion.state.State).To(Equal(test.expectedState))
			g.Expect(test.modelVersion.state.Reason).To(Equal(test.expectedReason))
			g.Expect(test.modelVersion.state.AvailableReplicas).To(Equal(test.expectedAvailableReplicas))
			g.Expect(test.modelVersion.state.Timestamp).To(Equal(test.expectedTimestamp))
		})
	}
}

func TestAddModelVersionIfNotExists(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                 string
		store                *LocalSchedulerStore
		modelVersion         *agent.ModelVersion
		expected             []uint32
		latest               uint32
		expectedVersion      uint32
		expectedAgentVersion uint32
	}

	tests := []test{
		{
			name: "Add new version when none exist",
			store: &LocalSchedulerStore{
				models: map[string]*Model{},
			},
			modelVersion: &agent.ModelVersion{
				Version: 1,
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo"},
				},
			},
			expected:             []uint32{1},
			latest:               1,
			expectedVersion:      1,
			expectedAgentVersion: 1,
		},
		{
			name: "AddNewVersion",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"foo": {
					versions: []*ModelVersion{},
				}},
			},
			modelVersion: &agent.ModelVersion{
				Version: 1,
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo"},
				},
			},
			expected:             []uint32{1},
			latest:               1,
			expectedVersion:      1,
			expectedAgentVersion: 1,
		},
		{
			name: "AddSameVersion - same generation",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"foo": {
					versions: []*ModelVersion{
						{
							version:   1,
							modelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 1}}},
							replicas:  map[int]ReplicaStatus{},
						},
					},
				}},
			},
			modelVersion: &agent.ModelVersion{
				Version: 1,
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 1}},
				},
			},
			expected:             []uint32{1},
			latest:               1,
			expectedVersion:      1,
			expectedAgentVersion: 1,
		},
		{
			name: "AddSecondVersion - no kubernetes generation",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"foo": {
					versions: []*ModelVersion{
						{
							version:   1,
							modelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo"}},
							replicas:  map[int]ReplicaStatus{},
						},
					},
				}},
			},
			modelVersion: &agent.ModelVersion{
				Version: 2,
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo"},
				},
			},
			expected:             []uint32{1, 2},
			latest:               2,
			expectedVersion:      2,
			expectedAgentVersion: 2,
		},
		{
			name: "AddSecondVersion",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"foo": {
					versions: []*ModelVersion{
						{
							version:   1,
							modelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 1}}},
							replicas:  map[int]ReplicaStatus{},
						},
					},
				}},
			},
			modelVersion: &agent.ModelVersion{
				Version: 2,
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 2}},
				},
			},
			expected:             []uint32{1, 2},
			latest:               2,
			expectedVersion:      2,
			expectedAgentVersion: 2,
		},
		{
			name: "Existing",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"foo": {
					versions: []*ModelVersion{
						{
							version:   1,
							modelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo"}},
							replicas:  map[int]ReplicaStatus{},
						},
					},
				}},
			},
			modelVersion: &agent.ModelVersion{
				Version: 1,
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo"},
				},
			},
			expected:             []uint32{1},
			latest:               1,
			expectedVersion:      1,
			expectedAgentVersion: 1,
		},
		{
			name: "AddThirdVersion",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"foo": {
					versions: []*ModelVersion{
						{
							version:   1,
							modelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 1}}},
							replicas:  map[int]ReplicaStatus{},
						},
						{
							version:   2,
							modelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 2}}},
							replicas:  map[int]ReplicaStatus{},
						},
					},
				}},
			},
			modelVersion: &agent.ModelVersion{
				Version: 3,
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 3}},
				},
			},
			expected:             []uint32{1, 2, 3},
			latest:               3,
			expectedVersion:      3,
			expectedAgentVersion: 3,
		},
		{
			name: "AddThirdVersionInMiddle",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"foo": {
					versions: []*ModelVersion{
						{
							version:   1,
							modelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 1}}},
							replicas:  map[int]ReplicaStatus{},
						},
						{
							version:   3,
							modelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 3}}},
							replicas:  map[int]ReplicaStatus{},
						},
					},
				}},
			},
			modelVersion: &agent.ModelVersion{
				Version: 2,
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 2}},
				},
			},
			expected:             []uint32{1, 2, 3},
			latest:               3,
			expectedVersion:      2,
			expectedAgentVersion: 2,
		},
		{
			name: "Add existing version - old generation - same spec",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"foo": {
					versions: []*ModelVersion{
						{
							version:   1,
							modelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 2}}},
							replicas:  map[int]ReplicaStatus{},
						},
					},
				}},
			},
			modelVersion: &agent.ModelVersion{
				Version: 1,
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 1}},
				},
			},
			expected:             []uint32{1},
			latest:               1,
			expectedVersion:      1,
			expectedAgentVersion: 1,
		},
		{
			name: "Add existing version - old generation - new spec",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"foo": {
					versions: []*ModelVersion{
						{
							version:   1,
							modelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 2}}, ModelSpec: &pb.ModelSpec{Uri: "dummy"}},
							replicas:  map[int]ReplicaStatus{},
						},
					},
				}},
			},
			modelVersion: &agent.ModelVersion{
				Version: 1,
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 1}},
				},
			},
			expected:             []uint32{1, 2, 3},
			latest:               3,
			expectedVersion:      2,
			expectedAgentVersion: 1, // this is a mismatch because the incoming model is the same as the existing model and it has different spec
		},
		{
			name: "Add existing version - new generation - new spec",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"foo": {
					versions: []*ModelVersion{
						{
							version:   1,
							modelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 2}}, ModelSpec: &pb.ModelSpec{Uri: "dummy"}},
							replicas:  map[int]ReplicaStatus{},
						},
					},
				}},
			},
			modelVersion: &agent.ModelVersion{
				Version: 1,
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 3}},
				},
			},
			expected:             []uint32{1, 2, 3}, // we still create version 3 because the incoming model has a different spec
			latest:               3,
			expectedVersion:      2,
			expectedAgentVersion: 1, // this is a mismatch because the incoming model is the same as the existing model and it has different spec
		},
		{
			name: "Add existing version - new generation - same spec",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"foo": {
					versions: []*ModelVersion{
						{
							version:   1,
							modelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 2}}},
							replicas:  map[int]ReplicaStatus{},
						},
					},
				}},
			},
			modelVersion: &agent.ModelVersion{
				Version: 1,
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 3}},
				},
			},
			expected:             []uint32{1},
			latest:               1,
			expectedVersion:      1,
			expectedAgentVersion: 1,
		},
		{
			name: "Add existing version - old generation - new spec - 2 versions",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"foo": {
					versions: []*ModelVersion{
						{
							version:   1,
							modelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 2}}, ModelSpec: &pb.ModelSpec{Uri: "dummy"}},
							replicas:  map[int]ReplicaStatus{},
						},
						{
							version:   2,
							modelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 3}}, ModelSpec: &pb.ModelSpec{Uri: "dummy"}},
							replicas:  map[int]ReplicaStatus{},
						},
					},
				}},
			},
			modelVersion: &agent.ModelVersion{
				Version: 1,
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 1}},
				},
			},
			expected:             []uint32{1, 2, 3, 4}, // create version 3 (incoming model) and 4 (promote max generation)
			latest:               4,
			expectedVersion:      3,
			expectedAgentVersion: 1, // this is a mismatch because the incoming model is the same as the existing model and it has different spec
		},
		{
			name: "Add new version - new generation",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"foo": {
					versions: []*ModelVersion{
						{
							version:   2,
							modelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 2}}, ModelSpec: &pb.ModelSpec{Uri: "dummy"}},
							replicas:  map[int]ReplicaStatus{},
						},
					},
				}},
			},
			modelVersion: &agent.ModelVersion{
				Version: 3,
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 3}},
				},
			},
			expected:             []uint32{2, 3}, // create version 3 (incoming model)
			latest:               3,
			expectedVersion:      3,
			expectedAgentVersion: 3,
		},
		{
			name: "Add new version - old generation",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"foo": {
					versions: []*ModelVersion{
						{
							version:   2,
							modelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 2}}, ModelSpec: &pb.ModelSpec{Uri: "dummy"}},
							replicas:  map[int]ReplicaStatus{},
						},
					},
				}},
			},
			modelVersion: &agent.ModelVersion{
				Version: 30,
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 1}},
				},
			},
			expected:             []uint32{2, 30, 31}, // create version 3 (incoming model)
			latest:               31,
			expectedVersion:      30,
			expectedAgentVersion: 30,
		},
		{
			name: "Add new version - same spec",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"foo": {
					versions: []*ModelVersion{
						{
							version:   2,
							modelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 2}}},
							replicas:  map[int]ReplicaStatus{},
						},
					},
				}},
			},
			modelVersion: &agent.ModelVersion{
				Version: 30,
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 1}},
				},
			},
			expected:             []uint32{2, 30, 31}, // create version 31 (new generation)
			latest:               31,
			expectedVersion:      30,
			expectedAgentVersion: 30,
		},
		{
			name: "Add new version - same spec - new generation",
			store: &LocalSchedulerStore{
				models: map[string]*Model{"foo": {
					versions: []*ModelVersion{
						{
							version:   2,
							modelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 2}}},
							replicas:  map[int]ReplicaStatus{},
						},
					},
				}},
			},
			modelVersion: &agent.ModelVersion{
				Version: 30,
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo", KubernetesMeta: &pb.KubernetesMeta{Generation: 30}},
				},
			},
			expected:             []uint32{2, 30},
			latest:               30,
			expectedVersion:      30,
			expectedAgentVersion: 30,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())
			ms := NewMemoryStore(logger, test.store, eventHub)
			_, mv := ms.addModelVersionIfNotExists(test.modelVersion)
			modelName := test.modelVersion.GetModel().GetMeta().GetName()
			g.Expect(test.store.models[modelName].GetVersions()).To(Equal(test.expected))
			g.Expect(test.store.models[modelName].Latest().version).To(Equal(test.latest))
			g.Expect(mv.GetVersion()).To(Equal(test.expectedVersion))
			g.Expect(mv.GetAgentVersion()).To(Equal(test.expectedAgentVersion))
		})
	}
}

func TestAddServerReplica(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                 string
		store                *LocalSchedulerStore
		req                  *agent.AgentSubscribeRequest
		expectedSnapshot     []*ServerSnapshot
		expectedModelEvents  int64
		expectedServerEvents int64
	}

	tests := []test{
		{
			name: "AddServerReplica - existing server",
			store: &LocalSchedulerStore{
				servers: map[string]*Server{
					"server1": {
						name: "server1",
						replicas: map[int]*ServerReplica{
							0: {},
							1: {},
						},
						expectedReplicas: 3,
						shared:           true,
					},
				},
			},
			req: &agent.AgentSubscribeRequest{
				ServerName: "server1",
				ReplicaIdx: 2,
				Shared:     true,
			},
			expectedSnapshot: []*ServerSnapshot{
				{
					Name: "server1",
					Replicas: map[int]*ServerReplica{
						0: {},
						1: {},
						2: {},
					},
					ExpectedReplicas: 3,
					Shared:           true,
				},
			},
			expectedModelEvents:  0,
			expectedServerEvents: 1,
		},
		{
			name: "AddServerReplica - new server",
			store: &LocalSchedulerStore{
				servers: map[string]*Server{},
			},
			req: &agent.AgentSubscribeRequest{
				ServerName: "server1",
				ReplicaIdx: 0,
				Shared:     true,
			},
			expectedSnapshot: []*ServerSnapshot{
				{
					Name: "server1",
					Replicas: map[int]*ServerReplica{
						0: {},
					},
					ExpectedReplicas: -1, // expected replicas is not set
					Shared:           true,
				},
			},
			expectedModelEvents:  0,
			expectedServerEvents: 1,
		},
		{
			name: "AddServerReplica - with loaded models",
			store: &LocalSchedulerStore{
				servers: map[string]*Server{},
				models:  map[string]*Model{},
			},
			req: &agent.AgentSubscribeRequest{
				ServerName: "server1",
				ReplicaIdx: 0,
				Shared:     true,
				LoadedModels: []*agent.ModelVersion{
					{
						Model: &pb.Model{
							Meta:      &pb.MetaData{Name: "model1"},
							ModelSpec: &pb.ModelSpec{}},
						Version: 1,
					},
					{
						Model: &pb.Model{
							Meta:      &pb.MetaData{Name: "model2"},
							ModelSpec: &pb.ModelSpec{}},
						Version: 1,
					},
				},
			},
			expectedSnapshot: []*ServerSnapshot{
				{
					Name: "server1",
					Replicas: map[int]*ServerReplica{
						0: {},
					},
					ExpectedReplicas: -1, // expected replicas is not set
					Shared:           true,
				},
			},
			expectedModelEvents:  2,
			expectedServerEvents: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())
			ms := NewMemoryStore(logger, test.store, eventHub)

			// register a callback to check if the event is triggered
			serverEvents := int64(0)
			eventHub.RegisterServerEventHandler(
				"handler-server",
				10,
				logger,
				func(event coordinator.ServerEventMsg) { atomic.AddInt64(&serverEvents, 1) },
			)

			modelEvents := int64(0)
			eventHub.RegisterModelEventHandler(
				"handler-model",
				10,
				logger,
				func(event coordinator.ModelEventMsg) { atomic.AddInt64(&modelEvents, 1) },
			)

			err = ms.AddServerReplica(test.req)
			g.Expect(err).To(BeNil())
			actualSnapshot, err := ms.GetServers(true, false)
			g.Expect(err).To(BeNil())
			for idx, server := range actualSnapshot {
				g.Expect(server.Name).To(Equal(test.expectedSnapshot[idx].Name))
				g.Expect(server.Shared).To(Equal(test.expectedSnapshot[idx].Shared))
				g.Expect(server.ExpectedReplicas).To(Equal(test.expectedSnapshot[idx].ExpectedReplicas))
				g.Expect(len(server.Replicas)).To(Equal(len(test.expectedSnapshot[idx].Replicas)))
			}

			time.Sleep(10 * time.Millisecond)
			g.Expect(atomic.LoadInt64(&serverEvents)).To(Equal(test.expectedServerEvents))
			g.Expect(atomic.LoadInt64(&modelEvents)).To(Equal(test.expectedModelEvents))
		})
	}
}

func TestRemoveServerReplica(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name           string
		store          *LocalSchedulerStore
		serverName     string
		replicaIdx     int
		serverExists   bool
		modelsReturned int
	}

	tests := []test{
		{
			name: "ReplicaRemovedButNotDeleted",
			store: &LocalSchedulerStore{
				servers: map[string]*Server{
					"server1": {
						name: "server1",
						replicas: map[int]*ServerReplica{
							0: {loadedModels: map[ModelVersionID]bool{{Name: "model1", Version: 1}: true}},
							1: {},
						},
						expectedReplicas: 2,
						shared:           true,
					},
				},
			},
			serverName:     "server1",
			replicaIdx:     0,
			serverExists:   true,
			modelsReturned: 0, // no models really defined in store
		},
		{
			name: "ReplicaRemovedAndDeleted",
			store: &LocalSchedulerStore{
				models: map[string]*Model{
					"model1": {
						versions: []*ModelVersion{
							{
								version: 1,
							},
						},
					},
				},
				servers: map[string]*Server{
					"server1": {
						name: "server1",
						replicas: map[int]*ServerReplica{
							0: {loadedModels: map[ModelVersionID]bool{{Name: "model1", Version: 1}: true}},
							1: {},
						},
						expectedReplicas: -1,
						shared:           true,
					},
				},
			},
			serverName:     "server1",
			replicaIdx:     0,
			serverExists:   true,
			modelsReturned: 1,
		},
		{
			name: "ReplicaRemovedAndServerDeleted",
			store: &LocalSchedulerStore{
				models: map[string]*Model{
					"model1": {
						versions: []*ModelVersion{
							{
								version: 1,
							},
						},
					},
				},
				servers: map[string]*Server{
					"server1": {
						name: "server1",
						replicas: map[int]*ServerReplica{
							0: {loadedModels: map[ModelVersionID]bool{{Name: "model1", Version: 1}: true}},
						},
						expectedReplicas: 0,
						shared:           true,
					},
				},
			},
			serverName:     "server1",
			replicaIdx:     0,
			serverExists:   false,
			modelsReturned: 1,
		},
		{
			name: "ReplicaRemovedAndServerDeleted but no model version in store",
			store: &LocalSchedulerStore{
				servers: map[string]*Server{
					"server1": {
						name: "server1",
						replicas: map[int]*ServerReplica{
							0: {loadedModels: map[ModelVersionID]bool{{Name: "model1", Version: 1}: true}},
						},
						expectedReplicas: 0,
						shared:           true,
					},
				},
			},
			serverName:     "server1",
			replicaIdx:     0,
			serverExists:   false,
			modelsReturned: 0,
		},
		{
			name: "ReplicaRemovedAndDeleted - loading models",
			store: &LocalSchedulerStore{
				models: map[string]*Model{
					"model1": {
						versions: []*ModelVersion{
							{
								version: 1,
							},
						},
					},
					"model2": {
						versions: []*ModelVersion{
							{
								version: 1,
							},
						},
					},
				},
				servers: map[string]*Server{
					"server1": {
						name: "server1",
						replicas: map[int]*ServerReplica{
							0: {
								loadedModels:  map[ModelVersionID]bool{{Name: "model1", Version: 1}: true},
								loadingModels: map[ModelVersionID]bool{{Name: "model2", Version: 1}: true},
							},
							1: {},
						},
						expectedReplicas: -1,
						shared:           true,
					},
				},
			},
			serverName:     "server1",
			replicaIdx:     0,
			serverExists:   true,
			modelsReturned: 2,
		},
		{
			name: "ReplicaRemovedAndDeleted - non latest models",
			store: &LocalSchedulerStore{
				models: map[string]*Model{
					"model1": {
						versions: []*ModelVersion{
							{
								version:  1,
								replicas: map[int]ReplicaStatus{0: {State: Loaded}},
							},
							{
								version:  2,
								replicas: map[int]ReplicaStatus{0: {State: LoadFailed}},
							},
						},
					},
				},
				servers: map[string]*Server{
					"server1": {
						name: "server1",
						replicas: map[int]*ServerReplica{
							0: {
								loadedModels: map[ModelVersionID]bool{{Name: "model1", Version: 1}: true},
							},
						},
						expectedReplicas: -1,
						shared:           true,
					},
				},
			},
			serverName:     "server1",
			replicaIdx:     0,
			serverExists:   false,
			modelsReturned: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())
			ms := NewMemoryStore(logger, test.store, eventHub)
			models, err := ms.RemoveServerReplica(test.serverName, test.replicaIdx)
			g.Expect(err).To(BeNil())
			g.Expect(test.modelsReturned).To(Equal(len(models)))
			server, err := ms.GetServer(test.serverName, false, true)
			if test.serverExists {
				g.Expect(err).To(BeNil())
				g.Expect(server).ToNot(BeNil())
			} else {
				g.Expect(err).ToNot(BeNil())
				g.Expect(server).To(BeNil())
			}
		})
	}
}

func TestDrainServerReplica(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name           string
		store          *LocalSchedulerStore
		serverName     string
		replicaIdx     int
		modelsReturned []string
	}

	// if we have models returned check status is Draining
	tests := []test{
		{
			name: "ReplicaSetDrainingNoModels",
			store: &LocalSchedulerStore{
				servers: map[string]*Server{
					"server1": {
						name: "server1",
						replicas: map[int]*ServerReplica{
							0: {},
							1: {},
						},
						expectedReplicas: 2,
						shared:           true,
					},
				},
			},
			serverName:     "server1",
			replicaIdx:     0,
			modelsReturned: nil,
		},
		{
			name: "ReplicaSetDrainingWithModels",
			store: &LocalSchedulerStore{
				models: map[string]*Model{
					"model1": {
						versions: []*ModelVersion{
							{
								version:  1,
								replicas: map[int]ReplicaStatus{0: {State: Loaded}},
							},
						},
					},
				},
				servers: map[string]*Server{
					"server1": {
						name: "server1",
						replicas: map[int]*ServerReplica{
							0: {loadedModels: map[ModelVersionID]bool{{Name: "model1", Version: 1}: true}},
							1: {},
						},
						expectedReplicas: -1,
						shared:           true,
					},
				},
			},
			serverName:     "server1",
			replicaIdx:     0,
			modelsReturned: []string{"model1"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())
			ms := NewMemoryStore(logger, test.store, eventHub)
			models, err := ms.DrainServerReplica(test.serverName, test.replicaIdx)
			g.Expect(err).To(BeNil())
			g.Expect(test.modelsReturned).To(Equal(models))
			server, err := ms.GetServer(test.serverName, false, true)
			g.Expect(err).To(BeNil())
			g.Expect(server).ToNot(BeNil())
			g.Expect(server.Replicas[test.replicaIdx].GetIsDraining()).To(BeTrue())

			if test.modelsReturned != nil {
				for _, model := range test.modelsReturned {
					modelVersion, _ := ms.GetModel(model)
					state := modelVersion.GetLatest().GetModelReplicaState(test.replicaIdx)
					g.Expect(state).To(Equal(Draining))
				}
			}
		})
	}
}
