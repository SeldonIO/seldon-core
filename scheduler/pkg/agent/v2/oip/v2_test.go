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

package oip

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/internal/testing_utils"
	testing_utils2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/internal/testing_utils"
	log "github.com/sirupsen/logrus"
)

func TestCommunicationErrors(t *testing.T) {
	// this should fail because of dns error
	g := NewGomegaWithT(t)
	modelName := "dummy"
	r := testing_utils.CreateTestV2Client([]string{modelName}, 200)
	err := r.LoadModel(modelName)
	g.Expect(err.ErrCode).To(Equal(interfaces.V2CommunicationErrCode))
}

func TestGrpcV2(t *testing.T) {
	g := NewGomegaWithT(t)

	mockMLServer := &testing_utils.MockGRPCMLServer{}
	backEndGRPCPort, err := testing_utils2.GetFreePortForTest()
	if err != nil {
		t.Fatal(err)
	}
	_ = mockMLServer.Setup(uint(backEndGRPCPort))
	go func() {
		_ = mockMLServer.Start()
	}()
	defer mockMLServer.Stop()

	time.Sleep(10 * time.Millisecond)

	v2Client := NewV2Client("", backEndGRPCPort, log.New())

	dummModel := "dummy"

	v2Err := v2Client.LoadModel(dummModel)
	g.Expect(v2Err).To(BeNil())

	v2Err = v2Client.UnloadModel(dummModel)
	g.Expect(v2Err).To(BeNil())

	v2Err = v2Client.UnloadModel(testing_utils.ModelNameMissing)
	g.Expect(v2Err.IsNotFound()).To(BeTrue())

	mockMLServer.SetModels([]interfaces.ServerModelInfo{
		{Name: dummModel, State: interfaces.ServerModelState_READY}, 
		{Name: "", State: interfaces.ServerModelState_UNAVAILABLE}})
	models, err := v2Client.GetModels()
	g.Expect(err).To(BeNil())
	g.Expect(models).To(Equal([]interfaces.ServerModelInfo{
		{Name: dummModel, State: interfaces.ServerModelState_READY}})) // empty string models should be discarded

	err = v2Client.Live()
	g.Expect(err).To(BeNil())

}

func TestGrpcV2WithError(t *testing.T) {
	g := NewGomegaWithT(t)

	// note no grpc server to respond

	backEndGRPCPort, err := testing_utils2.GetFreePortForTest()
	if err != nil {
		t.Fatal(err)
	}
	v2Client := NewV2Client("", backEndGRPCPort, log.New())

	dummModel := "dummy"

	v2Err := v2Client.LoadModel(dummModel)
	g.Expect(v2Err).NotTo(BeNil())

	v2Err = v2Client.UnloadModel(dummModel)
	g.Expect(v2Err).NotTo(BeNil())

	err = v2Client.Live()
	g.Expect(err).NotTo(BeNil())

}

func TestGrpcV2WithRetry(t *testing.T) {
	// note: we delay starting the server to simulate transient errors
	g := NewGomegaWithT(t)
	mockMLServer := &testing_utils.MockGRPCMLServer{}
	backEndGRPCPort, err := testing_utils2.GetFreePortForTest()
	if err != nil {
		t.Fatal(err)
	}
	_ = mockMLServer.Setup(uint(backEndGRPCPort))

	//initial conn setup
	go func() {
		_ = mockMLServer.Start()
	}()
	v2Client := NewV2Client("", backEndGRPCPort, log.New())
	err = v2Client.Live()
	g.Expect(err).To(BeNil())
	mockMLServer.Stop()

	// start the server in background after 0.5s
	go func() {
		time.Sleep(500 * time.Millisecond)
		_ = mockMLServer.Setup(uint(backEndGRPCPort))
		go func() {
			_ = mockMLServer.Start()
		}()

	}()
	defer mockMLServer.Stop()

	// make sure that we can still get to the server, this will require retries as the server starts after 0.5s
	for i := 0; i < 20; i++ {
		err = v2Client.Live()
		g.Expect(err).To(BeNil())
	}
}
