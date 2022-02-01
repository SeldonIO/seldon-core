package store

import (
	"testing"

	"github.com/seldonio/seldon-core/scheduler/pkg/coordinator"

	. "github.com/onsi/gomega"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	log "github.com/sirupsen/logrus"
)

func TestUpdateStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                string
		store               *LocalSchedulerStore
		modelName           string
		serverName          string
		version             uint32
		prevVersion         *uint32
		expectedModelStatus ModelState
	}
	prevVersion := uint32(1)
	tests := []test{
		{
			name: "Available",
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
									ModelSpec: &pb.ModelSpec{},
									DeploymentSpec: &pb.DeploymentSpec{
										Replicas: 1,
									},
								},
								server: "server2",
								replicas: map[int]ReplicaStatus{
									0: {State: Loaded},
								},
							},
							{
								version: 2,
								modelDefn: &pb.Model{
									Meta: &pb.MetaData{
										Name: "model",
									},
									ModelSpec: &pb.ModelSpec{},
									DeploymentSpec: &pb.DeploymentSpec{
										Replicas: 1,
									},
								},
								server: "server1",
								replicas: map[int]ReplicaStatus{
									0: {State: Available},
								},
							},
						},
					},
				},
				servers: map[string]*Server{
					"server1": {
						name: "server1",
						replicas: map[int]*ServerReplica{
							0: {},
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
			modelName:           "model",
			serverName:          "server2",
			version:             2,
			prevVersion:         nil,
			expectedModelStatus: ModelAvailable,
		},
		{
			name: "NotEnoughReplicasButPreviousAvailable",
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
									ModelSpec: &pb.ModelSpec{},
									DeploymentSpec: &pb.DeploymentSpec{
										Replicas: 1,
									},
								},
								server: "server2",
								replicas: map[int]ReplicaStatus{
									0: {State: Available},
								},
								state: ModelStatus{State: ModelAvailable},
							},
							{
								version: 2,
								modelDefn: &pb.Model{
									Meta: &pb.MetaData{
										Name: "model",
									},
									ModelSpec: &pb.ModelSpec{},
									DeploymentSpec: &pb.DeploymentSpec{
										Replicas: 2,
									},
								},
								server: "server1",
								replicas: map[int]ReplicaStatus{
									0: {State: Available},
								},
							},
						},
					},
				},
				servers: map[string]*Server{
					"server1": {
						name: "server1",
						replicas: map[int]*ServerReplica{
							0: {},
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
			modelName:           "model",
			serverName:          "server2",
			version:             2,
			prevVersion:         &prevVersion,
			expectedModelStatus: ModelAvailable,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			eventHub := &coordinator.ModelEventHub{}
			ms := NewMemoryStore(logger, test.store, eventHub)
			model, modelVersion, _, err := ms.getModelServer(test.modelName, test.version, test.serverName)
			var prevModelVersion *ModelVersion
			if test.prevVersion != nil {
				_, prevModelVersion, _, err = ms.getModelServer(test.modelName, *test.prevVersion, test.serverName)
				g.Expect(err).To(BeNil())
			}
			g.Expect(err).To(BeNil())
			isLatest := model.Latest().GetVersion() == modelVersion.GetVersion()
			ms.updateModelStatus(isLatest, model.deleted, modelVersion, prevModelVersion)
			g.Expect(modelVersion.state.State).To(Equal(test.expectedModelStatus))
		})
	}
}
