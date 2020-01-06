package seldon

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/grpc"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/predictor"
	"github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	"net/url"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type GrpcSeldonServer struct {
	Client    client.SeldonApiClient
	predictor *v1.PredictorSpec
	Log       logr.Logger
	ServerUrl *url.URL
	Namespace string
}

func NewGrpcSeldonServer(predictor *v1.PredictorSpec, client client.SeldonApiClient, serverUrl *url.URL, namespace string) *GrpcSeldonServer {
	return &GrpcSeldonServer{
		Client:    client,
		predictor: predictor,
		Log:       logf.Log.WithName("SeldonGrpcApi"),
		ServerUrl: serverUrl,
		Namespace: namespace,
	}
}

func (g GrpcSeldonServer) Predict(ctx context.Context, req *proto.SeldonMessage) (*proto.SeldonMessage, error) {
	seldonPredictorProcess := predictor.NewPredictorProcess(ctx, g.Client, logf.Log.WithName("SeldonMessageRestClient"), grpc.GetEventId(ctx), g.ServerUrl, g.Namespace)
	reqPayload := payload.ProtoPayload{Msg: req}
	resPayload, err := seldonPredictorProcess.Predict(g.predictor.Graph, &reqPayload)
	if err != nil {
		g.Log.Error(err, "Failed to call predict")
	}
	return resPayload.GetPayload().(*proto.SeldonMessage), err
}

func (g GrpcSeldonServer) SendFeedback(ctx context.Context, req *proto.Feedback) (*proto.SeldonMessage, error) {
	seldonPredictorProcess := predictor.NewPredictorProcess(ctx, g.Client, logf.Log.WithName("SeldonMessageRestClient"), grpc.GetEventId(ctx), g.ServerUrl, g.Namespace)
	reqPayload := payload.ProtoPayload{Msg: req}
	resPayload, err := seldonPredictorProcess.Feedback(g.predictor.Graph, &reqPayload)
	if err != nil {
		g.Log.Error(err, "Failed to call feedback")
	}
	return resPayload.GetPayload().(*proto.SeldonMessage), err
}
