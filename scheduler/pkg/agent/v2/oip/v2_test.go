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
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/internal/testing_utils"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
	log "github.com/sirupsen/logrus"
)

func createTestV2ClientMockResponders(host string, port int, modelName string, status int, state *testing_utils.V2State) {

	httpmock.RegisterResponder("POST", fmt.Sprintf("http://%s:%d/v2/repository/models/%s/load", host, port, modelName),
		state.LoadResponder(modelName, status))
	httpmock.RegisterResponder("POST", fmt.Sprintf("http://%s:%d/v2/repository/models/%s/unload", host, port, modelName),
		state.UnloadResponder(modelName, status))
}

func createTestV2ClientwithState(models []string, status int) (*V2Client, *testing_utils.V2State) {
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	host := "model-server"
	port := 8080
	v2 := NewV2Client(host, port, logger, false)
	state := &testing_utils.V2State{
		Models: make(map[string]bool, len(models)),
	}

	for _, model := range models {
		createTestV2ClientMockResponders(host, port, model, status, state)
	}
	// we do not care about ready in tests
	httpmock.RegisterResponder("GET", fmt.Sprintf("http://%s:%d/v2/health/live", host, port),
		httpmock.NewStringResponder(200, `{}`))
	return v2, state
}

func createTestV2Client(models []string, status int) *V2Client {
	v2, _ := createTestV2ClientwithState(models, status)
	return v2
}

func TestCommunicationErrors(t *testing.T) {
	// this should fail because of dns error
	g := NewGomegaWithT(t)
	modelName := "dummy"
	r := createTestV2Client([]string{modelName}, 200)
	err := r.LoadModel(modelName)
	g.Expect(err.ErrCode).To(Equal(interfaces.V2CommunicationErrCode))
}

func TestRequestErrors(t *testing.T) {
	// in this test we are not enabling httpmock
	// and therefore all http requests should fail because of dns / client error
	g := NewGomegaWithT(t)
	modelName := "dummy"
	v2 := NewV2Client("httpwrong://server", 0, log.New(), false)
	err := v2.LoadModel(modelName)
	g.Expect(err.ErrCode).To(Equal(interfaces.V2RequestErrCode))
}

func TestLoad(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		models []string
		status int
		err    error
	}
	tests := []test{
		{models: []string{"iris"}, status: 200, err: nil},
		{models: []string{"iris"}, status: 400, err: interfaces.ErrV2BadRequest},
	}
	for _, test := range tests {
		r := createTestV2Client(test.models, test.status)
		httpmock.ActivateNonDefault(r.HttpClient)
		for _, model := range test.models {
			err := r.LoadModel(model)
			if test.status == 200 {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).ToNot(BeNil())
				if test.err != nil {
					g.Expect(errors.Is(err.Err, test.err)).To(BeTrue())
				}
			}
		}
		g.Expect(httpmock.GetTotalCallCount()).To(Equal(len(test.models)))
		httpmock.DeactivateAndReset()
	}
}

func TestUnload(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		models []string
		status int
		err    error
	}
	tests := []test{
		{models: []string{"iris"}, status: 200, err: nil},
		{models: []string{"iris"}, status: 400, err: interfaces.ErrV2BadRequest},
	}
	for _, test := range tests {
		r := createTestV2Client(test.models, test.status)
		httpmock.ActivateNonDefault(r.HttpClient)
		for _, model := range test.models {
			err := r.UnloadModel(model)
			if test.status == 200 {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).ToNot(BeNil())
				if test.err != nil {
					g.Expect(errors.Is(err.Err, test.err)).To(BeTrue())
				}
			}
		}
		g.Expect(httpmock.GetTotalCallCount()).To(Equal(len(test.models)))
		httpmock.DeactivateAndReset()
	}
}

func TestGrpcV2(t *testing.T) {
	g := NewGomegaWithT(t)

	mockMLServer := &testing_utils.MockGRPCMLServer{}
	backEndGRPCPort, err := util.GetFreePortForTest()
	if err != nil {
		t.Fatal(err)
	}
	_ = mockMLServer.Setup(uint(backEndGRPCPort))
	go func() {
		_ = mockMLServer.Start()
	}()
	defer mockMLServer.Stop()

	time.Sleep(10 * time.Millisecond)

	v2Client := NewV2Client("", backEndGRPCPort, log.New(), true)

	dummModel := "dummy"

	v2Err := v2Client.LoadModel(dummModel)
	g.Expect(v2Err).To(BeNil())

	v2Err = v2Client.UnloadModel(dummModel)
	g.Expect(v2Err).To(BeNil())

	v2Err = v2Client.UnloadModel(testing_utils.ModelNameMissing)
	g.Expect(v2Err.IsNotFound()).To(BeTrue())

	mockMLServer.SetModels([]interfaces.ServerModelInfo{{dummModel, interfaces.ServerModelState_READY}, {"", interfaces.ServerModelState_UNAVAILABLE}})
	models, err := v2Client.GetModels()
	g.Expect(err).To(BeNil())
	g.Expect(models).To(Equal([]interfaces.ServerModelInfo{{dummModel, interfaces.ServerModelState_READY}})) // empty string models should be discarded

	err = v2Client.Live()
	g.Expect(err).To(BeNil())

}

func TestGrpcV2WithError(t *testing.T) {
	g := NewGomegaWithT(t)

	// note no grpc server to respond

	backEndGRPCPort, err := util.GetFreePortForTest()
	if err != nil {
		t.Fatal(err)
	}
	v2Client := NewV2Client("", backEndGRPCPort, log.New(), true)

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
	backEndGRPCPort, err := util.GetFreePortForTest()
	if err != nil {
		t.Fatal(err)
	}
	_ = mockMLServer.Setup(uint(backEndGRPCPort))

	//initial conn setup
	go func() {
		_ = mockMLServer.Start()
	}()
	v2Client := NewV2Client("", backEndGRPCPort, log.New(), true)
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
