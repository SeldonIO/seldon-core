package client

import (
	"github.com/golang/protobuf/jsonpb"
	api "github.com/seldonio/seldon-core/executor/api/grpc"
	"gotest.tools/assert"
	"testing"
)

func TestGetPayload(t *testing.T) {
	var sm api.SeldonMessage
	var data = `{"data":{"ndarray":[1.1,2]}}`
	jsonpb.UnmarshalString(data, &sm)

	var sp SeldonPayload = &SeldonMessagePayload{&sm}

	var sm2 *api.SeldonMessage
	sm2 = sp.GetPayload().(*api.SeldonMessage)

	ma := jsonpb.Marshaler{}
	msgStr, _ := ma.MarshalToString(sm2)
	assert.Equal(t, data, msgStr)
}
