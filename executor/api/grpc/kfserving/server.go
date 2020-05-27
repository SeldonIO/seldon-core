package kfserving

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/grpc"
	"github.com/seldonio/seldon-core/executor/api/grpc/kfserving/proto"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/predictor"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
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

func (g GrpcKFServingServer) ServerLive(ctx context.Context, request *proto.ServerLiveRequest) (*proto.ServerLiveResponse, error) {
	panic("implement me")
}

func (g GrpcKFServingServer) ServerReady(ctx context.Context, request *proto.ServerReadyRequest) (*proto.ServerReadyResponse, error) {
	panic("implement me")
}

func (g GrpcKFServingServer) ModelReady(ctx context.Context, request *proto.ModelReadyRequest) (*proto.ModelReadyResponse, error) {
	md := grpc.CollectMetadata(ctx)
	ctx = context.WithValue(ctx, payload.SeldonPUIDHeader, md.Get(payload.SeldonPUIDHeader)[0])
	seldonPredictorProcess := predictor.NewPredictorProcess(ctx, g.Client, logf.Log.WithName("infer"), g.ServerUrl, g.Namespace, md)
	reqPayload := payload.ProtoPayload{Msg: request}
	resPayload, err := seldonPredictorProcess.Status(g.predictor.Graph, request.Name, &reqPayload)
	if err != nil {
		return nil, err
	}
	return resPayload.GetPayload().(*proto.ModelReadyResponse), nil
}

func (g GrpcKFServingServer) ServerMetadata(ctx context.Context, request *proto.ServerMetadataRequest) (*proto.ServerMetadataResponse, error) {
	panic("implement me")
}

func (g GrpcKFServingServer) ModelMetadata(ctx context.Context, request *proto.ModelMetadataRequest) (*proto.ModelMetadataResponse, error) {
	md := grpc.CollectMetadata(ctx)
	ctx = context.WithValue(ctx, payload.SeldonPUIDHeader, md.Get(payload.SeldonPUIDHeader)[0])
	seldonPredictorProcess := predictor.NewPredictorProcess(ctx, g.Client, logf.Log.WithName("infer"), g.ServerUrl, g.Namespace, md)
	reqPayload := payload.ProtoPayload{Msg: request}
	resPayload, err := seldonPredictorProcess.Metadata(g.predictor.Graph, request.Name, &reqPayload)
	if err != nil {
		return nil, err
	}
	return resPayload.GetPayload().(*proto.ModelMetadataResponse), nil
}

func (g GrpcKFServingServer) ModelInfer(ctx context.Context, request *proto.ModelInferRequest) (*proto.ModelInferResponse, error) {
	md := grpc.CollectMetadata(ctx)
	ctx = context.WithValue(ctx, payload.SeldonPUIDHeader, md.Get(payload.SeldonPUIDHeader)[0])
	seldonPredictorProcess := predictor.NewPredictorProcess(ctx, g.Client, logf.Log.WithName("infer"), g.ServerUrl, g.Namespace, md)
	reqPayload := payload.ProtoPayload{Msg: request}
	resPayload, err := seldonPredictorProcess.Predict(g.predictor.Graph, &reqPayload)
	if err != nil {
		return nil, err
	}
	return resPayload.GetPayload().(*proto.ModelInferResponse), nil
}

func (g GrpcKFServingServer) ModelStreamInfer(server proto.GRPCInferenceService_ModelStreamInferServer) error {
	panic("implement me")
}

func (g GrpcKFServingServer) ModelConfig(ctx context.Context, request *proto.ModelConfigRequest) (*proto.ModelConfigResponse, error) {
	panic("implement me")
}

func (g GrpcKFServingServer) ModelStatistics(ctx context.Context, request *proto.ModelStatisticsRequest) (*proto.ModelStatisticsResponse, error) {
	panic("implement me")
}

func (g GrpcKFServingServer) RepositoryIndex(ctx context.Context, request *proto.RepositoryIndexRequest) (*proto.RepositoryIndexResponse, error) {
	panic("implement me")
}

func (g GrpcKFServingServer) RepositoryModelLoad(ctx context.Context, request *proto.RepositoryModelLoadRequest) (*proto.RepositoryModelLoadResponse, error) {
	panic("implement me")
}

func (g GrpcKFServingServer) RepositoryModelUnload(ctx context.Context, request *proto.RepositoryModelUnloadRequest) (*proto.RepositoryModelUnloadResponse, error) {
	panic("implement me")
}

func (g GrpcKFServingServer) SystemSharedMemoryStatus(ctx context.Context, request *proto.SystemSharedMemoryStatusRequest) (*proto.SystemSharedMemoryStatusResponse, error) {
	panic("implement me")
}

func (g GrpcKFServingServer) SystemSharedMemoryRegister(ctx context.Context, request *proto.SystemSharedMemoryRegisterRequest) (*proto.SystemSharedMemoryRegisterResponse, error) {
	panic("implement me")
}

func (g GrpcKFServingServer) SystemSharedMemoryUnregister(ctx context.Context, request *proto.SystemSharedMemoryUnregisterRequest) (*proto.SystemSharedMemoryUnregisterResponse, error) {
	panic("implement me")
}

func (g GrpcKFServingServer) CudaSharedMemoryStatus(ctx context.Context, request *proto.CudaSharedMemoryStatusRequest) (*proto.CudaSharedMemoryStatusResponse, error) {
	panic("implement me")
}

func (g GrpcKFServingServer) CudaSharedMemoryRegister(ctx context.Context, request *proto.CudaSharedMemoryRegisterRequest) (*proto.CudaSharedMemoryRegisterResponse, error) {
	panic("implement me")
}

func (g GrpcKFServingServer) CudaSharedMemoryUnregister(ctx context.Context, request *proto.CudaSharedMemoryUnregisterRequest) (*proto.CudaSharedMemoryUnregisterResponse, error) {
	panic("implement me")
}
