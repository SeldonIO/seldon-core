package payload

import (
	"github.com/golang/protobuf/proto"
	grpcSeldon "github.com/seldonio/seldon-core/executor/api/grpc/proto"
)

// SeldonMessage Payload
type SeldonMessagePayload struct {
	Msg         *grpcSeldon.SeldonMessage
	ContentType string
}

func (s *SeldonMessagePayload) GetPayload() interface{} {
	return s.Msg
}

func (s *SeldonMessagePayload) GetBytes() ([]byte, error) {
	data, err := proto.Marshal(s.Msg)
	if err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

func (s *SeldonMessagePayload) GetContentType() string {
	return s.ContentType
}

func (s *SeldonMessagePayload) SetPayload(payload interface{}) {
	s.Msg = payload.(*grpcSeldon.SeldonMessage)
}

// SeldonMessageList Payload
type SeldonMessageListPayload struct {
	Msg         *grpcSeldon.SeldonMessageList
	ContentType string
}

func (s *SeldonMessageListPayload) GetPayload() interface{} {
	return s.Msg
}

func (s *SeldonMessageListPayload) GetContentType() string {
	return s.ContentType
}

func (s *SeldonMessageListPayload) SetPayload(payload interface{}) {
	s.Msg = payload.(*grpcSeldon.SeldonMessageList)
}
