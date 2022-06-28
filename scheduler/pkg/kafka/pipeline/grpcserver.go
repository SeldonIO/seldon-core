package pipeline

import (
	"context"
	"fmt"
	"net"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"

	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"
	"github.com/seldonio/seldon-core/scheduler/pkg/metrics"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc/metadata"

	v2 "github.com/seldonio/seldon-core/scheduler/apis/mlops/v2_dataplane"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type GatewayGrpcServer struct {
	v2.UnimplementedGRPCInferenceServiceServer
	port       int
	grpcServer *grpc.Server
	gateway    PipelineInferer
	logger     log.FieldLogger
	metrics    metrics.MetricsHandler
}

func NewGatewayGrpcServer(port int, logger log.FieldLogger, gateway PipelineInferer, metricsHandler metrics.MetricsHandler) *GatewayGrpcServer {
	return &GatewayGrpcServer{
		port:    port,
		gateway: gateway,
		logger:  logger.WithField("source", "GatewayGrpcServer"),
		metrics: metricsHandler,
	}
}

const (
	maxConcurrentStreams = 1_000_000
)

func (g *GatewayGrpcServer) Stop() {
	g.grpcServer.Stop()
}

func (g *GatewayGrpcServer) Start() error {
	logger := g.logger.WithField("func", "Start")
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", g.port))
	if err != nil {
		logger.Errorf("unable to start gRPC listening server on port %d", g.port)
		return err
	}
	logger.Infof("Starting grpc server on port %d", g.port)
	opts := []grpc.ServerOption{}
	opts = append(opts, grpc.MaxConcurrentStreams(maxConcurrentStreams))
	opts = append(opts, grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(otelgrpc.UnaryServerInterceptor(), g.metrics.UnaryServerInterceptor())))
	g.grpcServer = grpc.NewServer(opts...)
	v2.RegisterGRPCInferenceServiceServer(g.grpcServer, g)
	return g.grpcServer.Serve(l)
}

func extractHeader(key string, md metadata.MD) string {
	values, ok := md[key]
	if ok {
		if len(values) > 0 {
			// note if there are more than one elements we just return the first one
			return values[0]
		}
	}
	return ""
}

func (g *GatewayGrpcServer) ModelInfer(ctx context.Context, r *v2.ModelInferRequest) (*v2.ModelInferResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("failed to find metadata looking for %s", resources.SeldonModelHeader))
	}
	header := extractHeader(resources.SeldonModelHeader, md)
	resourceName, isModel, err := createResourceNameFromHeader(header)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("failed to find valid header %s, found %s", resources.SeldonModelHeader, resourceName))
	}

	startTime := time.Now()
	b, err := proto.Marshal(r)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	kafkaRequest, err := g.gateway.Infer(ctx, resourceName, isModel, b, convertGrpcMetadataToKafkaHeaders(md))
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	meta := convertKafkaHeadersToGrpcMetadata(kafkaRequest.headers)
	meta[RequestIdHeader] = []string{kafkaRequest.key}
	err = grpc.SendHeader(ctx, meta)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	resProto := &v2.ModelInferResponse{}
	err = proto.Unmarshal(kafkaRequest.response, resProto)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	elapsedTime := time.Since(startTime).Seconds()
	go g.metrics.AddInferMetrics(header, "", metrics.MethodTypeGrpc, elapsedTime)

	return resProto, nil
}
