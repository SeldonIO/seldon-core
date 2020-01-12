package grpc

import (
	"context"
	"github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"google.golang.org/grpc/metadata"
	"testing"
)

func TestAddPuid(t *testing.T) {
	t.Logf("Started")
	g := gomega.NewGomegaWithT(t)

	ctx := context.Background()
	ctx = addSeldonPuid(ctx)

	g.Expect(ctx.Value(payload.SeldonPUIDHeader)).NotTo(gomega.BeNil())
}

func TestExistingPuid(t *testing.T) {
	t.Logf("Started")
	g := gomega.NewGomegaWithT(t)
	guid := "1"

	ctx := metadata.NewIncomingContext(context.TODO(), metadata.New(map[string]string{payload.SeldonPUIDHeader: guid}))
	ctx = addSeldonPuid(ctx)

	g.Expect(ctx.Value(payload.SeldonPUIDHeader)).NotTo(gomega.BeNil())
	g.Expect(ctx.Value(payload.SeldonPUIDHeader).(string)).To(gomega.Equal(guid))
}
