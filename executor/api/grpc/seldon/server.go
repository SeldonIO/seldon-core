package seldon

import (
	"context"
	"github.com/go-logr/logr"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/logger"
	"github.com/seldonio/seldon-core/executor/predictor"
	"github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"math"
	"net/url"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	GrpcServerDefaultMaxMessageSize = 1024 * 1024 * 100
)

type GrpcSeldonServer struct {
	Client    client.SeldonApiClient
	predictor *v1.PredictorSpec
	Log       logr.Logger
	ServerUrl *url.URL
	Namespace string
}

func CreateGrpcServer() *grpc.Server {
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(math.MaxInt32),
		grpc.MaxSendMsgSize(math.MaxInt32),
	}
	if opentracing.IsGlobalTracerRegistered() {
		opts = append(opts, grpc.UnaryInterceptor(grpc_opentracing.UnaryServerInterceptor()))
	}
	grpcServer := grpc.NewServer(opts...)
	return grpcServer
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

// Populate event ID from metadata
func getEventId(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		vals := md.Get(logger.CloudEventsIdHeader)
		if len(vals) == 1 {
			return vals[0]
		}
	}
	return ""
}

func (g GrpcSeldonServer) Predict(ctx context.Context, req *proto.SeldonMessage) (*proto.SeldonMessage, error) {
	seldonPredictorProcess := predictor.NewPredictorProcess(ctx, g.Client, logf.Log.WithName("SeldonMessageRestClient"), getEventId(ctx), g.ServerUrl, g.Namespace)
	reqPayload := payload.SeldonMessagePayload{Msg: req}
	resPayload, err := seldonPredictorProcess.Execute(g.predictor.Graph, &reqPayload)
	if err != nil {
		g.Log.Error(err, "Failed to call predict")
	}
	return resPayload.GetPayload().(*proto.SeldonMessage), err
}

func (g GrpcSeldonServer) SendFeedback(context.Context, *proto.Feedback) (*proto.SeldonMessage, error) {
	panic("implement me")
}
