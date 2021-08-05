package tensorflow

import (
	"context"
	"net/url"

	"github.com/go-logr/logr"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/grpc"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/predictor"
	"github.com/seldonio/seldon-core/executor/proto/tensorflow/serving"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type GrpcTensorflowServer struct {
	Client    client.SeldonApiClient
	predictor *v1.PredictorSpec
	Log       logr.Logger
	ServerUrl *url.URL
	Namespace string
}

func NewGrpcTensorflowServer(predictor *v1.PredictorSpec, client client.SeldonApiClient, serverUrl *url.URL, namespace string) *GrpcTensorflowServer {
	return &GrpcTensorflowServer{
		Client:    client,
		predictor: predictor,
		Log:       logf.Log.WithName("SeldonGrpcApi"),
		ServerUrl: serverUrl,
		Namespace: namespace,
	}
}

func (g *GrpcTensorflowServer) execute(ctx context.Context, req proto.Message, method string, modelName string) (payload.SeldonPayload, error) {
	md := grpc.CollectMetadata(ctx)
	ctx = context.WithValue(ctx, payload.SeldonPUIDHeaderIdentifier(payload.SeldonPUIDHeader), md.Get(payload.SeldonPUIDHeader)[0])
	seldonPredictorProcess := predictor.NewPredictorProcess(ctx, g.Client, logf.Log.WithName(method), g.ServerUrl, g.Namespace, md, modelName)
	reqPayload := payload.ProtoPayload{Msg: req}
	return seldonPredictorProcess.Predict(&g.predictor.Graph, &reqPayload)
}

func (g *GrpcTensorflowServer) Classify(ctx context.Context, req *serving.ClassificationRequest) (*serving.ClassificationResponse, error) {
	resPayload, err := g.execute(ctx, req, "GrpcClassify", req.GetModelSpec().GetName())
	if err != nil {
		return nil, err
	}
	return resPayload.GetPayload().(*serving.ClassificationResponse), nil
}

func (g *GrpcTensorflowServer) Regress(ctx context.Context, req *serving.RegressionRequest) (*serving.RegressionResponse, error) {
	resPayload, err := g.execute(ctx, req, "GrpcRegress", req.GetModelSpec().GetName())
	if err != nil {
		return nil, err
	}
	return resPayload.GetPayload().(*serving.RegressionResponse), nil
}

func (g *GrpcTensorflowServer) Predict(ctx context.Context, req *serving.PredictRequest) (*serving.PredictResponse, error) {
	resPayload, err := g.execute(ctx, req, "GrpcPredict", req.GetModelSpec().GetName())
	if err != nil {
		return nil, err
	}
	return resPayload.GetPayload().(*serving.PredictResponse), nil
}

// MultiInference API for multi-headed models.
func (g *GrpcTensorflowServer) MultiInference(ctx context.Context, req *serving.MultiInferenceRequest) (*serving.MultiInferenceResponse, error) {
	resPayload, err := g.execute(ctx, req, "GrpcMultiInference", "")
	if err != nil {
		return nil, err
	}
	return resPayload.GetPayload().(*serving.MultiInferenceResponse), nil
}

// GetModelMetadata - provides access to metadata for loaded models.
func (g *GrpcTensorflowServer) GetModelMetadata(ctx context.Context, req *serving.GetModelMetadataRequest) (*serving.GetModelMetadataResponse, error) {
	seldonPredictorProcess := predictor.NewPredictorProcess(ctx, g.Client, logf.Log.WithName("GrpcGetModelMetadata"), g.ServerUrl, g.Namespace, grpc.CollectMetadata(ctx), "")
	reqPayload := payload.ProtoPayload{Msg: req}
	resPayload, err := seldonPredictorProcess.Metadata(&g.predictor.Graph, req.ModelSpec.Name, &reqPayload)
	if err != nil {
		return nil, err
	}
	return resPayload.GetPayload().(*serving.GetModelMetadataResponse), nil
}

func (g *GrpcTensorflowServer) GetModelStatus(ctx context.Context, req *serving.GetModelStatusRequest) (*serving.GetModelStatusResponse, error) {
	seldonPredictorProcess := predictor.NewPredictorProcess(ctx, g.Client, logf.Log.WithName("GrpcGetModelStatus"), g.ServerUrl, g.Namespace, grpc.CollectMetadata(ctx), "")
	reqPayload := payload.ProtoPayload{Msg: req}
	resPayload, err := seldonPredictorProcess.Status(&g.predictor.Graph, req.ModelSpec.Name, &reqPayload)
	if err != nil {
		return nil, err
	}
	return resPayload.GetPayload().(*serving.GetModelStatusResponse), nil
}

func (g *GrpcTensorflowServer) HandleReloadConfigRequest(context.Context, *serving.ReloadConfigRequest) (*serving.ReloadConfigResponse, error) {
	return nil, errors.Errorf("Not implemented")
}
