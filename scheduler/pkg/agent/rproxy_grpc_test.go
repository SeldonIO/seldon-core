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
	"strconv"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	v2 "github.com/seldonio/seldon-core/apis/go/v2/mlops/v2_dataplane"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/internal/testing_utils"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/modelscaling"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/resources"
	testing_utils2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/internal/testing_utils"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

func setupReverseGRPCService(numModels int, modelPrefix string, backEndGRPCPort, rpPort, backEndServerPort int) *reverseGRPCProxy {
	logger := log.New()
	log.SetLevel(log.DebugLevel)

	v2Client := testing_utils.NewV2RestClientForTest("localhost", backEndServerPort, logger)
	localCacheManager := setupLocalTestManager(numModels, modelPrefix, v2Client, numModels-2, 1)
	modelScalingStatsCollector := modelscaling.NewDataPlaneStatsCollector(
		[]interfaces.ModelStatsKeeper{modelscaling.NewModelReplicaLagsKeeper(), modelscaling.NewModelReplicaLastUsedKeeper()},
		logger,
	)
	rp := NewReverseGRPCProxy(
		newFakeMetricsHandler(),
		logger,
		"localhost",
		uint(backEndGRPCPort),
		uint(rpPort),
		modelScalingStatsCollector,
	)
	rp.SetState(localCacheManager)
	return rp
}

func TestReverseGRPCServiceSmoke(t *testing.T) {
	//TODO break this down in proper tests
	g := NewGomegaWithT(t)

	dummyModelNamePrefix := "dummy_model"

	mockMLServerState := &mockMLServerState{
		models:         make(map[string]bool),
		modelsNotFound: make(map[string]bool),
		mu:             &sync.Mutex{},
	}

	serverPort, err := testing_utils2.GetFreePortForTest()
	if err != nil {
		t.Fatal(err)
	}
	mlserver := setupMockMLServer(mockMLServerState, serverPort)
	go func() {
		_ = mlserver.ListenAndServe()
	}()
	defer func() {
		_ = mlserver.Shutdown(context.Background())
	}()

	mockMLServer := &testing_utils.MockGRPCMLServer{}

	backEndGRPCPort, err := testing_utils2.GetFreePortForTest()
	if err != nil {
		t.Fatal(err)
	}

	err = mockMLServer.Setup(uint(backEndGRPCPort))
	g.Expect(err).To(BeNil())
	go func() {
		err := mockMLServer.Start()
		g.Expect(err).To(BeNil())
	}()
	defer mockMLServer.Stop()

	rpPort, err := testing_utils2.GetFreePortForTest()
	if err != nil {
		t.Fatal(err)
	}
	rpGRPC := setupReverseGRPCService(10, dummyModelNamePrefix, backEndGRPCPort, rpPort, serverPort)
	_ = rpGRPC.Start()

	t.Log("Testing model found")
	time.Sleep(50 * time.Millisecond)

	// load model
	err = rpGRPC.stateManager.LoadModelVersion(
		getDummyModelDetails(dummyModelNamePrefix+"_0", uint64(1), uint32(1)))
	g.Expect(err).To(BeNil())

	time.Sleep(50 * time.Millisecond)

	// client to proxy
	conn, err := grpc.Dial(":"+strconv.Itoa(rpPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Cannot connect to server (%s)", err)
	}
	defer conn.Close()

	doInfer := func(modelSuffix string) (*v2.ModelInferResponse, error) {
		client := v2.NewGRPCInferenceServiceClient(conn)
		ctx := context.Background()
		ctx = metadata.AppendToOutgoingContext(ctx, resources.SeldonInternalModelHeader, dummyModelNamePrefix+modelSuffix, resources.SeldonModelHeader, dummyModelNamePrefix+modelSuffix)
		return client.ModelInfer(ctx, &v2.ModelInferRequest{ModelName: dummyModelNamePrefix}) // note without suffix
	}

	doMeta := func(modelSuffix string) (*v2.ModelMetadataResponse, error) {
		client := v2.NewGRPCInferenceServiceClient(conn)
		ctx := context.Background()
		ctx = metadata.AppendToOutgoingContext(ctx, resources.SeldonInternalModelHeader, dummyModelNamePrefix+modelSuffix, resources.SeldonModelHeader, dummyModelNamePrefix+modelSuffix)
		return client.ModelMetadata(ctx, &v2.ModelMetadataRequest{Name: dummyModelNamePrefix}) // note without suffix
	}

	doModelReady := func(modelSuffix string) (*v2.ModelReadyResponse, error) {
		client := v2.NewGRPCInferenceServiceClient(conn)
		ctx := context.Background()
		ctx = metadata.AppendToOutgoingContext(ctx, resources.SeldonInternalModelHeader, dummyModelNamePrefix+modelSuffix, resources.SeldonModelHeader, dummyModelNamePrefix+modelSuffix)
		return client.ModelReady(ctx, &v2.ModelReadyRequest{Name: dummyModelNamePrefix}) // note without suffix
	}

	responseInfer, errInfer := doInfer("_0")
	g.Expect(errInfer).To(BeNil())
	g.Expect(responseInfer.ModelName).To(Equal(dummyModelNamePrefix + "_0"))
	g.Expect(responseInfer.ModelVersion).To(Equal("")) // in practice this should be something else

	t.Log("Testing model scaling stats")
	g.Expect(rpGRPC.modelScalingStatsCollector.StatKeepers[0].Get(dummyModelNamePrefix + "_0")).To(Equal(uint32(0)))
	g.Expect(rpGRPC.modelScalingStatsCollector.StatKeepers[1].Get(dummyModelNamePrefix + "_0")).Should(BeNumerically("<=", time.Now().Unix())) // only triggered when we get results back

	responseMeta, errMeta := doMeta("_0")
	g.Expect(responseMeta.Name).To(Equal(dummyModelNamePrefix + "_0"))
	g.Expect(responseMeta.Versions).To(Equal([]string{""})) // in practice this should be something else
	g.Expect(errMeta).To(BeNil())

	responseReady, errReady := doModelReady("_0")
	g.Expect(responseReady.Ready).To(Equal(true))
	g.Expect(mockMLServerState.isModelLoaded(dummyModelNamePrefix + "_0")).To(Equal(true))
	g.Expect(errReady).To(BeNil())

	t.Log("Testing lazy load")
	mockMLServerState.setModelServerUnloaded(dummyModelNamePrefix + "_0")
	responseInfer, errInfer = doInfer("_0")
	g.Expect(errInfer).To(BeNil())
	g.Expect(responseInfer.ModelName).To(Equal(dummyModelNamePrefix + "_0"))
	g.Expect(responseInfer.ModelVersion).To(Equal("")) // in practice this should be something else

	t.Log("Testing model not found")
	_, errInfer = doInfer("_1")
	g.Expect(errInfer).NotTo(BeNil())
	g.Expect(mockMLServerState.isModelLoaded(dummyModelNamePrefix + "_1")).To(Equal(false))

	_, errMeta = doMeta("_1")
	g.Expect(errMeta).NotTo(BeNil())

	t.Log("Testing status")
	g.Expect(rpGRPC.Ready()).To(Equal(true))
	_ = rpGRPC.Stop()
	g.Expect(rpGRPC.Ready()).To(Equal(false))

	t.Logf("Done")
}

func TestReverseGRPCServiceEarlyStop(t *testing.T) {
	//TODO break this down in proper tests
	g := NewGomegaWithT(t)

	dummyModelNamePrefix := "dummy_model"

	rpGRPC := setupReverseGRPCService(0, dummyModelNamePrefix, 1, 1, 1)
	err := rpGRPC.Stop()
	g.Expect(err).To(BeNil())
	ready := rpGRPC.Ready()
	g.Expect(ready).To(BeFalse())
}

func TestCreateOutgoingCtxWithRequestId(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name               string
		ctx                context.Context
		expectNewRequestId bool
		expectedRequestId  string
	}
	tests := []test{
		{
			name:               "No request id in incoming context",
			ctx:                context.TODO(),
			expectNewRequestId: true,
		},
		{
			name:               "request id already in incoming context",
			ctx:                metadata.NewIncomingContext(context.TODO(), metadata.New(map[string]string{util.RequestIdHeader: "1234"})),
			expectNewRequestId: false,
			expectedRequestId:  "1234",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rp := reverseGRPCProxy{
				logger: log.New(),
			}
			ctx, requestId := rp.createOutgoingCtxWithRequestId(test.ctx)
			if test.expectNewRequestId {
				md, _ := metadata.FromOutgoingContext(ctx)
				g.Expect(len(md.Get(util.RequestIdHeader))).To(Equal(1))
				g.Expect(md.Get(util.RequestIdHeader)[0]).To(Equal(requestId))
			} else {
				g.Expect(requestId).To(Equal(test.expectedRequestId))
			}
		})
	}
}
