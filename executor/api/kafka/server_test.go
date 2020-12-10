package kafka

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	. "github.com/onsi/gomega"
	seldon "github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"testing"
)

func TestGetProtoSeldonMessage(t *testing.T) {
	g := NewGomegaWithT(t)

	var sm seldon.SeldonMessage
	var data = `{"data":{"ndarray":[1.1,2]}}`
	jsonpb.UnmarshalString(data, &sm)

	b, err := proto.Marshal(&sm)
	g.Expect(err).To(BeNil())

	sm2, err := getProto("seldon.protos.SeldonMessage", b)
	g.Expect(err).To(BeNil())

	g.Expect(proto.Equal(sm2, &sm)).Should(Equal(true))
}
