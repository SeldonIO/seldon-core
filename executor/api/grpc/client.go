package grpc

import (
	"context"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"google.golang.org/grpc/metadata"
)

func AddSeldonPuidToGrpcContext(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, payload.SeldonPUIDHeader, ctx.Value(payload.SeldonPUIDHeader).(string))
}
