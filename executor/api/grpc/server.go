package grpc

import (
	"context"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"github.com/seldonio/seldon-core/executor/api/metric"
	"github.com/seldonio/seldon-core/executor/logger"
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

// Populate event ID from metadata
func GetEventId(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		vals := md.Get(logger.CloudEventsIdHeader)
		if len(vals) == 1 {
			return vals[0]
		}
	}
	return ""
}
