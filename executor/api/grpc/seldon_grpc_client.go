package api

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/grpc/proto"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"google.golang.org/grpc"
	"io"
	"math"
	"net/http"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type SeldonMessageGrpcClient struct {
	Log         logr.Logger
	callOptions []grpc.CallOption
	conns       map[string]*grpc.ClientConn
}

func NewSeldonGrpcClient() client.SeldonApiClient {
	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}
	smgc := SeldonMessageGrpcClient{
		Log:         logf.Log.WithName("SeldonMessageRestClient"),
		callOptions: opts,
		conns:       make(map[string]*grpc.ClientConn),
	}
	return smgc
}

func (s SeldonMessageGrpcClient) getConnection(host string, port int32) (*grpc.ClientConn, error) {
	k := fmt.Sprintf("%s:%d", host, port)
	if conn, ok := s.conns[k]; ok {
		return conn, nil
	} else {
		opts := []grpc.DialOption{
			grpc.WithInsecure(),
		}
		if opentracing.IsGlobalTracerRegistered() {
			opts = append(opts, grpc.WithUnaryInterceptor(grpc_opentracing.UnaryClientInterceptor()))
		}
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), opts...)
		if err != nil {
			return nil, err
		}
		s.conns[k] = conn
		return conn, nil
	}
}

func (s SeldonMessageGrpcClient) Predict(ctx context.Context, host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	conn, err := s.getConnection(host, port)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	grpcClient := proto.NewModelClient(conn)
	resp, err := grpcClient.Predict(context.Background(), msg.GetPayload().(*proto.SeldonMessage), s.callOptions...)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	reqPayload := payload.SeldonMessagePayload{Msg: resp}
	return &reqPayload, nil
}

func (s SeldonMessageGrpcClient) TransformInput(ctx context.Context, host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	conn, err := s.getConnection(host, port)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	grpcClient := proto.NewTransformerClient(conn)
	resp, err := grpcClient.TransformInput(context.Background(), msg.GetPayload().(*proto.SeldonMessage), s.callOptions...)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	reqPayload := payload.SeldonMessagePayload{Msg: resp}
	return &reqPayload, nil
}

func (s SeldonMessageGrpcClient) Route(ctx context.Context, host string, port int32, msg payload.SeldonPayload) (int, error) {
	conn, err := s.getConnection(host, port)
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

func (s SeldonMessageGrpcClient) Combine(ctx context.Context, host string, port int32, msgs []payload.SeldonPayload) (payload.SeldonPayload, error) {
	conn, err := s.getConnection(host, port)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	sms := make([]*proto.SeldonMessage, len(msgs))
	for i, sm := range msgs {
		sms[i] = sm.GetPayload().(*proto.SeldonMessage)
	}
	grpcClient := proto.NewCombinerClient(conn)
	sml := proto.SeldonMessageList{SeldonMessages: sms}
	resp, err := grpcClient.Aggregate(context.Background(), &sml, s.callOptions...)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	reqPayload := payload.SeldonMessagePayload{Msg: resp}
	return &reqPayload, nil
}

func (s SeldonMessageGrpcClient) TransformOutput(ctx context.Context, host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	conn, err := s.getConnection(host, port)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	grpcClient := proto.NewOutputTransformerClient(conn)
	resp, err := grpcClient.TransformOutput(context.Background(), msg.GetPayload().(*proto.SeldonMessage), s.callOptions...)
	if err != nil {
		return s.CreateErrorPayload(err), err
	}
	reqPayload := payload.SeldonMessagePayload{Msg: resp}
	return &reqPayload, nil
}

func (s SeldonMessageGrpcClient) Unmarshall(msg []byte) (payload.SeldonPayload, error) {
	panic("Not implemented")
}

func (s SeldonMessageGrpcClient) Marshall(out io.Writer, msg payload.SeldonPayload) error {
	panic("Not implemented")
}

func (s SeldonMessageGrpcClient) CreateErrorPayload(err error) payload.SeldonPayload {
	respFailed := proto.SeldonMessage{Status: &proto.Status{Code: http.StatusInternalServerError, Info: err.Error()}}
	res := payload.SeldonMessagePayload{Msg: &respFailed}
	return &res
}
