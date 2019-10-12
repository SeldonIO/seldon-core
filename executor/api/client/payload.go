package client

import (
	api "github.com/seldonio/seldon-core/executor/api/grpc"
)

type SeldonPayload interface {
	GetPayload() interface{}
}

// SeldonMessage Payload
type SeldonMessagePayload struct {
	Msg *api.SeldonMessage
}

func (s *SeldonMessagePayload) GetPayload()  interface{} {
	return s.Msg
}

func (s *SeldonMessagePayload) SetPayload(payload interface{}) {
	s.Msg = payload.(*api.SeldonMessage)
}

// SeldonMessageList Payload
type SeldonMessageListPayload struct {
	Msg *api.SeldonMessageList
}

func (s *SeldonMessageListPayload) GetPayload()  interface{} {
	return s.Msg
}

func (s *SeldonMessageListPayload) SetPayload(payload interface{}) {
	s.Msg = payload.(*api.SeldonMessageList)
}