package payload

import (
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"gotest.tools/assert"
)

func TestGetPayload(t *testing.T) {
	var sm proto.SeldonMessage
	var data = `{"data":{"ndarray":[1.1,2]}}`
	err := jsonpb.UnmarshalString(data, &sm)
	assert.NilError(t, err)

	var sp SeldonPayload = &ProtoPayload{&sm}

	var sm2 = sp.GetPayload().(*proto.SeldonMessage)

	ma := jsonpb.Marshaler{}
	msgStr, _ := ma.MarshalToString(sm2)
	assert.Equal(t, data, msgStr)
}
