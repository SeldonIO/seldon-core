package kfserving

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/seldonio/seldon-core/executor/api/client"
	grpc2 "github.com/seldonio/seldon-core/executor/api/grpc"
	"github.com/seldonio/seldon-core/executor/api/grpc/kfserving/inference"
	"github.com/seldonio/seldon-core/executor/api/payload"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"google.golang.org/grpc"
	"io"
	"math"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type KFServingGrpcClient struct {
	Log            logr.Logger
	callOptions    []grpc.CallOption
	conns          map[string]*grpc.ClientConn
	Predictor      *v1.PredictorSpec
	DeploymentName string
	annotations    map[string]string
}

func (s *KFServingGrpcClient) IsGrpc() bool {
	return true
}

func (s *KFServingGrpcClient) ModelMetadata(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.ModelMetadata, error) {
	panic("implement me")
}

func NewKFServingGrpcClient(predictor *v1.PredictorSpec, deploymentName string, annotations map[string]string) client.SeldonApiClient {
	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}
	smgc := KFServingGrpcClient{
		Log:            logf.Log.WithName("SeldonGrpcClient"),
		callOptions:    opts,
		conns:          make(map[string]*grpc.ClientConn),
		Predictor:      predictor,
		DeploymentName: deploymentName,
		annotations:    annotations,
	}
	return &smgc
}

func (s *KFServingGrpcClient) getConnection(host string, port int32, modelName string) (*grpc.ClientConn, error) {
	k := fmt.Sprintf("%s:%d", host, port)
	if conn, ok := s.conns[k]; ok {
		return conn, nil
	} else {
		opts := []grpc.DialOption{
			grpc.WithInsecure(),
		}
		opts = append(opts, grpc2.AddClientInterceptors(s.Predictor, s.DeploymentName, modelName, s.annotations, s.Log))
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), opts...)
		if err != nil {
			return nil, err
		}
		s.conns[k] = conn
		return conn, nil
	}
}

func (s *KFServingGrpcClient) Predict(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	conn, err := s.getConnection(host, port, modelName)
	if err != nil {
		return nil, err
	}
	grpcClient := inference.NewGRPCInferenceServiceClient(conn)
	ctx = grpc2.AddMetadataToOutgoingGrpcContext(ctx, meta)
	var resp *inference.ModelInferResponse
	switch v := msg.GetPayload().(type) {
	case *inference.ModelInferRequest:
		resp, err = grpcClient.ModelInfer(ctx, v, s.callOptions...)
	default:
		return nil, errors.Errorf("Invalid type %v", v)
	}
	if err != nil {
		return nil, err
	}
	resPayload := payload.ProtoPayload{Msg: resp}
	return &resPayload, nil
}

func (s *KFServingGrpcClient) TransformInput(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	panic("implement me")
}

func (s *KFServingGrpcClient) Route(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (int, error) {
	panic("implement me")
}

func (s *KFServingGrpcClient) Combine(ctx context.Context, modelName string, host string, port int32, msgs []payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	panic("implement me")
}

func (s *KFServingGrpcClient) TransformOutput(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	panic("implement me")
}

func (s *KFServingGrpcClient) Feedback(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	panic("implement me")
}

func (s *KFServingGrpcClient) Chain(ctx context.Context, modelName string, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	switch v := msg.GetPayload().(type) {
	case *inference.ModelInferRequest:
		s.Log.Info("Identity chain")
		return msg, nil
	case *inference.ModelInferResponse:
		s.Log.Info("Chain!")
		inputTensors := make([]*inference.ModelInferRequest_InferInputTensor, len(v.Outputs))
		for _, oTensor := range v.Outputs {
			inputTensor := &inference.ModelInferRequest_InferInputTensor{
				Name:       oTensor.Name,
				Datatype:   oTensor.Datatype,
				Shape:      oTensor.Shape,
				Parameters: oTensor.Parameters,
				Contents:   oTensor.Contents,
			}
			inputTensors = append(inputTensors, inputTensor)
		}
		pr := inference.ModelInferRequest{
			Inputs: inputTensors,
		}
		msg2 := payload.ProtoPayload{Msg: &pr}
		return &msg2, nil
	default:
		return nil, errors.Errorf("Invalid type %v", v)
	}
}

func (s *KFServingGrpcClient) Status(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	conn, err := s.getConnection(host, port, modelName)
	if err != nil {
		return nil, err
	}
	grpcClient := inference.NewGRPCInferenceServiceClient(conn)
	ctx = grpc2.AddMetadataToOutgoingGrpcContext(ctx, meta)
	var resp *inference.ModelReadyResponse
	switch v := msg.GetPayload().(type) {
	case *inference.ModelReadyRequest:
		resp, err = grpcClient.ModelReady(ctx, v, s.callOptions...)
	default:
		return nil, errors.Errorf("Invalid type %v", v)
	}
	if err != nil {
		return nil, err
	}
	resPayload := payload.ProtoPayload{Msg: resp}
	return &resPayload, nil
}

func (s *KFServingGrpcClient) Metadata(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	conn, err := s.getConnection(host, port, modelName)
	if err != nil {
		return nil, err
	}
	grpcClient := inference.NewGRPCInferenceServiceClient(conn)
	ctx = grpc2.AddMetadataToOutgoingGrpcContext(ctx, meta)
	var resp *inference.ModelMetadataResponse
	switch v := msg.GetPayload().(type) {
	case *inference.ModelMetadataRequest:
		resp, err = grpcClient.ModelMetadata(ctx, v, s.callOptions...)
	default:
		return nil, errors.Errorf("Invalid type %v", v)
	}
	if err != nil {
		return nil, err
	}
	resPayload := payload.ProtoPayload{Msg: resp}
	return &resPayload, nil
}

func (s *KFServingGrpcClient) Unmarshall(msg []byte, contentType string) (payload.SeldonPayload, error) {
	panic("implement me")
}

func (s *KFServingGrpcClient) Marshall(out io.Writer, msg payload.SeldonPayload) error {
	panic("implement me")
}

func (s *KFServingGrpcClient) CreateErrorPayload(err error) payload.SeldonPayload {
	panic("implement me")
}
