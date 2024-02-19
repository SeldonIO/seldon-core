/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package oip

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/internal/testing_utils"
	testing_utils2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/internal/testing_utils"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

func TestCommunicationErrors(t *testing.T) {
	// this should fail because of dns error
	g := NewGomegaWithT(t)
	modelName := "dummy"
	r := testing_utils.CreateTestV2Client([]string{modelName}, 200)
	err := r.LoadModel(modelName)
	g.Expect(err.ErrCode).To(Equal(interfaces.V2CommunicationErrCode))
}

func TestGRPCV2(t *testing.T) {
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

	v2Client := NewV2Client(GetV2ConfigWithDefaults("", backEndGRPCPort), log.New())

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

func TestGRPCV2Timeout(t *testing.T) {
	g := NewGomegaWithT(t)

	unloadSleep := 5 * time.Second
	loadSleep := 2 * time.Second
	controlPlaneSleep := 1 * time.Second
	mockMLServer := &testing_utils.MockGRPCMLServer{
		UnloadSleep: unloadSleep, LoadSleep: loadSleep, ControlPlaneSleep: controlPlaneSleep}
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

	v2Config := GetV2ConfigWithDefaults("", backEndGRPCPort)
	v2Config.GRPCModelServerUnloadTimeout = unloadSleep / 2
	v2Config.GRPCModelServerLoadTimeout = loadSleep / 2
	v2Config.GRPCControlPlaneTimeout = controlPlaneSleep / 2
	v2Client := NewV2Client(v2Config, log.New())

	dummModel := "dummy"

	v2Err := v2Client.LoadModel(dummModel)
	g.Expect(v2Err).NotTo(BeNil())
	g.Expect(v2Err.ErrCode).To(Equal(int(codes.DeadlineExceeded)))

	v2Err = v2Client.UnloadModel(dummModel)
	g.Expect(v2Err).NotTo(BeNil())
	g.Expect(v2Err.ErrCode).To(Equal(int(codes.DeadlineExceeded)))

	err = v2Client.Live()
	g.Expect(err).NotTo(BeNil())
	e, _ := status.FromError(err)
	g.Expect(e.Code()).To(Equal(codes.DeadlineExceeded))

	_, err = v2Client.getModelsGrpc()
	g.Expect(err).NotTo(BeNil())
	e, _ = status.FromError(err)
	g.Expect(e.Code()).To(Equal(codes.DeadlineExceeded))
}

func TestDefaultV2Config(t *testing.T) {
	g := NewGomegaWithT(t)

	v2Config := GetV2ConfigWithDefaults("", 0)
	g.Expect(v2Config.GRPCModelServerLoadTimeout).To(Equal(util.GRPCModelServerLoadTimeout))
	g.Expect(v2Config.GRPCModelServerUnloadTimeout).To(Equal(util.GRPCModelServerUnloadTimeout))
	g.Expect(v2Config.GRPCMaxMsgSizeBytes).To(Equal(util.GRPCMaxMsgSizeBytes))
	g.Expect(v2Config.GRPCControlPlaneTimeout).To(Equal(util.GRPCControlPlaneTimeout))
	g.Expect(v2Config.GRPCRetryBackoff).To(Equal(util.GRPCRetryBackoff))
	g.Expect(v2Config.GRPRetryMaxCount).To(Equal(uint(util.GRPCRetryMaxCount)))
	g.Expect(v2Config.Host).To(Equal(""))
	g.Expect(v2Config.GRPCPort).To(Equal(0))
}

func TestGrpcV2WithError(t *testing.T) {
	g := NewGomegaWithT(t)

	// note no grpc server to respond

	backEndGRPCPort, err := testing_utils2.GetFreePortForTest()
	if err != nil {
		t.Fatal(err)
	}
	v2Client := NewV2Client(GetV2ConfigWithDefaults("", backEndGRPCPort), log.New())

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
	v2Client := NewV2Client(GetV2ConfigWithDefaults("", backEndGRPCPort), log.New())
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
