package api

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/grpc/proto"
	"github.com/seldonio/seldon-core/executor/api/machinelearning/v1alpha2"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/predictor"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type GrpcSeldonServer struct {
	Client    client.SeldonApiClient
	predictor *v1alpha2.PredictorSpec
	Log       logr.Logger
}

func NewGrpcSeldonServer(logger logr.Logger, predictor *v1alpha2.PredictorSpec, client client.SeldonApiClient) *GrpcSeldonServer {
	return &GrpcSeldonServer{
		Client:    client,
		predictor: predictor,
		Log:       logger,
	}
}

func (g GrpcSeldonServer) Predict(ctx context.Context, req *proto.SeldonMessage) (*proto.SeldonMessage, error) {
	seldonPredictorProcess := &predictor.PredictorProcess{
		Client: NewSeldonGrpcClient(),
		Log:    logf.Log.WithName("SeldonMessageRestClient"),
	}

	reqPayload := payload.SeldonMessagePayload{Msg: req}
	resPayload, err := seldonPredictorProcess.Execute(g.predictor.Graph, &reqPayload)
	if err != nil {
		return nil, err
	}
	return resPayload.GetPayload().(*proto.SeldonMessage), nil
}

func (g GrpcSeldonServer) SendFeedback(context.Context, *proto.Feedback) (*proto.SeldonMessage, error) {
	panic("implement me")
}
