package payload

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	. "github.com/onsi/gomega"
	seldon "github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"testing"
)

func TestGetProtoPayload(t *testing.T) {
	g := NewGomegaWithT(t)
	var sm seldon.SeldonMessage
	var data = `{"data":{"ndarray":[1.1,2]}}`
	jsonpb.UnmarshalString(data, &sm)

	payload := ProtoPayload{Msg: &sm}
	b, err := payload.GetBytes()
	g.Expect(err).Should(BeNil())
	var sm2 seldon.SeldonMessage
	proto.Unmarshal(b, &sm2)

	g.Expect(proto.Equal(&sm2, &sm)).Should(Equal(true))

}
