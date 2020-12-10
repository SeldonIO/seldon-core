package payload

import "github.com/golang/protobuf/proto"

const (
	APPLICATION_TYPE_PROTOBUF = "application/protobuf"
)

type ProtoPayload struct {
	Msg proto.Message
}

func (s *ProtoPayload) GetPayload() interface{} {
	return s.Msg
}

func (s *ProtoPayload) GetContentType() string {
	return APPLICATION_TYPE_PROTOBUF
}

func (s *ProtoPayload) GetBytes() ([]byte, error) {
	data, err := proto.Marshal(s.Msg)
	if err != nil {
		return nil, err
	} else {
		return data, nil
	}
}
