package grpc

import (
	"context"
	guuid "github.com/google/uuid"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"github.com/seldonio/seldon-core/executor/api/metric"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/k8s"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"math"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strconv"
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

func CreateGrpcServer(spec *v1.PredictorSpec, deploymentName string) (*grpc.Server, error) {
	log := logf.Log.WithName("grpcServer")
	maxMsgSize := math.MaxInt32
	annotations, err := k8s.GetAnnotations()
	if err != nil {
		log.Error(err, "Failed to load annotations")
	}
	// Update from annotations
	if annotations != nil {
		sizeFromAnnotation, err := getMaxMsgSizeFromAnnotations(annotations)
		if err != nil {
			return nil, err
		} else if sizeFromAnnotation > 0 {
			maxMsgSize = sizeFromAnnotation
		}
	}

	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
	}

	interceptors := []grpc.UnaryServerInterceptor{metric.NewServerMetrics(spec, deploymentName).UnaryServerInterceptor()}
	if opentracing.IsGlobalTracerRegistered() {
		interceptors = append(interceptors, grpc_opentracing.UnaryServerInterceptor())
	}
	opts = append(opts, grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(interceptors...)))

	//if opentracing.IsGlobalTracerRegistered() {
	//		opts = append(opts, grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(grpc_opentracing.UnaryServerInterceptor(), metric.NewServerMetrics(spec, deploymentName).UnaryServerInterceptor())))
	//} else {
	//	opts = append(opts, grpc.UnaryInterceptor(metric.NewServerMetrics(spec, deploymentName).UnaryServerInterceptor()))
	//}
	//opts = append(opts, grpc.UnaryInterceptor(metric.NewServerMetrics(spec, deploymentName).UnaryServerInterceptor()))
	//

	grpcServer := grpc.NewServer(opts...)
	return grpcServer, nil
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
