package test

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type GrpcSeldonTestServer struct {
	log           logr.Logger
	delay         int
	modelMetadata *proto.SeldonModelMetadata
}

func NewSeldonTestServer(delay int, modelMetadata *proto.SeldonModelMetadata) *GrpcSeldonTestServer {
	return &GrpcSeldonTestServer{
		log:           logf.Log.WithName("GrpcSeldonTestServer"),
		delay:         delay,
		modelMetadata: modelMetadata,
	}
}

func (g GrpcSeldonTestServer) Predict(ctx context.Context, req *proto.SeldonMessage) (*proto.SeldonMessage, error) {
	if g.delay > 0 {
		time.Sleep(time.Duration(g.delay) * time.Second)
	}
	return req, nil
}

func (g GrpcSeldonTestServer) SendFeedback(ctx context.Context, req *proto.Feedback) (*proto.SeldonMessage, error) {
	panic("Not implemented")
}

func (g GrpcSeldonTestServer) Metadata(ctx context.Context, nothing *empty.Empty) (*proto.SeldonModelMetadata, error) {
	return g.modelMetadata, nil
}
