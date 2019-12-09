package api

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
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
}

func NewSeldonGrpcClient() client.SeldonApiClient {
	smgc := SeldonMessageGrpcClient{
		Log: logf.Log.WithName("SeldonMessageRestClient"),
		callOptions: []grpc.CallOption{
			grpc.MaxCallSendMsgSize(math.MaxInt32),
			grpc.MaxCallRecvMsgSize(math.MaxInt32),
		},
	}
	return smgc
}

func (s SeldonMessageGrpcClient) getConnection(host string, port int32) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (s SeldonMessageGrpcClient) Predict(host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
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

func (s SeldonMessageGrpcClient) TransformInput(host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
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

func (s SeldonMessageGrpcClient) Route(host string, port int32, msg payload.SeldonPayload) (int, error) {
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

func (s SeldonMessageGrpcClient) Combine(host string, port int32, msgs []payload.SeldonPayload) (payload.SeldonPayload, error) {
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

func (s SeldonMessageGrpcClient) TransformOutput(host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
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
