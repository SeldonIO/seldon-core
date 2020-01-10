package seldon

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/metric"
	"github.com/seldonio/seldon-core/executor/api/payload"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	"google.golang.org/grpc"
	"io"
	"math"
	"net/http"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type SeldonMessageGrpcClient struct {
	Log            logr.Logger
	callOptions    []grpc.CallOption
	conns          map[string]*grpc.ClientConn
	Predictor      *v1.PredictorSpec
	DeploymentName string
}

func NewSeldonGrpcClient(spec *v1.PredictorSpec, deploymentName string) client.SeldonApiClient {
	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}
	smgc := SeldonMessageGrpcClient{
		Log:            logf.Log.WithName("SeldonGrpcClient"),
		callOptions:    opts,
		conns:          make(map[string]*grpc.ClientConn),
		Predictor:      spec,
		DeploymentName: deploymentName,
	}
	return smgc
}

func (s SeldonMessageGrpcClient) getConnection(host string, port int32, modelName string) (*grpc.ClientConn, error) {
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

func (s SeldonMessageGrpcClient) Chain(ctx context.Context, modelName string, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	return msg, nil
}

func (s SeldonMessageGrpcClient) Predict(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	conn, err := s.getConnection(host, port, modelName)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	grpcClient := proto.NewModelClient(conn)
	resp, err := grpcClient.Predict(context.Background(), msg.GetPayload().(*proto.SeldonMessage), s.callOptions...)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	resPayload := payload.ProtoPayload{Msg: resp}
	return &resPayload, nil
}

func (s SeldonMessageGrpcClient) TransformInput(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	conn, err := s.getConnection(host, port, modelName)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	grpcClient := proto.NewTransformerClient(conn)
	resp, err := grpcClient.TransformInput(ctx, msg.GetPayload().(*proto.SeldonMessage), s.callOptions...)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	resPayload := payload.ProtoPayload{Msg: resp}
	return &resPayload, nil
}

func (s SeldonMessageGrpcClient) Route(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload) (int, error) {
	conn, err := s.getConnection(host, port, modelName)
	if err != nil {
		return 0, err
	}
	grpcClient := proto.NewRouterClient(conn)
	resp, err := grpcClient.Route(context.Background(), msg.GetPayload().(*proto.SeldonMessage), s.callOptions...)
	if err != nil {
		return 0, err
	}
	routes := ExtractRouteFromSeldonMessage(resp)
	//Only returning first route. API could be extended to allow multiple routes
	return routes[0], nil
}

func (s SeldonMessageGrpcClient) Combine(ctx context.Context, modelName string, host string, port int32, msgs []payload.SeldonPayload) (payload.SeldonPayload, error) {
	conn, err := s.getConnection(host, port, modelName)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	sms := make([]*proto.SeldonMessage, len(msgs))
	for i, sm := range msgs {
		sms[i] = sm.GetPayload().(*proto.SeldonMessage)
	}
	grpcClient := proto.NewCombinerClient(conn)
	sml := proto.SeldonMessageList{SeldonMessages: sms}
	resp, err := grpcClient.Aggregate(ctx, &sml, s.callOptions...)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	resPayload := payload.ProtoPayload{Msg: resp}
	return &resPayload, nil
}

func (s SeldonMessageGrpcClient) TransformOutput(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	conn, err := s.getConnection(host, port, modelName)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	grpcClient := proto.NewOutputTransformerClient(conn)
	resp, err := grpcClient.TransformOutput(ctx, msg.GetPayload().(*proto.SeldonMessage), s.callOptions...)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	resPayload := payload.ProtoPayload{Msg: resp}
	return &resPayload, nil
}

func (s SeldonMessageGrpcClient) Feedback(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	conn, err := s.getConnection(host, port, modelName)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	grpcClient := proto.NewModelClient(conn)
	resp, err := grpcClient.SendFeedback(ctx, msg.GetPayload().(*proto.Feedback), s.callOptions...)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	resPayload := payload.ProtoPayload{Msg: resp}
	return &resPayload, nil
}

func (s SeldonMessageGrpcClient) Unmarshall(msg []byte) (payload.SeldonPayload, error) {
	panic("Not implemented")
}

func (s SeldonMessageGrpcClient) Marshall(out io.Writer, msg payload.SeldonPayload) error {
	panic("Not implemented")
}

func (s SeldonMessageGrpcClient) CreateErrorPayload(err error) payload.SeldonPayload {
	respFailed := proto.SeldonMessage{Status: &proto.Status{Code: http.StatusInternalServerError, Info: err.Error()}}
	res := payload.ProtoPayload{Msg: &respFailed}
	return &res
}

func (s SeldonMessageGrpcClient) Status(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	return nil, errors.Errorf("Not implemented")
}
