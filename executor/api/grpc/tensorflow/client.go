package tensorflow

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/golang/protobuf/proto"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/seldonio/seldon-core/executor/api/client"
	grpc2 "github.com/seldonio/seldon-core/executor/api/grpc"
	"github.com/seldonio/seldon-core/executor/api/metric"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/proto/tensorflow/serving"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	"google.golang.org/grpc"
	"io"
	"math"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type TensorflowGrpcClient struct {
	Log            logr.Logger
	callOptions    []grpc.CallOption
	conns          map[string]*grpc.ClientConn
	Predictor      *v1.PredictorSpec
	DeploymentName string
}

func NewTensorflowGrpcClient(predictor *v1.PredictorSpec, deploymentName string) client.SeldonApiClient {
	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}
	smgc := TensorflowGrpcClient{
		Log:            logf.Log.WithName("SeldonGrpcClient"),
		callOptions:    opts,
		conns:          make(map[string]*grpc.ClientConn),
		Predictor:      predictor,
		DeploymentName: deploymentName,
	}
	return smgc
}

func (s TensorflowGrpcClient) getConnection(host string, port int32, modelName string) (*grpc.ClientConn, error) {
	k := fmt.Sprintf("%s:%d", host, port)
	if conn, ok := s.conns[k]; ok {
		return conn, nil
	} else {
		opts := []grpc.DialOption{
			grpc.WithInsecure(),
		}
		if opentracing.IsGlobalTracerRegistered() {
			opts = append(opts, grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(grpc_opentracing.UnaryClientInterceptor(),
				metric.NewClientMetrics(s.Predictor, s.DeploymentName, modelName).UnaryClientInterceptor())))
		} else {
			opts = append(opts, grpc.WithUnaryInterceptor(metric.NewClientMetrics(s.Predictor, s.DeploymentName, modelName).UnaryClientInterceptor()))
		}
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), opts...)
		if err != nil {
			return nil, err
		}
		s.conns[k] = conn
		return conn, nil
	}
}

// Allow PredictionResponses to be turned into PredictionRequests
func (s TensorflowGrpcClient) Chain(ctx context.Context, modelName string, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	switch v := msg.GetPayload().(type) {
	case *serving.PredictRequest, *serving.ClassificationRequest, *serving.MultiInferenceRequest:
		s.Log.Info("Identity chain")
		return msg, nil
	case *serving.PredictResponse:
		s.Log.Info("Chain!")
		pr := serving.PredictRequest{
			ModelSpec: &serving.ModelSpec{
				Name: modelName,
			},
			Inputs: v.Outputs,
		}
		msg2 := payload.ProtoPayload{Msg: &pr}
		return &msg2, nil
	default:
		return nil, errors.Errorf("Invalid type %v", v)
	}
}

func (s TensorflowGrpcClient) Predict(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	conn, err := s.getConnection(host, port, modelName)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	grpcClient := serving.NewPredictionServiceClient(conn)
	ctx = grpc2.AddSeldonPuidToGrpcContext(ctx)
	var resp proto.Message
	switch v := msg.GetPayload().(type) {
	case *serving.PredictRequest:
		resp, err = grpcClient.Predict(ctx, v, s.callOptions...)
	case *serving.ClassificationRequest:
		resp, err = grpcClient.Classify(ctx, v, s.callOptions...)
	case *serving.MultiInferenceRequest:
		resp, err = grpcClient.MultiInference(ctx, v, s.callOptions...)
	default:
		return nil, errors.Errorf("Invalid type %v", v)
	}
	if err != nil {
		return nil, err
	}
	resPayload := payload.ProtoPayload{Msg: resp}
	return &resPayload, nil
}

func (s TensorflowGrpcClient) TransformInput(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	return s.Predict(ctx, modelName, host, port, msg)
}

func (s TensorflowGrpcClient) Route(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload) (int, error) {
	panic("Not implemented")
}

func (s TensorflowGrpcClient) Combine(ctx context.Context, modelName string, host string, port int32, msgs []payload.SeldonPayload) (payload.SeldonPayload, error) {
	panic("Not implemented")
}

func (s TensorflowGrpcClient) TransformOutput(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	return s.Predict(ctx, modelName, host, port, msg)
}

func (s TensorflowGrpcClient) Status(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	conn, err := s.getConnection(host, port, modelName)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	grpcClient := serving.NewModelServiceClient(conn)
	var resp proto.Message
	resp, err = grpcClient.GetModelStatus(grpc2.AddSeldonPuidToGrpcContext(ctx), msg.GetPayload().(*serving.GetModelStatusRequest), s.callOptions...)
	if err != nil {
		return nil, err
	}
	resPayload := payload.ProtoPayload{Msg: resp}
	return &resPayload, nil
}

func (s TensorflowGrpcClient) Metadata(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	conn, err := s.getConnection(host, port, modelName)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	grpcClient := serving.NewPredictionServiceClient(conn)
	var resp proto.Message
	resp, err = grpcClient.GetModelMetadata(grpc2.AddSeldonPuidToGrpcContext(ctx), msg.GetPayload().(*serving.GetModelMetadataRequest), s.callOptions...)
	if err != nil {
		return nil, err
	}
	resPayload := payload.ProtoPayload{Msg: resp}
	return &resPayload, nil
}

func (s TensorflowGrpcClient) Feedback(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	panic("implement me")
}

func (s TensorflowGrpcClient) Unmarshall(msg []byte) (payload.SeldonPayload, error) {
	panic("Not implemented")
}

func (s TensorflowGrpcClient) Marshall(out io.Writer, msg payload.SeldonPayload) error {
	panic("Not implemented")
}

func (s TensorflowGrpcClient) CreateErrorPayload(err error) payload.SeldonPayload {
	panic("Not implemented")
}
