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
	"testing"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
)

func TestUpdateStatus(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()

	type test struct {
		name                string
		models              []*db.Model
		servers             []*db.Server
		modelName           string
		serverName          string
		version             uint32
		prevVersion         *uint32
		expectedModelStatus db.ModelState
	}
	prevVersion := uint32(1)
	tests := []test{
		{
			name: "Available",
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
								ModelSpec: &pb.ModelSpec{},
								DeploymentSpec: &pb.DeploymentSpec{
									Replicas: 1,
								},
							},
							Server: "server2",
							State:  &db.ModelStatus{},
							Replicas: map[int32]*db.ReplicaStatus{
								0: {State: db.ModelReplicaState_Loaded},
							},
						},
						{
							Version: 2,
							ModelDefn: &pb.Model{
								Meta: &pb.MetaData{
									Name: "model",
								},
								ModelSpec: &pb.ModelSpec{},
								DeploymentSpec: &pb.DeploymentSpec{
									Replicas: 1,
								},
							},
							Server: "server1",
							State:  &db.ModelStatus{},
							Replicas: map[int32]*db.ReplicaStatus{
								0: {State: db.ModelReplicaState_Available},
							},
						},
					},
				},
			},
			servers: []*db.Server{
				{
					Name: "server1",
					Replicas: map[int32]*db.ServerReplica{
						0: {},
					},
				},
				{
					Name: "server2",
					Replicas: map[int32]*db.ServerReplica{
						0: {},
					},
				},
			},
			modelName:           "model",
			serverName:          "server2",
			version:             2,
			prevVersion:         nil,
			expectedModelStatus: db.ModelState_ModelAvailable,
		},
		{
			name: "Available - Min replicas",
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
								ModelSpec: &pb.ModelSpec{},
								DeploymentSpec: &pb.DeploymentSpec{
									Replicas: 1,
								},
							},
							State:  &db.ModelStatus{},
							Server: "server2",
							Replicas: map[int32]*db.ReplicaStatus{
								0: {State: db.ModelReplicaState_Loaded},
							},
						},
						{
							Version: 2,
							ModelDefn: &pb.Model{
								Meta: &pb.MetaData{
									Name: "model",
								},
								ModelSpec: &pb.ModelSpec{},
								DeploymentSpec: &pb.DeploymentSpec{
									Replicas:    2,
									MinReplicas: 1,
								},
							},
							State:  &db.ModelStatus{},
							Server: "server1",
							Replicas: map[int32]*db.ReplicaStatus{
								0: {State: db.ModelReplicaState_Available},
							},
						},
					},
				},
			},
			servers: []*db.Server{
				{
					Name: "server1",
					Replicas: map[int32]*db.ServerReplica{
						0: {},
					},
				},
				{
					Name: "server2",
					Replicas: map[int32]*db.ServerReplica{
						0: {},
					},
				},
			},
			modelName:           "model",
			serverName:          "server2",
			version:             2,
			prevVersion:         nil,
			expectedModelStatus: db.ModelState_ModelAvailable,
		},
		{
			name: "NotEnoughReplicasButPreviousAvailable",
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
								ModelSpec: &pb.ModelSpec{},
								DeploymentSpec: &pb.DeploymentSpec{
									Replicas: 1,
								},
							},
							Server: "server2",
							Replicas: map[int32]*db.ReplicaStatus{
								0: {State: db.ModelReplicaState_Available},
							},
							State: &db.ModelStatus{State: db.ModelState_ModelAvailable},
						},
						{
							Version: 2,
							ModelDefn: &pb.Model{
								Meta: &pb.MetaData{
									Name: "model",
								},
								ModelSpec: &pb.ModelSpec{},
								DeploymentSpec: &pb.DeploymentSpec{
									Replicas: 2,
								},
							},
							Server: "server1",
							State:  &db.ModelStatus{},
							Replicas: map[int32]*db.ReplicaStatus{
								0: {State: db.ModelReplicaState_Available},
							},
						},
					},
				},
			},
			servers: []*db.Server{
				{
					Name: "server1",
					Replicas: map[int32]*db.ServerReplica{
						0: {},
					},
				},
				{
					Name: "server2",
					Replicas: map[int32]*db.ServerReplica{
						0: {},
					},
				},
			},
			modelName:           "model",
			serverName:          "server2",
			version:             2,
			prevVersion:         &prevVersion,
			expectedModelStatus: db.ModelState_ModelAvailable,
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

			// Get model and version for testing
			model, modelVersion, _, err := ms.getModelServer(test.modelName, test.version, test.serverName)
			g.Expect(err).To(BeNil())

			var prevModelVersion *db.ModelVersion
			if test.prevVersion != nil {
				_, prevModelVersion, _, err = ms.getModelServer(test.modelName, *test.prevVersion, test.serverName)
				g.Expect(err).To(BeNil())
			}

			// Update model status
			isLatest := model.Latest().GetVersion() == modelVersion.GetVersion()
			err = ms.updateModelStatus(isLatest, model.Deleted, modelVersion, prevModelVersion, model)
			g.Expect(err).To(BeNil())
			g.Expect(modelVersion.State.State).To(Equal(test.expectedModelStatus))
		})
	}
}
