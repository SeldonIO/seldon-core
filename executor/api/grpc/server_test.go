package grpc

import (
	"context"
	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"google.golang.org/grpc/metadata"
	"strings"
	"testing"
)

func TestAddPuid(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()
	meta := CollectMetadata(ctx)

	g.Expect(meta[payload.SeldonPUIDHeader]).NotTo(BeNil())
}

func TestExistingPuid(t *testing.T) {
	g := NewGomegaWithT(t)
	guid := "1"

	ctx := metadata.NewIncomingContext(context.TODO(), metadata.New(map[string]string{payload.SeldonPUIDHeader: guid}))
	meta := CollectMetadata(ctx)

	g.Expect(meta[strings.ToLower(payload.SeldonPUIDHeader)]).NotTo(BeNil())
	g.Expect(meta[strings.ToLower(payload.SeldonPUIDHeader)][0]).To(Equal(guid))
}
