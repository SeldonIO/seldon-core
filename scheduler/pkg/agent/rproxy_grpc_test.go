package agent

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	v2 "github.com/seldonio/seldon-core/scheduler/apis/mlops/v2_dataplane"
)

const (
	backEndGRPCServerPort = 8087
)

type mockGRPCMLServer struct {
	v2.UnimplementedGRPCInferenceServiceServer
	modelPrefix string
}

func (mlserver *mockGRPCMLServer) Start(port uint) error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)
	v2.RegisterGRPCInferenceServiceServer(grpcServer, mlserver)

	fmt.Printf("starting gRPC listening server on port %d", port)
	return grpcServer.Serve(l)
}

func (mlserver *mockGRPCMLServer) ModelInfer(ctx context.Context, r *v2.ModelInferRequest) (*v2.ModelInferResponse, error) {
	return &v2.ModelInferResponse{ModelName: r.ModelName}, nil
}

func (mlserver *mockGRPCMLServer) ModelMetadata(ctx context.Context, r *v2.ModelMetadataRequest) (*v2.ModelMetadataResponse, error) {
	return &v2.ModelMetadataResponse{Name: r.Name}, nil
}

func (mlserver *mockGRPCMLServer) ModelReady(ctx context.Context, r *v2.ModelReadyRequest) (*v2.ModelReadyResponse, error) {
	return &v2.ModelReadyResponse{Ready: true}, nil
}

func setupMockGRPCMLServer(modelPrefix string) {
	mockMLServer := mockGRPCMLServer{modelPrefix: modelPrefix}
	if err := mockMLServer.Start(backEndGRPCServerPort); err != nil {
		log.Fatal(err)
	}
}

func setupReverseGRPCService(numModels int, modelPrefix string) *reverseGRPCProxy {
	logger := log.New()
	log.SetLevel(log.DebugLevel)

	v2Client := NewV2Client("localhost", backEndServerPort, logger)
	localCacheManager := setupLocalTestManager(numModels, modelPrefix, v2Client, numModels-2)
	rp := NewReverseGRPCProxy(logger, "localhost", backEndGRPCServerPort, ReverseGRPCProxyPort)
	rp.SetState(localCacheManager)
	return rp
}

func TestReverseGRPCServiceSmoke(t *testing.T) {
	//TODO break this down in proper tests
	g := NewGomegaWithT(t)

	dummyModelNamePrefix := "dummy_model"

	go setupMockMLServer()
	go setupMockGRPCMLServer(dummyModelNamePrefix)

	rpGRPC := setupReverseGRPCService(10, dummyModelNamePrefix)
	_ = rpGRPC.Start()

	t.Log("Testing model found")

	// load model
	loaded := rpGRPC.stateManager.modelVersions.addModelVersion(
		getDummyModelDetails(dummyModelNamePrefix+"_0", uint64(1), uint32(1)))

	g.Expect(loaded).To(Equal(true))

	time.Sleep(10 * time.Millisecond)

	// client to proxy
	conn, err := grpc.Dial(":"+strconv.Itoa(ReverseGRPCProxyPort), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Cannot connect to server (%s)", err)
	}
	defer conn.Close()

	doInfer := func(modelSuffix string) (*v2.ModelInferResponse, error) {
		client := v2.NewGRPCInferenceServiceClient(conn)
		ctx := context.Background()
		ctx = metadata.AppendToOutgoingContext(ctx, resources.SeldonInternalModel, dummyModelNamePrefix+modelSuffix)
		return client.ModelInfer(ctx, &v2.ModelInferRequest{ModelName: dummyModelNamePrefix}) // note without suffix
	}

	doMeta := func(modelSuffix string) (*v2.ModelMetadataResponse, error) {
		client := v2.NewGRPCInferenceServiceClient(conn)
		ctx := context.Background()
		ctx = metadata.AppendToOutgoingContext(ctx, resources.SeldonInternalModel, dummyModelNamePrefix+modelSuffix)
		return client.ModelMetadata(ctx, &v2.ModelMetadataRequest{Name: dummyModelNamePrefix}) // note without suffix
	}

	doModelReady := func(modelSuffix string) (*v2.ModelReadyResponse, error) {
		client := v2.NewGRPCInferenceServiceClient(conn)
		ctx := context.Background()
		ctx = metadata.AppendToOutgoingContext(ctx, resources.SeldonInternalModel, dummyModelNamePrefix+modelSuffix)
		return client.ModelReady(ctx, &v2.ModelReadyRequest{Name: dummyModelNamePrefix}) // note without suffix
	}

	responseInfer, errInfer := doInfer("_0")
	g.Expect(responseInfer.ModelName).To(Equal(dummyModelNamePrefix + "_0"))
	g.Expect(errInfer).To(BeNil())

	responseMeta, errMeta := doMeta("_0")
	g.Expect(responseMeta.Name).To(Equal(dummyModelNamePrefix + "_0"))
	g.Expect(errMeta).To(BeNil())

	responseReady, errReady := doModelReady("_0")
	g.Expect(responseReady.Ready).To(Equal(true))
	g.Expect(errReady).To(BeNil())

	t.Log("Testing model not found")
	_, errInfer = doInfer("_1")
	g.Expect(errInfer).NotTo(BeNil())

	_, errMeta = doMeta("_1")
	g.Expect(errMeta).NotTo(BeNil())

	t.Log("Testing status")
	g.Expect(rpGRPC.Ready()).To(Equal(true))
	_ = rpGRPC.Stop()
	g.Expect(rpGRPC.Ready()).To(Equal(false))

	t.Logf("Done")

}
