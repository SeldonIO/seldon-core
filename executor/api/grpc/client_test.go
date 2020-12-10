package grpc

import (
	"context"
	"github.com/golang/protobuf/jsonpb"
	proto2 "github.com/golang/protobuf/proto"
	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"google.golang.org/grpc/metadata"
	"reflect"
	"testing"
)

func TestAddPuidToCtx(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	ctx := context.Background()
	meta := CollectMetadata(ctx)
	ctx = AddMetadataToOutgoingGrpcContext(ctx, meta)

	md, ok := metadata.FromOutgoingContext(ctx)
	g.Expect(ok).To(BeTrue())
	g.Expect(md.Get(payload.SeldonPUIDHeader)).NotTo(BeNil())

}

func getProto(messageType string, messageBytes []byte) proto2.Message {
	pbtype := proto2.MessageType(messageType)
	msg := reflect.New(pbtype.Elem()).Interface().(proto2.Message)
	proto2.Unmarshal(messageBytes, msg)
	return msg
}

func TestUnmarshal(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	var sm proto.SeldonMessage
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`
	err := jsonpb.UnmarshalString(data, &sm)
	g.Expect(err).Should(BeNil())

	tyName := proto2.MessageName(&sm)
	b, err := proto2.Marshal(&sm)
	g.Expect(err).Should(BeNil())

	sm2 := getProto(tyName, b)

	m := jsonpb.Marshaler{}
	sm2Str, err := m.MarshalToString(sm2)
	g.Expect(err).Should(BeNil())
	smStr, err := m.MarshalToString(&sm)
	g.Expect(err).Should(BeNil())
	g.Expect(sm2Str).To(Equal(smStr))
}
