package grpc

import (
	"context"
	"github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"google.golang.org/grpc/metadata"
	"testing"
)

func TestAddPuidToCtx(t *testing.T) {
	t.Logf("Started")
	g := gomega.NewGomegaWithT(t)

	ctx := context.Background()
	meta := CollectMetadata(ctx)
	ctx = AddMetadataToOutgoingGrpcContext(ctx, meta)

	md, ok := metadata.FromOutgoingContext(ctx)
	g.Expect(ok).To(gomega.BeTrue())
	g.Expect(md.Get(payload.SeldonPUIDHeader)).NotTo(gomega.BeNil())

}
