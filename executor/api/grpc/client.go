package grpc

import (
	"context"
	"google.golang.org/grpc/metadata"
)

func AddMetadataToOutgoingGrpcContext(ctx context.Context, meta map[string][]string) context.Context {
	for k, vv := range meta {
		for _, v := range vv {
			ctx = metadata.AppendToOutgoingContext(ctx, k, v)
		}
	}
	return ctx
}
