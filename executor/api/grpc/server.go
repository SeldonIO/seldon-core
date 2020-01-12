package grpc

import (
	"context"
	guuid "github.com/google/uuid"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"github.com/seldonio/seldon-core/executor/api/metric"
	"github.com/seldonio/seldon-core/executor/api/payload"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"math"
)

const (
	ProtobufContentType = "application/protobuf"
)

func CreateGrpcServer(spec *v1.PredictorSpec, deploymentName string) *grpc.Server {
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(math.MaxInt32),
		grpc.MaxSendMsgSize(math.MaxInt32),
	}
	if opentracing.IsGlobalTracerRegistered() {
		opts = append(opts, grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(seldonPuidUnaryInterceptor(), grpc_opentracing.UnaryServerInterceptor(), metric.NewServerMetrics(spec, deploymentName).UnaryServerInterceptor())))
	} else {
		opts = append(opts, grpc.UnaryInterceptor(metric.NewServerMetrics(spec, deploymentName).UnaryServerInterceptor()))
	}

	grpcServer := grpc.NewServer(opts...)
	return grpcServer
}

// Add Seldon Puid if missing to context
func addSeldonPuid(ctx context.Context) context.Context {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		vals := md.Get(payload.SeldonPUIDHeader)
		if len(vals) == 1 {
			return context.WithValue(ctx, payload.SeldonPUIDHeader, vals[0])
		}
	}
	return context.WithValue(ctx, payload.SeldonPUIDHeader, guuid.New().String())
}

func seldonPuidUnaryInterceptor() func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = addSeldonPuid(ctx)
		resp, err := handler(ctx, req)
		return resp, err
	}
}
