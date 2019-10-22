package api

import (
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/seldonio/seldon-core/executor/api/grpc/proto"
	"testing"
)

func TestSum(t *testing.T) {
	var sm proto.SeldonMessage
	var data = ` {"data":{"ndarray":[1.1,2.0]}}
`
	jsonpb.UnmarshalString(data, &sm)

	ma := jsonpb.Marshaler{}
	msgStr, _ := ma.MarshalToString(&sm)

	fmt.Printf("hello %s", msgStr)
}
