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
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type SeldonMessageGrpcClient struct {
	Log logr.Logger
}

func NewSeldonGrpcClient() client.SeldonApiClient {
	smgc := SeldonMessageGrpcClient{
		Log: logf.Log.WithName("SeldonMessageRestClient"),
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
		return nil, err
	}
	grpcClient := proto.NewModelClient(conn)
	resp, err := grpcClient.Predict(context.Background(), msg.GetPayload().(*proto.SeldonMessage))
	if err != nil {
		return nil, err
	}
	reqPayload := payload.SeldonMessagePayload{Msg: resp}
	return &reqPayload, nil
}

func (s SeldonMessageGrpcClient) TransformInput(host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	panic("implement me")
}

func (s SeldonMessageGrpcClient) Route(host string, port int32, msg payload.SeldonPayload) (int, error) {
	panic("implement me")
}

func (s SeldonMessageGrpcClient) Combine(host string, port int32, msgs []payload.SeldonPayload) (payload.SeldonPayload, error) {
	panic("implement me")
}

func (s SeldonMessageGrpcClient) TransformOutput(host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	panic("implement me")
}

func (s SeldonMessageGrpcClient) Unmarshall(msg []byte) (payload.SeldonPayload, error) {
	panic("implement me")
}

func (s SeldonMessageGrpcClient) Marshall(out io.Writer, msg payload.SeldonPayload) error {
	panic("implement me")
}

func (s SeldonMessageGrpcClient) CreateErrorPayload(err error) (payload.SeldonPayload, error) {
	panic("implement me")
}
