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
		opts = append(opts, grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(grpc_opentracing.UnaryServerInterceptor(), metric.NewServerMetrics(spec, deploymentName).UnaryServerInterceptor())))
	} else {
		opts = append(opts, grpc.UnaryInterceptor(metric.NewServerMetrics(spec, deploymentName).UnaryServerInterceptor()))
	}

	grpcServer := grpc.NewServer(opts...)
	return grpcServer
}

func CollectMetadata(ctx context.Context) map[string][]string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		val := md.Get(payload.SeldonPUIDHeader)
		if len(val) == 0 {
			md.Set(payload.SeldonPUIDHeader, guuid.New().String())
		}
		return md
	} else {
		return map[string][]string{payload.SeldonPUIDHeader: []string{guuid.New().String()}}
	}
}
