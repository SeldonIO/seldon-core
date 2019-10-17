package api

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/golang/protobuf/jsonpb"
)

type GrpcSeldonServer struct {
	Log logr.Logger
}

func NewGrpcSeldonServer(logger logr.Logger) *GrpcSeldonServer {
	return &GrpcSeldonServer{Log: logger}
}

func (g GrpcSeldonServer) Predict(context.Context, *SeldonMessage) (*SeldonMessage, error) {
	var sm SeldonMessage
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`
	err := jsonpb.UnmarshalString(data, &sm)
	if err != nil {
		return nil, err
	}
	return &sm, nil
}

func (g GrpcSeldonServer) SendFeedback(context.Context, *Feedback) (*SeldonMessage, error) {
	panic("implement me")
}
