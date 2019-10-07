package client

import (
	api "github.com/seldonio/seldon-core/executor/api/grpc"
)

type SeldonPayload interface {
	GetPayload() interface{}
}

type SeldonMessagePayload struct {
	Msg *api.SeldonMessage
}

func (s *SeldonMessagePayload) GetPayload()  interface{} {
	return s.Msg
}

func (s *SeldonMessagePayload) SetPayload(payload interface{}) {
	s.Msg = payload.(*api.SeldonMessage)
}