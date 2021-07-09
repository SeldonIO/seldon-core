package seldon

import (
	"context"
	"github.com/go-logr/logr"
	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/grpc"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/predictor"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	protoGrpc "google.golang.org/grpc"
	protoGrpcMetadata "google.golang.org/grpc/metadata"
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
	md := grpc.CollectMetadata(ctx)
	header := protoGrpcMetadata.Pairs(payload.SeldonPUIDHeader, md.Get(payload.SeldonPUIDHeader)[0])
	protoGrpc.SendHeader(ctx, header)
	ctx = context.WithValue(ctx, payload.SeldonPUIDHeader, md.Get(payload.SeldonPUIDHeader)[0])
	seldonPredictorProcess := predictor.NewPredictorProcess(ctx, g.Client, logf.Log.WithName("SeldonMessageRestClient"), g.ServerUrl, g.Namespace, md, "")
	reqPayload := payload.ProtoPayload{Msg: req}
	resPayload, err := seldonPredictorProcess.Predict(&g.predictor.Graph, &reqPayload)
	if err != nil {
		g.Log.Error(err, "Failed to call predict")
		return payloadToMessage(resPayload), err
	}
	return payloadToMessage(resPayload), nil
}

func (g GrpcSeldonServer) SendFeedback(ctx context.Context, req *proto.Feedback) (*proto.SeldonMessage, error) {
	md := grpc.CollectMetadata(ctx)
	header := protoGrpcMetadata.Pairs(payload.SeldonPUIDHeader, md.Get(payload.SeldonPUIDHeader)[0])
	protoGrpc.SetHeader(ctx, header)
	seldonPredictorProcess := predictor.NewPredictorProcess(ctx, g.Client, logf.Log.WithName("SeldonMessageRestClient"), g.ServerUrl, g.Namespace, md, "")
	reqPayload := payload.ProtoPayload{Msg: req}
	resPayload, err := seldonPredictorProcess.Feedback(&g.predictor.Graph, &reqPayload)
	if err != nil {
		g.Log.Error(err, "Failed to call feedback")
		return payloadToMessage(resPayload), err
	}
	return payloadToMessage(resPayload), nil
}

func (g GrpcSeldonServer) ModelMetadata(ctx context.Context, req *proto.SeldonModelMetadataRequest) (*proto.SeldonModelMetadata, error) {
	seldonPredictorProcess := predictor.NewPredictorProcess(ctx, g.Client, logf.Log.WithName("SeldonMessageRestClient"), g.ServerUrl, g.Namespace, grpc.CollectMetadata(ctx), req.GetName())
	resPayload, err := seldonPredictorProcess.Metadata(&g.predictor.Graph, req.GetName(), nil)
	if err != nil {
		return nil, err
	}
	return payloadToModelMetadata(resPayload), nil
}

func (g GrpcSeldonServer) GraphMetadata(ctx context.Context, req *empty.Empty) (*proto.SeldonGraphMetadata, error) {

	seldonPredictorProcess := predictor.NewPredictorProcess(ctx, g.Client, logf.Log.WithName("SeldonMessageRestClient"), g.ServerUrl, g.Namespace, grpc.CollectMetadata(ctx), "")

	graphMetadata, err := seldonPredictorProcess.GraphMetadata(g.predictor)
	if err != nil {
		return nil, err
	}

	output := &proto.SeldonGraphMetadata{
		Name:    graphMetadata.Name,
		Inputs:  graphMetadata.GraphInputs.([]*proto.SeldonMessageMetadata),
		Outputs: graphMetadata.GraphOutputs.([]*proto.SeldonMessageMetadata),
	}
	output.Models = map[string]*proto.SeldonModelMetadata{}
	for name, modelMetadata := range graphMetadata.Models {
		output.Models[name] = &proto.SeldonModelMetadata{
			Name:     modelMetadata.Name,
			Versions: modelMetadata.Versions,
			Platform: modelMetadata.Platform,
			Inputs:   modelMetadata.Inputs.([]*proto.SeldonMessageMetadata),
			Outputs:  modelMetadata.Outputs.([]*proto.SeldonMessageMetadata),
			Custom:   modelMetadata.Custom,
		}
	}

	return output, nil
}

func payloadToMessage(p payload.SeldonPayload) *proto.SeldonMessage {
	if m, ok := p.GetPayload().(*proto.SeldonMessage); ok {
		return m
	}
	return nil
}

func payloadToModelMetadata(p payload.SeldonPayload) *proto.SeldonModelMetadata {
	if m, ok := p.GetPayload().(*proto.SeldonModelMetadata); ok {
		return m
	}
	return nil
}
