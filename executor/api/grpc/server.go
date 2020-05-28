package grpc

import (
	"context"
	"math"
	"strconv"

	"github.com/go-logr/logr"
	guuid "github.com/google/uuid"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"github.com/seldonio/seldon-core/executor/api/metric"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/k8s"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	ProtobufContentType = "application/protobuf"
)

func getMaxMsgSizeFromAnnotations(annotations map[string]string) (int, error) {
	val := annotations[k8s.ANNOTATION_GRPC_MAX_MESSAGE_SIZE]
	if val != "" {
		converted, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return 0, err
		} else {
			return int(converted), nil
		}
	} else {
		return 0, nil
	}
}

func CreateGrpcServer(spec *v1.PredictorSpec, deploymentName string, annotations map[string]string, logger logr.Logger) (*grpc.Server, error) {
	maxMsgSize := math.MaxInt32
	// Update from annotations
	if annotations != nil {
		sizeFromAnnotation, err := getMaxMsgSizeFromAnnotations(annotations)
		if err != nil {
			return nil, err
		} else if sizeFromAnnotation > 0 {
			maxMsgSize = sizeFromAnnotation
		}
	}

	logger.Info("Setting max message size ", "size", maxMsgSize)
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
	}

	interceptors := []grpc.UnaryServerInterceptor{metric.NewServerMetrics(spec, deploymentName).UnaryServerInterceptor()}
	if opentracing.IsGlobalTracerRegistered() {
		interceptors = append(interceptors, grpc_opentracing.UnaryServerInterceptor())
	}
	opts = append(opts, grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(interceptors...)))

	grpcServer := grpc.NewServer(opts...)
	return grpcServer, nil
}

func CollectMetadata(ctx context.Context) metadata.MD {
	if mdFromIncoming, ok := metadata.FromIncomingContext(ctx); ok {
		val := mdFromIncoming.Get(payload.SeldonPUIDHeader)
		if len(val) == 0 {
			mdFromIncoming.Set(payload.SeldonPUIDHeader, guuid.New().String())
		}
		return mdFromIncoming
	}
	return metadata.New(map[string]string{payload.SeldonPUIDHeader: guuid.New().String()})
}
