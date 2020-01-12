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
	ctx = addSeldonPuid(ctx)

	g.Expect(ctx.Value(payload.SeldonPUIDHeader)).NotTo(gomega.BeNil())

	ctx = AddSeldonPuidToGrpcContext(ctx)

	md, ok := metadata.FromOutgoingContext(ctx)
	g.Expect(ok).To(gomega.Equal(true))
	g.Expect(md.Get(payload.SeldonPUIDHeader)).NotTo(gomega.BeNil())
}
