package kfserving

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/grpc"
	"github.com/seldonio/seldon-core/executor/api/grpc/kfserving/inference"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/predictor"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	protoGrpc "google.golang.org/grpc"
	protoGrpcMetadata "google.golang.org/grpc/metadata"
	"net/url"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type GrpcKFServingServer struct {
	Client    client.SeldonApiClient
	predictor *v1.PredictorSpec
	Log       logr.Logger
	ServerUrl *url.URL
	Namespace string
}

func NewGrpcKFServingServer(predictor *v1.PredictorSpec, client client.SeldonApiClient, serverUrl *url.URL, namespace string) *GrpcKFServingServer {
	return &GrpcKFServingServer{
		Client:    client,
		predictor: predictor,
		Log:       logf.Log.WithName("KFServingGrpcApi"),
		ServerUrl: serverUrl,
		Namespace: namespace,
	}
}

func (g GrpcKFServingServer) ServerLive(ctx context.Context, request *inference.ServerLiveRequest) (*inference.ServerLiveResponse, error) {
	panic("Not Implemented")
}

func (g GrpcKFServingServer) ServerReady(ctx context.Context, request *inference.ServerReadyRequest) (*inference.ServerReadyResponse, error) {
	panic("Not Implemented")
}

func (g GrpcKFServingServer) ModelReady(ctx context.Context, request *inference.ModelReadyRequest) (*inference.ModelReadyResponse, error) {
	md := grpc.CollectMetadata(ctx)
	header := protoGrpcMetadata.Pairs(payload.SeldonPUIDHeader, md.Get(payload.SeldonPUIDHeader)[0])
	protoGrpc.SetHeader(ctx, header)
	ctx = context.WithValue(ctx, payload.SeldonPUIDHeader, md.Get(payload.SeldonPUIDHeader)[0])
	seldonPredictorProcess := predictor.NewPredictorProcess(ctx, g.Client, logf.Log.WithName("infer"), g.ServerUrl, g.Namespace, md, request.GetName())
	reqPayload := payload.ProtoPayload{Msg: request}
	resPayload, err := seldonPredictorProcess.Status(&g.predictor.Graph, request.Name, &reqPayload)
	if err != nil {
		return nil, err
	}
	return resPayload.GetPayload().(*inference.ModelReadyResponse), nil
}

func (g GrpcKFServingServer) ServerMetadata(ctx context.Context, request *inference.ServerMetadataRequest) (*inference.ServerMetadataResponse, error) {
	panic("Not Implemented")
}

func (g GrpcKFServingServer) ModelMetadata(ctx context.Context, request *inference.ModelMetadataRequest) (*inference.ModelMetadataResponse, error) {
	md := grpc.CollectMetadata(ctx)
	header := protoGrpcMetadata.Pairs(payload.SeldonPUIDHeader, md.Get(payload.SeldonPUIDHeader)[0])
	protoGrpc.SetHeader(ctx, header)
	ctx = context.WithValue(ctx, payload.SeldonPUIDHeader, md.Get(payload.SeldonPUIDHeader)[0])
	seldonPredictorProcess := predictor.NewPredictorProcess(ctx, g.Client, logf.Log.WithName("infer"), g.ServerUrl, g.Namespace, md, request.GetName())
	reqPayload := payload.ProtoPayload{Msg: request}
	resPayload, err := seldonPredictorProcess.Metadata(&g.predictor.Graph, request.Name, &reqPayload)
	if err != nil {
		return nil, err
	}
	return resPayload.GetPayload().(*inference.ModelMetadataResponse), nil
}

func (g GrpcKFServingServer) ModelInfer(ctx context.Context, request *inference.ModelInferRequest) (*inference.ModelInferResponse, error) {
	md := grpc.CollectMetadata(ctx)
	header := protoGrpcMetadata.Pairs(payload.SeldonPUIDHeader, md.Get(payload.SeldonPUIDHeader)[0])
	protoGrpc.SetHeader(ctx, header)
	ctx = context.WithValue(ctx, payload.SeldonPUIDHeader, md.Get(payload.SeldonPUIDHeader)[0])
	seldonPredictorProcess := predictor.NewPredictorProcess(ctx, g.Client, logf.Log.WithName("infer"), g.ServerUrl, g.Namespace, md, request.GetModelName())
	reqPayload := payload.ProtoPayload{Msg: request}
	resPayload, err := seldonPredictorProcess.Predict(&g.predictor.Graph, &reqPayload)
	if err != nil {
		return nil, err
	}
	return resPayload.GetPayload().(*inference.ModelInferResponse), nil
}

func (g GrpcKFServingServer) ModelStreamInfer(server inference.GRPCInferenceService_ModelStreamInferServer) error {
	panic("Not Implemented")
}

func (g GrpcKFServingServer) ModelConfig(ctx context.Context, request *inference.ModelConfigRequest) (*inference.ModelConfigResponse, error) {
	panic("Not Implemented")
}

func (g GrpcKFServingServer) ModelStatistics(ctx context.Context, request *inference.ModelStatisticsRequest) (*inference.ModelStatisticsResponse, error) {
	panic("Not Implemented")
}

func (g GrpcKFServingServer) RepositoryIndex(ctx context.Context, request *inference.RepositoryIndexRequest) (*inference.RepositoryIndexResponse, error) {
	panic("Not Implemented")
}

func (g GrpcKFServingServer) RepositoryModelLoad(ctx context.Context, request *inference.RepositoryModelLoadRequest) (*inference.RepositoryModelLoadResponse, error) {
	panic("Not Implemented")
}

func (g GrpcKFServingServer) RepositoryModelUnload(ctx context.Context, request *inference.RepositoryModelUnloadRequest) (*inference.RepositoryModelUnloadResponse, error) {
	panic("Not Implemented")
}

func (g GrpcKFServingServer) SystemSharedMemoryStatus(ctx context.Context, request *inference.SystemSharedMemoryStatusRequest) (*inference.SystemSharedMemoryStatusResponse, error) {
	panic("Not Implemented")
}

func (g GrpcKFServingServer) SystemSharedMemoryRegister(ctx context.Context, request *inference.SystemSharedMemoryRegisterRequest) (*inference.SystemSharedMemoryRegisterResponse, error) {
	panic("Not Implemented")
}

func (g GrpcKFServingServer) SystemSharedMemoryUnregister(ctx context.Context, request *inference.SystemSharedMemoryUnregisterRequest) (*inference.SystemSharedMemoryUnregisterResponse, error) {
	panic("Not Implemented")
}

func (g GrpcKFServingServer) CudaSharedMemoryStatus(ctx context.Context, request *inference.CudaSharedMemoryStatusRequest) (*inference.CudaSharedMemoryStatusResponse, error) {
	panic("Not Implemented")
}

func (g GrpcKFServingServer) CudaSharedMemoryRegister(ctx context.Context, request *inference.CudaSharedMemoryRegisterRequest) (*inference.CudaSharedMemoryRegisterResponse, error) {
	panic("Not Implemented")
}

func (g GrpcKFServingServer) CudaSharedMemoryUnregister(ctx context.Context, request *inference.CudaSharedMemoryUnregisterRequest) (*inference.CudaSharedMemoryUnregisterResponse, error) {
	panic("Not Implemented")
}
