/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package agent

import (
	"context"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	pba "github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	pbad "github.com/seldonio/seldon-core/apis/go/v2/mlops/agent_debug"
	pbs "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/internal/testing_utils"
)

func setupService(numModels int, modelPrefix string, capacity int) *AgentDebug {
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	stateManager := setupLocalTestManager(numModels, modelPrefix, nil, capacity, 1)
	clientDebugService := NewAgentDebug(logger, GRPCDebugServicePort)
	clientDebugService.SetState(stateManager)
	return clientDebugService
}

func TestAgentDebugServiceSmoke(t *testing.T) {
	// TODO break this down in proper tests
	g := NewGomegaWithT(t)

	service := setupService(10, "dummy", 10)

	response, err := service.ReplicaStatus(context.TODO(), &pbad.ReplicaStatusRequest{})

	g.Expect(err).To(BeNil())

	// we create capacity above with 10 models
	g.Expect(response.GetAvailableMemoryBytes()).To(Equal(uint64(10)))

	mem := uint64(1)
	httpmock.ActivateNonDefault(service.stateManager.v2Client.(*testing_utils.V2RestClientForTest).HttpClient)
	err = service.stateManager.LoadModelVersion(
		&pba.ModelVersion{
			Model: &pbs.Model{
				Meta: &pbs.MetaData{
					Name: "dummy_1_1",
				},
				ModelSpec: &pbs.ModelSpec{
					Uri:              "gs://dummy",
					MemoryBytes:      &mem,
					ModelRuntimeInfo: getModelRuntimeInfo(1),
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

func TestAgentDebugEarlyStop(t *testing.T) {
	// TODO break this down in proper tests
	g := NewGomegaWithT(t)

	service := setupService(10, "dummy", 10)
	err := service.Stop()
	g.Expect(err).To(BeNil())
	ready := service.Ready()
	g.Expect(ready).To(BeFalse())
}
