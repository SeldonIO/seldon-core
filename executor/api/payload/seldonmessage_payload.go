package payload

import "github.com/seldonio/seldon-core/executor/api/grpc/proto"

// SeldonMessage Payload
type SeldonMessagePayload struct {
	Msg         *proto.SeldonMessage
	ContentType string
}

func (s *SeldonMessagePayload) GetPayload() interface{} {
	return s.Msg
}

func (s *SeldonMessagePayload) GetContentType() string {
	return s.ContentType
}

func (s *SeldonMessagePayload) SetPayload(payload interface{}) {
	s.Msg = payload.(*proto.SeldonMessage)
}

// SeldonMessageList Payload
type SeldonMessageListPayload struct {
	Msg         *proto.SeldonMessageList
	ContentType string
}

func (s *SeldonMessageListPayload) GetPayload() interface{} {
	return s.Msg
}

func (s *SeldonMessageListPayload) GetContentType() string {
	return s.ContentType
}

func (s *SeldonMessageListPayload) SetPayload(payload interface{}) {
	s.Msg = payload.(*proto.SeldonMessageList)
}
