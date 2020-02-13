package payload

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"gotest.tools/assert"
	"testing"
)

func TestGetPayload(t *testing.T) {
	var sm proto.SeldonMessage
	var data = `{"data":{"ndarray":[1.1,2]}}`
	jsonpb.UnmarshalString(data, &sm)

	var sp SeldonPayload = &ProtoPayload{&sm}

	var sm2 *proto.SeldonMessage
	sm2 = sp.GetPayload().(*proto.SeldonMessage)

	ma := jsonpb.Marshaler{}
	msgStr, _ := ma.MarshalToString(sm2)
	assert.Equal(t, data, msgStr)
}
