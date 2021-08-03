package grpc

import (
	"context"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"github.com/seldonio/seldon-core/executor/api/metric"
	"github.com/seldonio/seldon-core/executor/k8s"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"google.golang.org/grpc"
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

func AddClientInterceptors(predictor *v1.PredictorSpec, deploymentName, modelName string, annotations map[string]string, log logr.Logger) grpc.DialOption {
	interceptors := []grpc.UnaryClientInterceptor{metric.NewClientMetrics(predictor, deploymentName, modelName).UnaryClientInterceptor()}
	if opentracing.IsGlobalTracerRegistered() {
		interceptors = append(interceptors, grpc_opentracing.UnaryClientInterceptor())
	}
	if annotations != nil {
		val := annotations[k8s.ANNOTATION_GRPC_TIMEOUT]
		if val != "" {
			timeout, err := strconv.Atoi(val)
			if err != nil {
				log.Error(err, "Failed to parse annotation to int so will ignore", k8s.ANNOTATION_GRPC_TIMEOUT, val)
			} else {
				dur := time.Millisecond * time.Duration(timeout)
				log.Info("Adding grpc timeout to client", "value", timeout, "seconds", dur)

				interceptors = append(interceptors, unaryClientInterceptorWithTimeout(dur))
			}
		}
	}
	return grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(interceptors...))
}

func unaryClientInterceptorWithTimeout(timeout time.Duration) func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		err := invoker(ctxWithTimeout, method, req, reply, cc, opts...)
		return err
	}
}
