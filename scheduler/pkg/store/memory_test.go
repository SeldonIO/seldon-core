/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package store

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
)

func TestUpdateModel(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()

	type test struct {
		name            string
		models          []*db.Model
		loadModelReq    *pb.LoadModelRequest
		expectedVersion uint32
		err             error
	}

	tests := []test{
		{
			name:   "simple",
			models: []*db.Model{},
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
			name:   "simple with generation",
			models: []*db.Model{},
			loadModelReq: &pb.LoadModelRequest{
				Model: &pb.Model{
					Meta: &pb.MetaData{
						Name: "model",
						KubernetesMeta: &pb.KubernetesMeta{
							Generation: 100,
						},
					},
				},
			},
			expectedVersion: 100,
		},
		{
			name: "VersionAlreadyExists",
			models: []*db.Model{
				{
					Name: "model",
					Versions: []*db.ModelVersion{
						{
							Version: 1,
							ModelDefn: &pb.Model{
								Meta: &pb.MetaData{
									Name: "model",
								},
							},
							Replicas: make(map[int32]*db.ReplicaStatus),
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
			models: []*db.Model{
				{
					Name: "model",
					Versions: []*db.ModelVersion{
						{
							Version: 1,
							ModelDefn: &pb.Model{
								Meta: &pb.MetaData{
									Name: "model",
								},
							},
							Replicas: make(map[int32]*db.ReplicaStatus),
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
			models: []*db.Model{
				{
					Name: "model",
					Versions: []*db.ModelVersion{
						{
							Version: 1,
							ModelDefn: &pb.Model{
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
							Replicas: make(map[int32]*db.ReplicaStatus),
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
			models: []*db.Model{
				{
					Name: "model",
					Versions: []*db.ModelVersion{
						{
							Version: 1,
							ModelDefn: &pb.Model{
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
							Replicas: make(map[int32]*db.ReplicaStatus),
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
			name:   "ModelNameIsNotValid",
			models: []*db.Model{},
			loadModelReq: &pb.LoadModelRequest{
				Model: &pb.Model{
					Meta: &pb.MetaData{
						Name: "this.Name",
					},
				},
			},
			expectedVersion: 1,
			err:             errors.New("model this.Name does not have a valid name - it must be alphanumeric and not contains dots (.)"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())

			// Create storage instances
			modelStorage := NewInMemoryStorage[*db.Model]()
			serverStorage := NewInMemoryStorage[*db.Server]()

			// Populate storage with test data
			for _, model := range test.models {
				err := modelStorage.Insert(ctx, model)
				g.Expect(err).To(BeNil())
			}

			// Create MemoryStore with populated storage
			ms := NewModelServerStore(logger, modelStorage, serverStorage, eventHub)

			err = ms.UpdateModel(test.loadModelReq)
			if test.err != nil {
				g.Expect(err.Error()).To(BeIdenticalTo(test.err.Error()))
				return
			}

			g.Expect(err).To(BeNil())
			model, err := modelStorage.Get(ctx, test.loadModelReq.GetModel().GetMeta().GetName())
			g.Expect(err).To(BeNil())
			latest := model.Latest()

			g.Expect(proto.Equal(latest.ModelDefn, test.loadModelReq.Model)).To(BeTrue())

			g.Expect(latest.ModelDefn).To(Equal(test.loadModelReq.Model))
			g.Expect(latest.GetVersion()).To(Equal(test.expectedVersion))
		})
	}
}

func TestGetModel(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()

	type test struct {
		name     string
		models   []*db.Model
		key      string
		versions int
		err      error
	}

	tests := []test{
		{
			name:     "NoModel",
			models:   []*db.Model{},
			key:      "model",
			versions: 0,
			err:      ErrNotFound,
		},
		{
			name: "VersionAlreadyExists",
			models: []*db.Model{
				{
					Name: "model",
					Versions: []*db.ModelVersion{
						{
							ModelDefn: &pb.Model{
								Meta: &pb.MetaData{
									Name: "model",
								},
							},
							Replicas: make(map[int32]*db.ReplicaStatus),
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

			// Create storage instances
			modelStorage := NewInMemoryStorage[*db.Model]()
			serverStorage := NewInMemoryStorage[*db.Server]()

			// Populate storage with test data
			for _, model := range test.models {
				err := modelStorage.Insert(ctx, model)
				g.Expect(err).To(BeNil())
			}

			// Create MemoryStore with populated storage
			ms := NewModelServerStore(logger, modelStorage, serverStorage, eventHub)

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

func TestGetServer(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()

	type test struct {
		name     string
		models   []*db.Model
		servers  []*db.Server
		key      string
		isErr    bool
		expected *db.Server
	}

	tests := []test{
		{
			name:     "NoServer",
			models:   []*db.Model{},
			servers:  []*db.Server{},
			key:      "server",
			isErr:    true,
			expected: nil,
		},
		{
			name:   "ServerExists",
			models: []*db.Model{},
			servers: []*db.Server{
				{
					Name: "server",
					Replicas: map[int32]*db.ServerReplica{
						0: {},
					},
					ExpectedReplicas: 1,
					MinReplicas:      0,
					MaxReplicas:      0,
				},
			},
			key:   "server",
			isErr: false,
			expected: &db.Server{
				Name:             "server",
				ExpectedReplicas: 1,
				MinReplicas:      0,
				MaxReplicas:      0,
				Replicas: map[int32]*db.ServerReplica{
					0: {},
				},
			},
		},
		{
			name: "ServerExistsWithModel",
			models: []*db.Model{
				{
					Name: "model",
					Versions: []*db.ModelVersion{
						{
							Version: 1,
							ModelDefn: &pb.Model{
								Meta: &pb.MetaData{
									Name: "model",
								},
							},
							Replicas: make(map[int32]*db.ReplicaStatus),
						},
					},
				},
			},
			servers: []*db.Server{
				{
					Name: "server",
					Replicas: map[int32]*db.ServerReplica{
						0: {
							LoadedModels: []*db.ModelVersionID{{
								Version: 1,
								Name:    "model",
							},
							},
						},
					},
					ExpectedReplicas: 1,
					MinReplicas:      0,
					MaxReplicas:      0,
				},
			},
			key:   "server",
			isErr: false,
			expected: &db.Server{
				Name:             "server",
				ExpectedReplicas: 1,
				MinReplicas:      0,
				MaxReplicas:      0,
				Replicas: map[int32]*db.ServerReplica{
					0: {
						LoadedModels: []*db.ModelVersionID{
							{
								Name:    "model",
								Version: 1,
							},
						}},
				},
			},
		},
		{
			name: "ServerWithEmptyReplicas",
			models: []*db.Model{
				{
					Name: "model",
					Versions: []*db.ModelVersion{
						{
							Version: 1,
							ModelDefn: &pb.Model{
								Meta: &pb.MetaData{
									Name: "model",
								},
							},
							Replicas: make(map[int32]*db.ReplicaStatus),
						},
					},
				},
			},
			servers: []*db.Server{
				{
					Name: "server",
					Replicas: map[int32]*db.ServerReplica{
						0: {
							LoadedModels: []*db.ModelVersionID{{
								Version: 1,
								Name:    "model",
							},
							},
						},
						1: {},
					},
					ExpectedReplicas: 1,
					MinReplicas:      0,
					MaxReplicas:      0,
				},
			},
			key:   "server",
			isErr: false,
			expected: &db.Server{
				Name:             "server",
				ExpectedReplicas: 1,
				MinReplicas:      0,
				MaxReplicas:      0,
				Replicas: map[int32]*db.ServerReplica{
					0: {
						LoadedModels: []*db.ModelVersionID{{
							Version: 1,
							Name:    "model",
						},
						},
					},
					1: {
						LoadedModels: []*db.ModelVersionID{},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())

			// Create storage instances
			modelStorage := NewInMemoryStorage[*db.Model]()
			serverStorage := NewInMemoryStorage[*db.Server]()

			// Populate storage with test data
			for _, model := range test.models {
				err := modelStorage.Insert(ctx, model)
				g.Expect(err).To(BeNil())
			}
			for _, server := range test.servers {
				err := serverStorage.Insert(ctx, server)
				g.Expect(err).To(BeNil())
			}

			// Create MemoryStore with populated storage
			ms := NewModelServerStore(logger, modelStorage, serverStorage, eventHub)

			server, _, err := ms.GetServer(test.key, true)
			if !test.isErr {
				g.Expect(err).To(BeNil())
				g.Expect(server.Name).To(Equal(test.expected.Name))
				g.Expect(server.ExpectedReplicas).To(Equal(test.expected.ExpectedReplicas))
				for k, v := range server.Replicas {
					g.Expect(len(v.LoadedModels)).To(Equal(len(test.expected.Replicas[k].LoadedModels)))
					for i, m := range v.LoadedModels {
						g.Expect(proto.Equal(m, test.expected.Replicas[k].LoadedModels[i])).To(BeTrue())
					}
				}
				return
			}
			g.Expect(err).ToNot(BeNil())
		})
	}
}

func TestRemoveModel(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()

	type test struct {
		name    string
		models  []*db.Model
		servers []*db.Server
		key     string
		err     bool
	}

	tests := []test{
		{
			name:    "NoModel",
			models:  []*db.Model{},
			servers: []*db.Server{},
			key:     "model",
			err:     true,
		},
		{
			name: "VersionAlreadyExists",
			models: []*db.Model{
				{
					Name: "model",
					Versions: []*db.ModelVersion{
						{
							Version: 1,
							ModelDefn: &pb.Model{
								Meta: &pb.MetaData{
									Name: "model",
								},
							},
							Replicas: make(map[int32]*db.ReplicaStatus),
							State:    &db.ModelStatus{},
						},
					},
				},
			},
			servers: []*db.Server{},
			key:     "model",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())

			// Create storage instances
			modelStorage := NewInMemoryStorage[*db.Model]()
			serverStorage := NewInMemoryStorage[*db.Server]()

			// Populate storage with test data
			for _, model := range test.models {
				err := modelStorage.Insert(ctx, model)
				g.Expect(err).To(BeNil())
			}
			for _, server := range test.servers {
				err := serverStorage.Insert(ctx, server)
				g.Expect(err).To(BeNil())
			}

			// Create MemoryStore with populated storage
			ms := NewModelServerStore(logger, modelStorage, serverStorage, eventHub)
			err = ms.RemoveModel(&pb.UnloadModelRequest{Model: &pb.ModelReference{Name: test.key}})
			if !test.err {
				g.Expect(err).To(BeNil())
				return
			}
			g.Expect(err).ToNot(BeNil())
		})
	}
}

func TestUpdateLoadedModels(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()
	memBytes := uint64(1)

	type test struct {
		name               string
		models             []*db.Model
		servers            []*db.Server
		modelName          string
		version            uint32
		serverKey          string
		replicas           []*db.ServerReplica
		expectedStates     map[int]db.ModelReplicaState
		err                bool
		isModelDeleted     bool
		expectedModelState *db.ModelStatus
	}

	tests := []test{
		{
			name: "ModelVersionNotLatest",
			models: []*db.Model{{
				Name: "model",
				Versions: []*db.ModelVersion{
					{
						ModelDefn: &pb.Model{
							Meta: &pb.MetaData{
								Name: "model",
							},
							ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
						Server:   "server",
						Version:  1,
						Replicas: map[int32]*db.ReplicaStatus{},
						State:    &db.ModelStatus{},
					},
					{
						ModelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
						Server:    "server",
						Version:   2,
						Replicas:  map[int32]*db.ReplicaStatus{},
						State:     &db.ModelStatus{},
					},
				},
			}},
			servers: []*db.Server{
				{
					Name: "server",
				},
			},
			modelName: "model",
			version:   1,
			serverKey: "server",
			replicas:  nil,
			err:       true,
		},
		{
			name: "UpdatedVersionsOK",
			models: []*db.Model{{
				Name: "model",
				Versions: []*db.ModelVersion{
					{
						ModelDefn: &pb.Model{
							Meta: &pb.MetaData{
								Name: "model",
							},
							ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
						Server:   "server",
						Version:  1,
						Replicas: map[int32]*db.ReplicaStatus{},
						State:    &db.ModelStatus{},
					},
				},
			}},
			servers: []*db.Server{
				{
					Name: "server",
					Replicas: map[int32]*db.ServerReplica{
						0: {},
						1: {},
					},
				},
			},
			modelName: "model",
			version:   1,
			serverKey: "server",
			replicas: []*db.ServerReplica{
				{ReplicaIdx: 0}, {ReplicaIdx: 1},
			},
			expectedStates: map[int]db.ModelReplicaState{0: db.ModelReplicaState_LoadRequested, 1: db.ModelReplicaState_LoadRequested},
		},
		{
			name: "WithAlreadyLoadedModels",
			models: []*db.Model{{
				Name: "model",
				Versions: []*db.ModelVersion{
					{
						ModelDefn: &pb.Model{
							Meta: &pb.MetaData{
								Name: "model",
							},
							ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
						Server:  "server",
						Version: 1,
						Replicas: map[int32]*db.ReplicaStatus{
							0: {State: db.ModelReplicaState_Loaded},
						},
						State: &db.ModelStatus{},
					},
				},
			}},
			servers: []*db.Server{
				{
					Name: "server",
					Replicas: map[int32]*db.ServerReplica{
						0: {},
						1: {},
					},
				},
			},
			modelName: "model",
			version:   1,
			serverKey: "server",
			replicas: []*db.ServerReplica{
				{ReplicaIdx: 0}, {ReplicaIdx: 1},
			},
			expectedStates: map[int]db.ModelReplicaState{0: db.ModelReplicaState_Loaded, 1: db.ModelReplicaState_LoadRequested},
		},
		{
			name: "UnloadModelsNotSelected",
			models: []*db.Model{{
				Name: "model",
				Versions: []*db.ModelVersion{
					{
						ModelDefn: &pb.Model{
							Meta: &pb.MetaData{
								Name: "model",
							},
							ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
						Server:  "server",
						Version: 1,
						Replicas: map[int32]*db.ReplicaStatus{
							0: {State: db.ModelReplicaState_Loaded},
						},
						State: &db.ModelStatus{},
					},
				},
			}},
			servers: []*db.Server{
				{
					Name: "server",
					Replicas: map[int32]*db.ServerReplica{
						0: {},
						1: {},
					},
				},
			},
			modelName: "model",
			version:   1,
			serverKey: "server",
			replicas: []*db.ServerReplica{
				{ReplicaIdx: 1},
			},
			expectedStates: map[int]db.ModelReplicaState{0: db.ModelReplicaState_UnloadEnvoyRequested, 1: db.ModelReplicaState_LoadRequested},
		},
		{
			name: "DeletedModel",
			models: []*db.Model{{
				Name: "model",
				Versions: []*db.ModelVersion{
					{
						ModelDefn: &pb.Model{
							Meta: &pb.MetaData{
								Name: "model",
							},
							ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
						Server:  "server",
						Version: 1,
						Replicas: map[int32]*db.ReplicaStatus{
							0: {State: db.ModelReplicaState_Loaded},
							1: {State: db.ModelReplicaState_Loading},
						},
						State: &db.ModelStatus{},
					},
				},
			}},
			servers: []*db.Server{
				{
					Name: "server",
					Replicas: map[int32]*db.ServerReplica{
						0: {},
						1: {},
					},
				},
			},
			modelName:      "model",
			version:        1,
			serverKey:      "server",
			replicas:       []*db.ServerReplica{},
			isModelDeleted: true,
			expectedStates: map[int]db.ModelReplicaState{0: db.ModelReplicaState_UnloadEnvoyRequested, 1: db.ModelReplicaState_UnloadEnvoyRequested},
		},
		{
			name: "DeletedModelNoReplicas",
			models: []*db.Model{{
				Name: "model",
				Versions: []*db.ModelVersion{
					{
						ModelDefn: &pb.Model{
							Meta: &pb.MetaData{
								Name: "model",
							},
							ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
						Server:  "server",
						Version: 1,
						Replicas: map[int32]*db.ReplicaStatus{
							0: {State: db.ModelReplicaState_Unloaded},
						},
						State: &db.ModelStatus{},
					},
				},
			}},
			servers: []*db.Server{
				{
					Name: "server",
					Replicas: map[int32]*db.ServerReplica{
						0: {},
						1: {},
					},
				},
			},
			modelName:      "model",
			version:        1,
			serverKey:      "server",
			replicas:       []*db.ServerReplica{},
			isModelDeleted: true,
			expectedStates: map[int]db.ModelReplicaState{0: db.ModelReplicaState_Unloaded},
		},
		{
			name: "ServerChanged",
			models: []*db.Model{{
				Name: "model",
				Versions: []*db.ModelVersion{
					{
						ModelDefn: &pb.Model{
							Meta: &pb.MetaData{
								Name: "model",
							},
							ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
						Server:   "server1",
						Version:  1,
						Replicas: map[int32]*db.ReplicaStatus{},
						State:    &db.ModelStatus{},
					},
				},
			}},
			servers: []*db.Server{
				{
					Name: "server1",
					Replicas: map[int32]*db.ServerReplica{
						0: {},
						1: {},
					},
				},
				{
					Name: "server2",
					Replicas: map[int32]*db.ServerReplica{
						0: {},
						1: {},
					},
				},
			},
			modelName: "model",
			version:   1,
			serverKey: "server2",
			replicas: []*db.ServerReplica{
				{ReplicaIdx: 0}, {ReplicaIdx: 1},
			},
			expectedStates: map[int]db.ModelReplicaState{0: db.ModelReplicaState_LoadRequested, 1: db.ModelReplicaState_LoadRequested},
		},
		{
			name: "WithDrainingServerReplicaSameServer",
			models: []*db.Model{{
				Name: "model",
				Versions: []*db.ModelVersion{
					{
						ModelDefn: &pb.Model{
							Meta: &pb.MetaData{
								Name: "model",
							},
							ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
						Server:  "server",
						Version: 1,
						Replicas: map[int32]*db.ReplicaStatus{
							0: {State: db.ModelReplicaState_Draining},
						},
						State: &db.ModelStatus{},
					},
				},
			}},
			servers: []*db.Server{
				{
					Name: "server",
					Replicas: map[int32]*db.ServerReplica{
						0: {IsDraining: true},
						1: {},
					},
				},
			},
			modelName: "model",
			version:   1,
			serverKey: "server",
			replicas: []*db.ServerReplica{
				{ReplicaIdx: 1},
			},
			expectedStates: map[int]db.ModelReplicaState{0: db.ModelReplicaState_Draining, 1: db.ModelReplicaState_LoadRequested},
		},
		{
			name: "WithDrainingServerReplicaNewServer",
			models: []*db.Model{{
				Name: "model",
				Versions: []*db.ModelVersion{
					{
						ModelDefn: &pb.Model{
							Meta: &pb.MetaData{
								Name: "model",
							},
							ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes},
						},
						Server:  "server1",
						Version: 1,
						Replicas: map[int32]*db.ReplicaStatus{
							0: {State: db.ModelReplicaState_Draining},
						},
						State: &db.ModelStatus{},
					},
				},
			}},
			servers: []*db.Server{
				{
					Name: "server1",
					Replicas: map[int32]*db.ServerReplica{
						0: {IsDraining: true},
					},
				},
				{
					Name: "server2",
					Replicas: map[int32]*db.ServerReplica{
						0: {},
					},
				},
			},
			modelName: "model",
			version:   1,
			serverKey: "server2",
			replicas: []*db.ServerReplica{
				{ReplicaIdx: 0},
			},
			expectedStates: map[int]db.ModelReplicaState{0: db.ModelReplicaState_LoadRequested},
		},
		{
			name: "DeleteFailedScheduleModel",
			models: []*db.Model{{
				Name: "model",
				Versions: []*db.ModelVersion{
					{
						ModelDefn: &pb.Model{
							Meta: &pb.MetaData{
								Name: "model",
							},
							ModelSpec: &pb.ModelSpec{
								MemoryBytes: &memBytes,
							}},
						Server:   "",
						Version:  1,
						Replicas: map[int32]*db.ReplicaStatus{},
					},
				},
			}},
			servers:        []*db.Server{},
			modelName:      "model",
			version:        1,
			serverKey:      "",
			replicas:       []*db.ServerReplica{},
			isModelDeleted: true,
			expectedStates: map[int]db.ModelReplicaState{},
		},
		{
			name: "ProgressModelLoading",
			models: []*db.Model{
				{
					Name: "my-model",
					Versions: []*db.ModelVersion{
						{
							ModelDefn: &pb.Model{
								Meta: &pb.MetaData{
									Name: "my-model",
								},
								ModelSpec: &pb.ModelSpec{
									MemoryBytes: &memBytes,
								},
								DeploymentSpec: &pb.DeploymentSpec{
									Replicas: 1,
								},
							},
							Server:  "server",
							Version: 1,
							Replicas: map[int32]*db.ReplicaStatus{
								0: {State: db.ModelReplicaState_Available},
								1: {State: db.ModelReplicaState_Unloaded},
							},
							State: &db.ModelStatus{State: db.ModelState_ModelProgressing},
						},
					},
				}},
			servers: []*db.Server{
				{
					Name: "server",
					Replicas: map[int32]*db.ServerReplica{
						0: {},
						1: {},
					},
				},
			},
			modelName: "my-model",
			version:   1,
			serverKey: "server",
			replicas: []*db.ServerReplica{
				{ReplicaIdx: 0}, {ReplicaIdx: 0},
			},
			expectedStates:     map[int]db.ModelReplicaState{0: db.ModelReplicaState_Available, 1: db.ModelReplicaState_Unloaded},
			expectedModelState: &db.ModelStatus{State: db.ModelState_ModelAvailable},
		},
		{
			name: "PartiallyAvailableModels",
			models: []*db.Model{
				{
					Name: "my-model",
					Versions: []*db.ModelVersion{
						{
							ModelDefn: &pb.Model{
								Meta: &pb.MetaData{
									Name: "my-model",
								},
								ModelSpec: &pb.ModelSpec{
									MemoryBytes: &memBytes,
								},
								DeploymentSpec: &pb.DeploymentSpec{
									Replicas:    3,
									MinReplicas: 2,
								},
							},
							Server:  "server",
							Version: 1,
							Replicas: map[int32]*db.ReplicaStatus{
								0: {State: db.ModelReplicaState_Available},
								1: {State: db.ModelReplicaState_Available},
							},
							State: &db.ModelStatus{State: db.ModelState_ModelProgressing},
						},
					},
				}},
			servers: []*db.Server{
				{
					Name: "server",
					Replicas: map[int32]*db.ServerReplica{
						0: {},
						1: {},
					},
				},
			},
			modelName: "my-model",
			version:   1,
			serverKey: "server",
			replicas: []*db.ServerReplica{
				{ReplicaIdx: 0}, {ReplicaIdx: 1},
			},
			expectedStates:     map[int]db.ModelReplicaState{0: db.ModelReplicaState_Available, 1: db.ModelReplicaState_Available},
			expectedModelState: &db.ModelStatus{State: db.ModelState_ModelAvailable},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())

			// Create storage instances
			modelStorage := NewInMemoryStorage[*db.Model]()
			serverStorage := NewInMemoryStorage[*db.Server]()

			// Populate storage with test data
			for _, model := range test.models {
				if test.isModelDeleted {
					model.Deleted = true
				}
				err := modelStorage.Insert(ctx, model)
				g.Expect(err).To(BeNil())
			}
			for _, server := range test.servers {
				err := serverStorage.Insert(ctx, server)
				g.Expect(err).To(BeNil())
			}

			ms := NewModelServerStore(logger, modelStorage, serverStorage, eventHub)
			msg, err := ms.updateLoadedModelsImpl(test.modelName, test.version, test.serverKey, test.replicas)
			if test.err {
				g.Expect(err).ToNot(BeNil())
				g.Expect(msg).To(BeNil())
				return
			}

			g.Expect(err).To(BeNil())
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.ModelName).To(Equal(test.modelName))

			model, err := ms.GetModel(test.modelName)
			g.Expect(err).To(BeNil())

			mv := model.Latest()
			g.Expect(mv).ToNot(BeNil())

			for replicaIdx, state := range test.expectedStates {
				g.Expect(mv).ToNot(BeNil())
				g.Expect(mv.GetModelReplicaState(replicaIdx)).To(Equal(state))
				ss, _, err := ms.GetServer(test.serverKey, false)
				g.Expect(err).To(BeNil())

				if state == db.ModelReplicaState_LoadRequested {
					g.Expect(ss.Replicas[int32(replicaIdx)].GetReservedMemory()).To(Equal(memBytes))
					continue
				}
				g.Expect(ss.Replicas[int32(replicaIdx)].GetReservedMemory()).To(Equal(uint64(0)))
			}
			if test.expectedModelState != nil {
				g.Expect(mv.State.State).To(Equal(test.expectedModelState.State))
			}
		})
	}
}

func TestUpdateModelState(t *testing.T) {
	g := NewGomegaWithT(t)
	memBytes := uint64(1)

	type test struct {
		name                   string
		modelName              string
		models                 []*db.Model
		servers                []*db.Server
		version                uint32
		serverKey              string
		replicaIdx             int
		expectedState          db.ModelReplicaState
		desiredState           db.ModelReplicaState
		availableMemory        uint64
		modelRuntimeInfo       *pb.ModelRuntimeInfo
		numModelVersionsLoaded int
		modelVersionLoaded     bool
		deleted                bool
		err                    bool
	}

	tests := []test{
		{
			name: "LoadedModel",
			models: []*db.Model{{
				Name: "model",
				Versions: []*db.ModelVersion{
					{
						ModelDefn: &pb.Model{
							Meta: &pb.MetaData{
								Name: "model",
							},
							ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
						Version:  1,
						Replicas: map[int32]*db.ReplicaStatus{},
						State:    &db.ModelStatus{},
					},
				},
			}},
			servers: []*db.Server{
				{
					Name: "server",
					Replicas: map[int32]*db.ServerReplica{
						0: {LoadedModels: []*db.ModelVersionID{}, ReservedMemory: memBytes, UniqueLoadedModels: map[string]bool{}},
						1: {LoadedModels: []*db.ModelVersionID{}, ReservedMemory: memBytes, UniqueLoadedModels: map[string]bool{}},
					},
				},
			},
			modelName:              "model",
			version:                1,
			serverKey:              "server",
			replicaIdx:             0,
			expectedState:          db.ModelReplicaState_ModelReplicaStateUnknown,
			desiredState:           db.ModelReplicaState_Loaded,
			numModelVersionsLoaded: 1,
			modelVersionLoaded:     true,
			availableMemory:        20,
			modelRuntimeInfo:       &pb.ModelRuntimeInfo{ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{Mlserver: &pb.MLServerModelSettings{ParallelWorkers: uint32(1)}}},
		},
		{
			name: "UnloadedModel",
			models: []*db.Model{{
				Name: "model",
				Versions: []*db.ModelVersion{
					{
						ModelDefn: &pb.Model{
							Meta: &pb.MetaData{
								Name: "model",
							},
							ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
						Version:  1,
						Replicas: map[int32]*db.ReplicaStatus{},
						State:    &db.ModelStatus{},
					},
				},
			}},
			servers: []*db.Server{
				{
					Name: "server",
					Replicas: map[int32]*db.ServerReplica{
						0: {LoadedModels: []*db.ModelVersionID{}, ReservedMemory: memBytes, UniqueLoadedModels: map[string]bool{}},
						1: {LoadedModels: []*db.ModelVersionID{}, ReservedMemory: memBytes, UniqueLoadedModels: map[string]bool{}},
					},
				},
			},
			modelName:              "model",
			version:                1,
			serverKey:              "server",
			replicaIdx:             0,
			expectedState:          db.ModelReplicaState_ModelReplicaStateUnknown,
			desiredState:           db.ModelReplicaState_Unloaded,
			numModelVersionsLoaded: 0,
			modelVersionLoaded:     false,
			availableMemory:        20,
			modelRuntimeInfo:       &pb.ModelRuntimeInfo{ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{Mlserver: &pb.MLServerModelSettings{ParallelWorkers: uint32(1)}}},
		},
		{
			name: "Unloaded model but not matching expected state",
			models: []*db.Model{{
				Name: "model",
				Versions: []*db.ModelVersion{
					{
						ModelDefn: &pb.Model{
							Meta: &pb.MetaData{
								Name: "model",
							},
							ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
						Version: 1,
						Replicas: map[int32]*db.ReplicaStatus{
							0: {State: db.ModelReplicaState_LoadRequested},
						},
						State: &db.ModelStatus{},
					},
				},
			}},
			servers: []*db.Server{
				{
					Name: "server",
					Replicas: map[int32]*db.ServerReplica{
						0: {LoadedModels: []*db.ModelVersionID{}, ReservedMemory: memBytes, UniqueLoadedModels: map[string]bool{}},
						1: {LoadedModels: []*db.ModelVersionID{}, ReservedMemory: memBytes, UniqueLoadedModels: map[string]bool{}},
					},
				},
			},
			modelName:              "model",
			version:                1,
			serverKey:              "server",
			replicaIdx:             0,
			expectedState:          db.ModelReplicaState_Unloading,
			desiredState:           db.ModelReplicaState_Unloaded,
			numModelVersionsLoaded: 0,
			modelVersionLoaded:     false,
			availableMemory:        20,
			err:                    true,
		},
		{
			name: "DeletedModel",
			models: []*db.Model{{
				Name: "my-model",
				Versions: []*db.ModelVersion{
					{
						ModelDefn: &pb.Model{
							Meta: &pb.MetaData{
								Name: "my-model",
							},
							ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
						Version:  1,
						Replicas: map[int32]*db.ReplicaStatus{},
						State:    &db.ModelStatus{},
					},
				},
			}},
			servers: []*db.Server{
				{
					Name: "server",
					Replicas: map[int32]*db.ServerReplica{
						0: {LoadedModels: []*db.ModelVersionID{}, ReservedMemory: memBytes, UniqueLoadedModels: make(map[string]bool)},
						1: {LoadedModels: []*db.ModelVersionID{}, ReservedMemory: memBytes, UniqueLoadedModels: make(map[string]bool)},
					},
				},
			},
			modelName:              "my-model",
			version:                1,
			serverKey:              "server",
			replicaIdx:             0,
			expectedState:          db.ModelReplicaState_ModelReplicaStateUnknown,
			desiredState:           db.ModelReplicaState_Unloaded,
			numModelVersionsLoaded: 0,
			modelVersionLoaded:     false,
			availableMemory:        20,
			deleted:                true,
			modelRuntimeInfo:       &pb.ModelRuntimeInfo{ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{Mlserver: &pb.MLServerModelSettings{ParallelWorkers: uint32(1)}}},
		},
		{
			name: "Model updated but not latest on replica which is loaded",
			models: []*db.Model{{
				Name: "foo",
				Versions: []*db.ModelVersion{
					{
						ModelDefn: &pb.Model{
							Meta: &pb.MetaData{
								Name: "foo",
							},
							ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
						Version: 1,
						Replicas: map[int32]*db.ReplicaStatus{
							0: {State: db.ModelReplicaState_Unloading},
						},
						State: &db.ModelStatus{},
					},
					{
						ModelDefn: &pb.Model{
							Meta: &pb.MetaData{
								Name: "foo",
							},
							ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
						Version: 2,
						Replicas: map[int32]*db.ReplicaStatus{
							0: {State: db.ModelReplicaState_Loaded},
						},
						State: &db.ModelStatus{},
					},
				},
			}},
			servers: []*db.Server{
				{
					Name: "server",
					Replicas: map[int32]*db.ServerReplica{
						0: {LoadedModels: []*db.ModelVersionID{{Name: "foo", Version: 2}, {Name: "foo", Version: 1}}, ReservedMemory: memBytes, UniqueLoadedModels: map[string]bool{"foo": true}},
						1: {LoadedModels: []*db.ModelVersionID{}, ReservedMemory: memBytes, UniqueLoadedModels: map[string]bool{}},
					},
				},
			},
			modelName:              "foo",
			version:                1,
			serverKey:              "server",
			replicaIdx:             0,
			expectedState:          db.ModelReplicaState_Unloading,
			desiredState:           db.ModelReplicaState_Unloaded,
			numModelVersionsLoaded: 1,
			modelVersionLoaded:     false,
			availableMemory:        20,
			modelRuntimeInfo:       &pb.ModelRuntimeInfo{ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{Mlserver: &pb.MLServerModelSettings{ParallelWorkers: uint32(1)}}},
			err:                    false,
		},
		{
			name: "Model updated but not latest on replica which is Available",
			models: []*db.Model{{
				Name: "foo",
				Versions: []*db.ModelVersion{
					{
						ModelDefn: &pb.Model{
							Meta: &pb.MetaData{
								Name: "foo",
							},
							ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
						Version: 1,
						Replicas: map[int32]*db.ReplicaStatus{
							0: {State: db.ModelReplicaState_Unloading},
						},
						State: &db.ModelStatus{},
					},
					{
						ModelDefn: &pb.Model{
							Meta: &pb.MetaData{
								Name: "foo",
							},
							ModelSpec: &pb.ModelSpec{MemoryBytes: &memBytes}},
						Version: 2,
						Replicas: map[int32]*db.ReplicaStatus{
							0: {State: db.ModelReplicaState_Available},
						},
						State: &db.ModelStatus{},
					},
				},
			}},
			servers: []*db.Server{
				{
					Name: "server",
					Replicas: map[int32]*db.ServerReplica{
						0: {LoadedModels: []*db.ModelVersionID{{Name: "foo", Version: 2}, {Name: "foo", Version: 1}}, ReservedMemory: memBytes * 2, UniqueLoadedModels: map[string]bool{"foo": true}},
						1: {LoadedModels: []*db.ModelVersionID{{Name: "foo", Version: 2}, {Name: "foo", Version: 1}}, ReservedMemory: memBytes, UniqueLoadedModels: map[string]bool{}},
					},
				},
			},
			modelName:              "foo",
			version:                1,
			serverKey:              "server",
			replicaIdx:             0,
			expectedState:          db.ModelReplicaState_Unloading,
			desiredState:           db.ModelReplicaState_Unloaded,
			numModelVersionsLoaded: 1,
			modelVersionLoaded:     false,
			availableMemory:        20,
			modelRuntimeInfo:       &pb.ModelRuntimeInfo{ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{Mlserver: &pb.MLServerModelSettings{ParallelWorkers: uint32(2)}}},
			err:                    false,
		},
		{
			name: "Existing ModelRuntimeInfo is not overwritten",
			models: []*db.Model{{
				Name: "my-model",
				Versions: []*db.ModelVersion{
					{
						ModelDefn: &pb.Model{
							Meta: &pb.MetaData{
								Name: "my-model",
							},
							ModelSpec: &pb.ModelSpec{
								MemoryBytes: &memBytes,
								ModelRuntimeInfo: &pb.ModelRuntimeInfo{
									ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{
										Mlserver: &pb.MLServerModelSettings{
											ParallelWorkers: uint32(2),
										},
									},
								},
							},
						},
						Version:  1,
						Replicas: map[int32]*db.ReplicaStatus{},
						State:    &db.ModelStatus{},
					},
				},
			}},
			servers: []*db.Server{
				{
					Name: "server",
					Replicas: map[int32]*db.ServerReplica{
						0: {LoadedModels: []*db.ModelVersionID{}, ReservedMemory: memBytes, UniqueLoadedModels: map[string]bool{}},
						1: {LoadedModels: []*db.ModelVersionID{}, ReservedMemory: memBytes, UniqueLoadedModels: map[string]bool{}},
					},
				},
			},
			modelName:              "my-model",
			version:                1,
			serverKey:              "server",
			replicaIdx:             0,
			expectedState:          db.ModelReplicaState_ModelReplicaStateUnknown,
			desiredState:           db.ModelReplicaState_Loaded,
			numModelVersionsLoaded: 1,
			modelVersionLoaded:     true,
			availableMemory:        20,
			modelRuntimeInfo:       &pb.ModelRuntimeInfo{ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{Mlserver: &pb.MLServerModelSettings{ParallelWorkers: uint32(500)}}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			ctx := context.Background()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())

			// Create storage instances
			modelStorage := NewInMemoryStorage[*db.Model]()
			serverStorage := NewInMemoryStorage[*db.Server]()

			// Populate storage with test data
			for _, model := range test.models {
				if test.deleted {
					model.Deleted = true
				}
				err := modelStorage.Insert(ctx, model)
				g.Expect(err).To(BeNil())
			}
			for _, server := range test.servers {
				err := serverStorage.Insert(ctx, server)
				g.Expect(err).To(BeNil())
			}

			var expectedModelRuntimeInfo *pb.ModelRuntimeInfo
			model, err := modelStorage.Get(ctx, test.modelName)
			g.Expect(err).To(BeNil())

			if model.GetVersion(test.version).ModelDefn.ModelSpec.ModelRuntimeInfo != nil {
				expectedModelRuntimeInfo = model.GetVersion(test.version).ModelDefn.ModelSpec.ModelRuntimeInfo
			} else {
				expectedModelRuntimeInfo = test.modelRuntimeInfo
			}

			var modelEvt *coordinator.ModelEventMsg
			muModelEvt := &sync.Mutex{}

			eventHub.RegisterModelEventHandler(
				"handler-model",
				10,
				logger,
				func(event coordinator.ModelEventMsg) {
					muModelEvt.Lock()
					modelEvt = &event
					muModelEvt.Unlock()
				},
			)

			var serverEvt *coordinator.ServerEventMsg
			muServerEvt := &sync.Mutex{}

			eventHub.RegisterServerEventHandler(
				"handler-server",
				10,
				logger,
				func(event coordinator.ServerEventMsg) {
					if event.UpdateContext == coordinator.SERVER_SCALE_DOWN {
						muServerEvt.Lock()
						serverEvt = &event
						muServerEvt.Unlock()
					}
				},
			)

			getModel := func(name string) *db.Model {
				model, err := modelStorage.Get(ctx, name)
				g.Expect(err).To(BeNil())
				return model
			}

			getServer := func(name string) *db.Server {
				server, err := serverStorage.Get(ctx, name)
				g.Expect(err).To(BeNil())
				return server
			}

			ms := NewModelServerStore(logger, modelStorage, serverStorage, eventHub)
			err = ms.UpdateModelState(test.modelName, test.version, test.serverKey, test.replicaIdx, &test.availableMemory, test.expectedState, test.desiredState, "", test.modelRuntimeInfo)
			if !test.err {
				g.Expect(err).To(BeNil())
				if !test.deleted {
					g.Expect(getModel(test.modelName).GetVersion(test.version).GetModelReplicaState(test.replicaIdx)).To(Equal(test.desiredState))

					found := false
					for _, v := range getServer(test.serverKey).Replicas[int32(test.replicaIdx)].LoadedModels {
						if v.Name == test.modelName && v.Version == test.version && test.modelVersionLoaded {
							found = true
						}
					}
					g.Expect(found).To(Equal(test.modelVersionLoaded))

					g.Expect(getServer(test.serverKey).Replicas[int32(test.replicaIdx)].GetNumLoadedModels()).To(Equal(test.numModelVersionsLoaded))
				} else {
					g.Expect(getModel(test.modelName).Latest().State.State).To(Equal(db.ModelState_ModelTerminated))
				}

				if expectedModelRuntimeInfo != nil {
					g.Expect(getModel(test.modelName).GetVersion(test.version).ModelDefn.ModelSpec.ModelRuntimeInfo).To(Equal(expectedModelRuntimeInfo))
				}
			} else {
				g.Expect(err).ToNot(BeNil())
			}

			if test.desiredState == db.ModelReplicaState_LoadFailed || test.desiredState == db.ModelReplicaState_Loaded {
				g.Expect(getServer(test.serverKey).Replicas[int32(test.replicaIdx)].GetReservedMemory()).To(Equal(uint64(0)))
			} else {
				g.Expect(getServer(test.serverKey).Replicas[int32(test.replicaIdx)].GetReservedMemory()).To(Equal(getModel(test.modelName).GetVersion(test.version).GetRequiredMemory()))
			}

			toUniqueModels := func(loadedModels []*db.ModelVersionID) map[string]bool {
				if loadedModels == nil {
					return nil
				}
				uniqueModels := make(map[string]bool)
				for _, key := range loadedModels {
					uniqueModels[key.Name] = true
				}
				return uniqueModels
			}

			uniqueLoadedModels := toUniqueModels(getServer(test.serverKey).Replicas[int32(test.replicaIdx)].LoadedModels)
			g.Expect(uniqueLoadedModels).To(Equal(getServer(test.serverKey).Replicas[int32(test.replicaIdx)].UniqueLoadedModels))

			// allow events to propagate
			time.Sleep(500 * time.Millisecond)

			if !test.err {
				muModelEvt.Lock()
				g.Expect(modelEvt).ToNot(BeNil())
				g.Expect(modelEvt.ModelVersion).To(Equal(test.version))
				muModelEvt.Unlock()
			}
			if test.name == "DeletedModel" {
				muServerEvt.Lock()
				g.Expect(serverEvt).ToNot(BeNil())
				g.Expect(serverEvt.UpdateContext).To(Equal(coordinator.SERVER_SCALE_DOWN))
				g.Expect(serverEvt.ServerName).To(Equal(test.serverKey))
				muServerEvt.Unlock()
			}
		})
	}
}

func TestUpdateModelStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                      string
		deleted                   bool
		modelVersion              *db.ModelVersion
		prevAvailableModelVersion *db.ModelVersion
		expectedState             db.ModelState
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
			modelVersion: &db.ModelVersion{
				Server:    "server",
				ModelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
				Version:   1,
				Replicas: map[int32]*db.ReplicaStatus{
					0: {State: db.ModelReplicaState_Available, Timestamp: timestamppb.New(d1)},
				},
				State: &db.ModelStatus{State: db.ModelState_ModelProgressing},
			},
			prevAvailableModelVersion: nil,
			expectedState:             db.ModelState_ModelAvailable,
			expectedAvailableReplicas: 1,
			expectedReason:            "",
			expectedTimestamp:         d1,
		},
		{
			name:    "Scaled Down",
			deleted: false,
			modelVersion: &db.ModelVersion{
				Server:    "server",
				ModelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 0}},
				Version:   1,
				Replicas: map[int32]*db.ReplicaStatus{
					0: {State: db.ModelReplicaState_Unloaded, Timestamp: timestamppb.New(d1)},
				},
				State: &db.ModelStatus{State: db.ModelState_ModelProgressing},
			},
			prevAvailableModelVersion: nil,
			expectedState:             db.ModelState_ModelScaledDown,
			expectedAvailableReplicas: 0,
			expectedReason:            "",
			expectedTimestamp:         d1,
		},
		{
			name:    "Progressing",
			deleted: false,
			modelVersion: &db.ModelVersion{
				Server:    "server",
				ModelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 2}},
				Version:   1,
				Replicas: map[int32]*db.ReplicaStatus{
					0: {State: db.ModelReplicaState_Available, Timestamp: timestamppb.New(d1)},
					1: {State: db.ModelReplicaState_Loading, Timestamp: timestamppb.New(d1)},
				},
				State: &db.ModelStatus{State: db.ModelState_ModelProgressing},
			},
			prevAvailableModelVersion: nil,
			expectedState:             db.ModelState_ModelProgressing,
			expectedAvailableReplicas: 1,
			expectedReason:            "",
			expectedTimestamp:         d1,
		},
		{
			name:    "Failed",
			deleted: false,
			modelVersion: &db.ModelVersion{
				Server:    "server",
				ModelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
				Version:   1,
				Replicas: map[int32]*db.ReplicaStatus{
					0: {State: db.ModelReplicaState_LoadFailed, Reason: r1, Timestamp: timestamppb.New(d1)},
				},
				State: &db.ModelStatus{State: db.ModelState_ModelProgressing},
			},
			prevAvailableModelVersion: nil,
			expectedState:             db.ModelState_ModelFailed,
			expectedAvailableReplicas: 0,
			expectedReason:            r1,
			expectedTimestamp:         d1,
		},
		{
			name:    "AvailableAndFailed",
			deleted: false,
			modelVersion: &db.ModelVersion{
				Server:    "server",
				ModelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 2}},
				Version:   1,
				Replicas: map[int32]*db.ReplicaStatus{
					0: {State: db.ModelReplicaState_Loaded, Timestamp: timestamppb.New(d1)},
					1: {State: db.ModelReplicaState_LoadFailed, Reason: r1, Timestamp: timestamppb.New(d2)},
				},
				State: &db.ModelStatus{State: db.ModelState_ModelProgressing},
			},
			prevAvailableModelVersion: nil,
			expectedState:             db.ModelState_ModelFailed,
			expectedAvailableReplicas: 0,
			expectedReason:            r1,
			expectedTimestamp:         d2,
		},
		{
			name:    "TwoFailed",
			deleted: false,
			modelVersion: &db.ModelVersion{
				Server:    "server",
				ModelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 2}},
				Version:   1,
				Replicas: map[int32]*db.ReplicaStatus{
					0: {State: db.ModelReplicaState_LoadFailed, Reason: r1, Timestamp: timestamppb.New(d1)},
					1: {State: db.ModelReplicaState_LoadFailed, Reason: r2, Timestamp: timestamppb.New(d2)},
				},
				State: &db.ModelStatus{State: db.ModelState_ModelProgressing},
			},
			prevAvailableModelVersion: nil,
			expectedState:             db.ModelState_ModelFailed,
			expectedAvailableReplicas: 0,
			expectedReason:            r2,
			expectedTimestamp:         d2,
		},
		{
			name:    "AvailableV2",
			deleted: false,
			modelVersion: &db.ModelVersion{
				Server:    "server",
				ModelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
				Version:   1,
				Replicas: map[int32]*db.ReplicaStatus{
					0: {State: db.ModelReplicaState_Loading, Timestamp: timestamppb.New(d1)},
					1: {State: db.ModelReplicaState_Available, Timestamp: timestamppb.New(d2)},
				},
				State: &db.ModelStatus{State: db.ModelState_ModelProgressing},
			},
			prevAvailableModelVersion: &db.ModelVersion{
				Server:    "server",
				ModelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
				Version:   1,
				Replicas: map[int32]*db.ReplicaStatus{
					0: {State: db.ModelReplicaState_Available, Timestamp: timestamppb.New(d1)},
				},
				State: &db.ModelStatus{State: db.ModelState_ModelAvailable},
			},
			expectedState:             db.ModelState_ModelAvailable,
			expectedAvailableReplicas: 1,
			expectedReason:            "",
			expectedTimestamp:         d2,
		},
		{
			name:    "Terminating",
			deleted: true,
			modelVersion: &db.ModelVersion{
				Server:    "server",
				ModelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 2}},
				Version:   1,
				Replicas: map[int32]*db.ReplicaStatus{
					0: {State: db.ModelReplicaState_Unloading, Timestamp: timestamppb.New(d1)},
					1: {State: db.ModelReplicaState_Unloading, Timestamp: timestamppb.New(d2)},
				},
				State: &db.ModelStatus{State: db.ModelState_ModelProgressing},
			},
			expectedState:             db.ModelState_ModelTerminating,
			expectedAvailableReplicas: 0,
			expectedReason:            "",
			expectedTimestamp:         d2,
		},
		{
			name:    "TerminatingLoadingReplicas",
			deleted: true,
			modelVersion: &db.ModelVersion{
				Server:    "server",
				ModelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 2}},
				Version:   2,
				Replicas: map[int32]*db.ReplicaStatus{
					0: {State: db.ModelReplicaState_Loading, Timestamp: timestamppb.New(d1)},
					1: {State: db.ModelReplicaState_Loading, Timestamp: timestamppb.New(d2)},
				},
				State: &db.ModelStatus{State: db.ModelState_ModelProgressing},
			},
			expectedState:             db.ModelState_ModelTerminating,
			expectedAvailableReplicas: 0,
			expectedReason:            "",
			expectedTimestamp:         d2,
		},
		{
			name:    "Terminated",
			deleted: true,
			modelVersion: &db.ModelVersion{
				Server:    "server",
				ModelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 2}},
				Version:   1,
				Replicas: map[int32]*db.ReplicaStatus{
					0: {State: db.ModelReplicaState_Unloaded, Timestamp: timestamppb.New(d1)},
					1: {State: db.ModelReplicaState_Unloaded, Timestamp: timestamppb.New(d2)},
				},
				State: &db.ModelStatus{State: db.ModelState_ModelProgressing},
			},
			expectedState:             db.ModelState_ModelTerminated,
			expectedAvailableReplicas: 0,
			expectedReason:            "",
			expectedTimestamp:         d2,
		},
		{
			name:    "TerminateFailed",
			deleted: true,
			modelVersion: &db.ModelVersion{
				Server:    "server",
				ModelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 2}},
				Version:   1,
				Replicas: map[int32]*db.ReplicaStatus{
					0: {State: db.ModelReplicaState_UnloadFailed, Reason: r1, Timestamp: timestamppb.New(d1)},
					1: {State: db.ModelReplicaState_Unloaded, Timestamp: timestamppb.New(d2)},
				},
				State: &db.ModelStatus{State: db.ModelState_ModelProgressing},
			},
			expectedState:             db.ModelState_ModelTerminateFailed,
			expectedAvailableReplicas: 0,
			expectedReason:            r1,
			expectedTimestamp:         d1,
		},
		{
			name:    "AvailableV2PrevTerminated",
			deleted: false,
			modelVersion: &db.ModelVersion{
				Server:    "server",
				ModelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 2}},
				Version:   2,
				Replicas: map[int32]*db.ReplicaStatus{
					0: {State: db.ModelReplicaState_Available, Timestamp: timestamppb.New(d1)},
					1: {State: db.ModelReplicaState_Available, Timestamp: timestamppb.New(d2)},
				},
				State: &db.ModelStatus{State: db.ModelState_ModelProgressing},
			},
			prevAvailableModelVersion: &db.ModelVersion{
				Server:    "server",
				ModelDefn: &pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 2}},
				Version:   2,
				Replicas: map[int32]*db.ReplicaStatus{
					0: {State: db.ModelReplicaState_Available, Timestamp: timestamppb.New(d1)},
					1: {State: db.ModelReplicaState_Available, Timestamp: timestamppb.New(d2)},
				},
				State: &db.ModelStatus{State: db.ModelState_ModelTerminating},
			},
			expectedState:             db.ModelState_ModelAvailable,
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

			const modelName = "some-model"

			modelStore := NewInMemoryStorage[*db.Model]()
			err = modelStore.Insert(context.TODO(), &db.Model{Name: modelName})
			g.Expect(err).To(BeNil())

			ms := NewModelServerStore(logger, modelStore, NewInMemoryStorage[*db.Server](), eventHub)
			err = ms.updateModelStatus(true, test.deleted, test.modelVersion, test.prevAvailableModelVersion, &db.Model{
				Name: modelName,
			})
			g.Expect(err).To(BeNil())
			g.Expect(test.modelVersion.State.State).To(Equal(test.expectedState))
			g.Expect(test.modelVersion.State.Reason).To(Equal(test.expectedReason))
			g.Expect(test.modelVersion.State.AvailableReplicas).To(Equal(test.expectedAvailableReplicas))
			g.Expect(test.modelVersion.State.Timestamp.AsTime()).To(Equal(test.expectedTimestamp))
		})
	}
}

func TestAddModelVersionIfNotExists(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		models       []*db.Model
		modelVersion *agent.ModelVersion
		expected     []*db.ModelVersion
		latest       uint32
	}

	tests := []test{
		{
			name:   "Add new version when none exist",
			models: []*db.Model{},
			modelVersion: &agent.ModelVersion{
				Version: 1,
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo"},
				},
			},
			expected: []*db.ModelVersion{
				{
					Version:   1,
					ModelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo"}},
					State:     &db.ModelStatus{ModelGwState: db.ModelState_ModelCreate},
				},
			},
			latest: 1,
		},
		{
			name: "AddNewVersion",
			models: []*db.Model{
				{
					Name:     "foo",
					Versions: []*db.ModelVersion{},
				},
			},
			modelVersion: &agent.ModelVersion{
				Version: 1,
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo"},
				},
			},
			expected: []*db.ModelVersion{
				{
					Version:   1,
					ModelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo"}},
					State:     &db.ModelStatus{ModelGwState: db.ModelState_ModelCreate},
				},
			},
			latest: 1,
		},
		{
			name: "AddSecondVersion",
			models: []*db.Model{
				{
					Name: "foo",
					Versions: []*db.ModelVersion{
						{
							Version:   1,
							ModelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo"}},
							Replicas:  map[int32]*db.ReplicaStatus{},
							State:     &db.ModelStatus{},
						},
					},
				},
			},
			modelVersion: &agent.ModelVersion{
				Version: 2,
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo"},
				},
			},
			expected: []*db.ModelVersion{
				{
					Version:   1,
					ModelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo"}},
					State:     &db.ModelStatus{},
				},
				{
					Version:   2,
					ModelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo"}},
					State:     &db.ModelStatus{ModelGwState: db.ModelState_ModelCreate},
				},
			},
			latest: 2,
		},
		{
			name: "Existing",
			models: []*db.Model{
				{
					Name: "foo",
					Versions: []*db.ModelVersion{
						{
							Version:   1,
							ModelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo"}},
							Replicas:  map[int32]*db.ReplicaStatus{},
							State:     &db.ModelStatus{},
						},
					},
				},
			},
			modelVersion: &agent.ModelVersion{
				Version: 1,
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo"},
				},
			},
			expected: []*db.ModelVersion{
				{
					Version:   1,
					ModelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo"}},
					State:     &db.ModelStatus{},
				},
			},
			latest: 1,
		},
		{
			name: "AddThirdVersion",
			models: []*db.Model{
				{
					Name: "foo",
					Versions: []*db.ModelVersion{
						{
							Version:   1,
							ModelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo"}},
							Replicas:  map[int32]*db.ReplicaStatus{},
							State:     &db.ModelStatus{},
						},
						{
							Version:   2,
							ModelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo"}},
							Replicas:  map[int32]*db.ReplicaStatus{},
							State:     &db.ModelStatus{},
						},
					},
				},
			},
			modelVersion: &agent.ModelVersion{
				Version: 3,
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo"},
				},
			},
			expected: []*db.ModelVersion{
				{
					Version:   1,
					ModelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo"}},
					State:     &db.ModelStatus{},
				},
				{
					Version:   2,
					ModelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo"}},
					State:     &db.ModelStatus{},
				},
				{
					Version:   3,
					ModelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo"}},
					State:     &db.ModelStatus{ModelGwState: db.ModelState_ModelCreate},
				},
			},
			latest: 3,
		},
		{
			name: "AddThirdVersionInMiddle",
			models: []*db.Model{
				{
					Name: "foo",
					Versions: []*db.ModelVersion{
						{
							Version:   1,
							ModelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo"}},
							Replicas:  map[int32]*db.ReplicaStatus{},
							State:     &db.ModelStatus{},
						},
						{
							Version:   3,
							ModelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo"}},
							Replicas:  map[int32]*db.ReplicaStatus{},
							State:     &db.ModelStatus{},
						},
					},
				},
			},
			modelVersion: &agent.ModelVersion{
				Version: 2,
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo"},
				},
			},
			expected: []*db.ModelVersion{
				{
					Version:   1,
					ModelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo"}},
					State:     &db.ModelStatus{},
				},
				{
					Version:   2,
					ModelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo"}},
					State:     &db.ModelStatus{ModelGwState: db.ModelState_ModelCreate},
				},
				{
					Version:   3,
					ModelDefn: &pb.Model{Meta: &pb.MetaData{Name: "foo"}},
					State:     &db.ModelStatus{},
				},
			},
			latest: 3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())

			modelStorage := NewInMemoryStorage[*db.Model]()
			serverStorage := NewInMemoryStorage[*db.Server]()

			// Populate storage with test data
			for _, model := range test.models {
				err := modelStorage.Insert(context.TODO(), model)
				g.Expect(err).To(BeNil())
			}

			ms := NewModelServerStore(logger, modelStorage, serverStorage, eventHub)
			model, _, err := ms.addModelVersionIfNotExists(test.modelVersion)
			g.Expect(err).To(BeNil())
			g.Expect(len(model.Versions)).To(Equal(len(test.expected)))
			for i, modelVersion := range model.Versions {
				g.Expect(proto.Equal(modelVersion, test.expected[i])).To(BeTrue())
			}
			g.Expect(model.Latest().Version).To(Equal(test.latest))
		})
	}
}

func TestAddServerReplica(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()

	type test struct {
		name                 string
		models               []*db.Model
		servers              []*db.Server
		req                  *agent.AgentSubscribeRequest
		expectedSnapshot     []*db.Server
		expectedModelEvents  int64
		expectedServerEvents int64
	}

	tests := []test{
		{
			name:   "AddServerReplica - existing server",
			models: []*db.Model{},
			servers: []*db.Server{
				{
					Name: "server1",
					Replicas: map[int32]*db.ServerReplica{
						0: {},
						1: {},
					},
					ExpectedReplicas: 3,
					Shared:           true,
				},
			},
			req: &agent.AgentSubscribeRequest{
				ServerName: "server1",
				ReplicaIdx: 2,
				Shared:     true,
			},
			expectedSnapshot: []*db.Server{
				{
					Name: "server1",
					Replicas: map[int32]*db.ServerReplica{
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
			name:    "AddServerReplica - new server",
			models:  []*db.Model{},
			servers: []*db.Server{},
			req: &agent.AgentSubscribeRequest{
				ServerName: "server1",
				ReplicaIdx: 0,
				Shared:     true,
			},
			expectedSnapshot: []*db.Server{
				{
					Name: "server1",
					Replicas: map[int32]*db.ServerReplica{
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
			name:    "AddServerReplica - with loaded models",
			models:  []*db.Model{},
			servers: []*db.Server{},
			req: &agent.AgentSubscribeRequest{
				ServerName: "server1",
				ReplicaIdx: 0,
				Shared:     true,
				LoadedModels: []*agent.ModelVersion{
					{
						Model: &pb.Model{
							Meta:      &pb.MetaData{Name: "model1"},
							ModelSpec: &pb.ModelSpec{},
						},
						Version: 1,
					},
					{
						Model: &pb.Model{
							Meta:      &pb.MetaData{Name: "model2"},
							ModelSpec: &pb.ModelSpec{},
						},
						Version: 1,
					},
				},
			},
			expectedSnapshot: []*db.Server{
				{
					Name: "server1",
					Replicas: map[int32]*db.ServerReplica{
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

			// Create storage instances
			modelStorage := NewInMemoryStorage[*db.Model]()
			serverStorage := NewInMemoryStorage[*db.Server]()

			// Populate storage with test data
			for _, model := range test.models {
				err := modelStorage.Insert(ctx, model)
				g.Expect(err).To(BeNil())
			}
			for _, server := range test.servers {
				err := serverStorage.Insert(ctx, server)
				g.Expect(err).To(BeNil())
			}

			// Create MemoryStore with populated storage
			ms := NewModelServerStore(logger, modelStorage, serverStorage, eventHub)

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
			actualSnapshot, err := ms.GetServers()
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
	ctx := context.Background()

	type test struct {
		name           string
		models         []*db.Model
		servers        []*db.Server
		serverName     string
		replicaIdx     int
		serverExists   bool
		modelsReturned int
	}

	tests := []test{
		{
			name:   "ReplicaRemovedButNotDeleted",
			models: []*db.Model{},
			servers: []*db.Server{
				{
					Name: "server1",
					Replicas: map[int32]*db.ServerReplica{
						0: {
							LoadedModels: []*db.ModelVersionID{{Name: "model1", Version: 1}},
						},
						1: {},
					},
					ExpectedReplicas: 2,
					Shared:           true,
				},
			},
			serverName:     "server1",
			replicaIdx:     0,
			serverExists:   true,
			modelsReturned: 0, // no models really defined in store
		},
		{
			name: "ReplicaRemovedAndDeleted",
			models: []*db.Model{
				{
					Name: "model1",
					Versions: []*db.ModelVersion{
						{
							Version:  1,
							Replicas: make(map[int32]*db.ReplicaStatus),
						},
					},
				},
			},
			servers: []*db.Server{
				{
					Name: "server1",
					Replicas: map[int32]*db.ServerReplica{
						0: {
							LoadedModels: []*db.ModelVersionID{{Name: "model1", Version: 1}},
						},
						1: {},
					},
					ExpectedReplicas: -1,
					Shared:           true,
				},
			},
			serverName:     "server1",
			replicaIdx:     0,
			serverExists:   true,
			modelsReturned: 1,
		},
		{
			name: "ReplicaRemovedAndServerDeleted",
			models: []*db.Model{
				{
					Name: "model1",
					Versions: []*db.ModelVersion{
						{
							Version:  1,
							Replicas: make(map[int32]*db.ReplicaStatus),
						},
					},
				},
			},
			servers: []*db.Server{
				{
					Name: "server1",
					Replicas: map[int32]*db.ServerReplica{
						0: {
							LoadedModels: []*db.ModelVersionID{{Name: "model1", Version: 1}},
						},
					},
					ExpectedReplicas: 0,
					Shared:           true,
				},
			},
			serverName:     "server1",
			replicaIdx:     0,
			serverExists:   false,
			modelsReturned: 1,
		},
		{
			name:   "ReplicaRemovedAndServerDeleted but no model version in store",
			models: []*db.Model{},
			servers: []*db.Server{
				{
					Name: "server1",
					Replicas: map[int32]*db.ServerReplica{
						0: {
							LoadedModels: []*db.ModelVersionID{{Name: "model1", Version: 1}},
						},
					},
					ExpectedReplicas: 0,
					Shared:           true,
				},
			},
			serverName:     "server1",
			replicaIdx:     0,
			serverExists:   false,
			modelsReturned: 0,
		},
		{
			name: "ReplicaRemovedAndDeleted - loading models",
			models: []*db.Model{
				{
					Name: "model1",
					Versions: []*db.ModelVersion{
						{
							Version:  1,
							Replicas: make(map[int32]*db.ReplicaStatus),
						},
					},
				},
				{
					Name: "model2",
					Versions: []*db.ModelVersion{
						{
							Version:  1,
							Replicas: make(map[int32]*db.ReplicaStatus),
						},
					},
				},
			},
			servers: []*db.Server{
				{
					Name: "server1",
					Replicas: map[int32]*db.ServerReplica{
						0: {
							LoadedModels: []*db.ModelVersionID{
								{Name: "model1", Version: 1},
								{Name: "model2", Version: 1},
							},
						},
						1: {},
					},
					ExpectedReplicas: -1,
					Shared:           true,
				},
			},
			serverName:     "server1",
			replicaIdx:     0,
			serverExists:   true,
			modelsReturned: 2,
		},
		{
			name: "ReplicaRemovedAndDeleted - non latest models",
			models: []*db.Model{
				{
					Name: "model1",
					Versions: []*db.ModelVersion{
						{
							Version: 1,
							Replicas: map[int32]*db.ReplicaStatus{
								0: {State: db.ModelReplicaState_Loaded},
							},
							State:     &db.ModelStatus{},
							ModelDefn: &pb.Model{Meta: &pb.MetaData{Name: "model1"}},
						},
						{
							Version: 2,
							Replicas: map[int32]*db.ReplicaStatus{
								0: {State: db.ModelReplicaState_LoadFailed},
							},
							State:     &db.ModelStatus{},
							ModelDefn: &pb.Model{Meta: &pb.MetaData{Name: "model1"}},
						},
					},
				},
			},
			servers: []*db.Server{
				{
					Name: "server1",
					Replicas: map[int32]*db.ServerReplica{
						0: {
							LoadedModels: []*db.ModelVersionID{{Name: "model1", Version: 1}},
						},
					},
					ExpectedReplicas: -1,
					Shared:           true,
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

			// Create storage instances
			modelStorage := NewInMemoryStorage[*db.Model]()
			serverStorage := NewInMemoryStorage[*db.Server]()

			// Populate storage with test data
			for _, model := range test.models {
				err := modelStorage.Insert(ctx, model)
				g.Expect(err).To(BeNil())
			}
			for _, server := range test.servers {
				err := serverStorage.Insert(ctx, server)
				g.Expect(err).To(BeNil())
			}

			// Create MemoryStore with populated storage
			ms := NewModelServerStore(logger, modelStorage, serverStorage, eventHub)

			models, err := ms.RemoveServerReplica(test.serverName, test.replicaIdx)
			g.Expect(err).To(BeNil())
			g.Expect(test.modelsReturned).To(Equal(len(models)))
			server, _, err := ms.GetServer(test.serverName, false)
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
	ctx := context.Background()

	type test struct {
		name           string
		models         []*db.Model
		servers        []*db.Server
		serverName     string
		replicaIdx     int
		modelsReturned []string
	}

	// if we have models returned check status is Draining
	tests := []test{
		{
			name: "ReplicaSetDrainingNoModels",
			servers: []*db.Server{
				{
					Name: "server1",
					Replicas: map[int32]*db.ServerReplica{
						0: {},
						1: {},
					},
					ExpectedReplicas: 2,
					Shared:           true,
				},
			},
			serverName:     "server1",
			replicaIdx:     0,
			modelsReturned: []string{},
		},
		{
			name: "ReplicaSetDrainingWithLoadedModels",
			models: []*db.Model{
				{
					Name: "model1",
					Versions: []*db.ModelVersion{
						{
							Version:  1,
							Replicas: map[int32]*db.ReplicaStatus{0: {State: db.ModelReplicaState_Loaded}},
						},
					},
				},
			},
			servers: []*db.Server{
				{
					Name: "server1",
					Replicas: map[int32]*db.ServerReplica{
						0: {LoadedModels: []*db.ModelVersionID{{Name: "model1", Version: 1}}},
						1: {},
					},
					ExpectedReplicas: -1,
					Shared:           true,
				},
			},
			serverName:     "server1",
			replicaIdx:     0,
			modelsReturned: []string{"model1"},
		},
		{
			name: "ReplicaSetDrainingWithLoadedAndLoadingModels",
			models: []*db.Model{
				{
					Name: "model1",
					Versions: []*db.ModelVersion{
						{
							Version: 1,
							Replicas: map[int32]*db.ReplicaStatus{
								0: {State: db.ModelReplicaState_Loaded}},
						},
					},
				},
				{
					Name: "model2",
					Versions: []*db.ModelVersion{
						{
							Version: 1,
							Replicas: map[int32]*db.ReplicaStatus{
								0: {State: db.ModelReplicaState_Loading}},
						},
					},
				},
			},
			servers: []*db.Server{
				{
					Name: "server1",
					Replicas: map[int32]*db.ServerReplica{
						0: {
							LoadedModels: []*db.ModelVersionID{
								{Name: "model1", Version: 1},
							},
							LoadingModels: []*db.ModelVersionID{
								{Name: "model2", Version: 1},
							},
						},
						1: {},
					},
					ExpectedReplicas: -1,
					Shared:           true,
				},
			},
			serverName:     "server1",
			replicaIdx:     0,
			modelsReturned: []string{"model1", "model2"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())

			// Create storage instances
			modelStorage := NewInMemoryStorage[*db.Model]()
			serverStorage := NewInMemoryStorage[*db.Server]()

			// Populate storage with test data
			for _, model := range test.models {
				err := modelStorage.Insert(ctx, model)
				g.Expect(err).To(BeNil())
			}
			for _, server := range test.servers {
				err := serverStorage.Insert(ctx, server)
				g.Expect(err).To(BeNil())
			}

			// Create MemoryStore with populated storage
			ms := NewModelServerStore(logger, modelStorage, serverStorage, eventHub)

			models, err := ms.DrainServerReplica(test.serverName, test.replicaIdx)
			g.Expect(err).To(BeNil())
			g.Expect(test.modelsReturned).To(Equal(models))
			server, _, err := ms.GetServer(test.serverName, false)
			g.Expect(err).To(BeNil())
			g.Expect(server).ToNot(BeNil())
			g.Expect(server.Replicas[int32(test.replicaIdx)].GetIsDraining()).To(BeTrue())

			if test.modelsReturned != nil {
				for _, model := range test.modelsReturned {
					dbModel, _ := ms.GetModel(model)
					state := dbModel.Latest().GetModelReplicaState(test.replicaIdx)
					g.Expect(state).To(Equal(db.ModelReplicaState_Draining))
				}
			}
		})
	}
}
