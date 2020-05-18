package grpc

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"google.golang.org/grpc/metadata"
)

func TestAddPuid(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()
	meta := CollectMetadata(ctx)

	g.Expect(meta.Get(payload.SeldonPUIDHeader)).NotTo(BeNil())
}

func TestExistingPuid(t *testing.T) {
	g := NewGomegaWithT(t)
	puid := "1"

	ctx := metadata.NewIncomingContext(context.TODO(), metadata.New(map[string]string{payload.SeldonPUIDHeader: puid}))
	meta := CollectMetadata(ctx)

	g.Expect(meta.Get(payload.SeldonPUIDHeader)).NotTo(BeNil())
	g.Expect(meta.Get(payload.SeldonPUIDHeader)[0]).To(Equal(puid))
}
