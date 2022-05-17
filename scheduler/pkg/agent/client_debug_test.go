package agent

import (
	"context"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/gomega"
	pba "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	pbad "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent_debug"
	pbs "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	log "github.com/sirupsen/logrus"
)

func setupService(numModels int, modelPrefix string, capacity int) *ClientDebug {
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	stateManager := setupLocalTestManager(numModels, modelPrefix, nil, capacity, 1)
	clientDebugService := NewClientDebug(logger, GRPCDebugServicePort)
	clientDebugService.SetState(stateManager)
	return clientDebugService
}

func TestAgentDebugServiceSmoke(t *testing.T) {
	//TODO break this down in proper tests
	g := NewGomegaWithT(t)

	service := setupService(10, "dummy", 10)

	response, err := service.ReplicaStatus(context.TODO(), &pbad.ReplicaStatusRequest{})

	g.Expect(err).To(BeNil())

	// we create capacity above with 10 models
	g.Expect(response.GetAvailableMemoryBytes()).To(Equal(uint64(10)))

	mem := uint64(1)
	httpmock.ActivateNonDefault(service.stateManager.v2Client.httpClient)
	err = service.stateManager.LoadModelVersion(
		&pba.ModelVersion{
			Model: &pbs.Model{
				Meta: &pbs.MetaData{
					Name: "dummy_1_1",
				},
				ModelSpec: &pbs.ModelSpec{
					Uri:         "gs://dummy",
					MemoryBytes: &mem,
				},
			},
		},
	)
	g.Expect(err).To(BeNil())
	httpmock.DeactivateAndReset()

	response, err = service.ReplicaStatus(context.TODO(), &pbad.ReplicaStatusRequest{})
	g.Expect(err).To(BeNil())

	// we loaded one model
	g.Expect(response.GetAvailableMemoryBytes()).To(Equal(uint64(9)))

	// check that we get it back
	models := response.Models
	g.Expect(len(models)).To(Equal(1))
	g.Expect(models[0].Name).To(Equal("dummy_1_1"))
	g.Expect(models[0].State).To(Equal(pbad.ModelReplicaState_InMemory))
	// we check up to a second resolution because of latency
	actualTs := models[0].GetLastAccessed().AsTime().Truncate(time.Second)
	expectedTs := time.Now().Truncate(time.Second)
	if !actualTs.Equal(expectedTs) {
		t.Errorf("Timestamps do not match")
	}

	t.Logf("Done!")
}
