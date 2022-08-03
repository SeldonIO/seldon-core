package pipeline

import (
	"context"
	"fmt"
	"net"
	"time"

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
	metrics    metrics.PipelineMetricsHandler
}

func NewGatewayGrpcServer(port int, logger log.FieldLogger, gateway PipelineInferer, metricsHandler metrics.PipelineMetricsHandler) *GatewayGrpcServer {
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
	opts = append(opts, grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()))
	g.grpcServer = grpc.NewServer(opts...)
	v2.RegisterGRPCInferenceServiceServer(g.grpcServer, g)
	return g.grpcServer.Serve(l)
}

func extractHeader(key string, md metadata.MD) string {
	values, ok := md[key]
	if ok {
		if len(values) > 0 {
			// note if there are more than one elements we just return the last one assuming that was added last
			return values[len(values)-1]
		}
	}
	return ""
}

func (g *GatewayGrpcServer) ModelInfer(ctx context.Context, r *v2.ModelInferRequest) (*v2.ModelInferResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("failed to find metadata looking for %s", resources.SeldonModelHeader))
	}
	g.logger.Debugf("Seldon model header %v and seldon internal model header %v", md[resources.SeldonModelHeader], md[resources.SeldonInternalModelHeader])
	header := extractHeader(resources.SeldonInternalModelHeader, md) // Internal model header has precedence
	if header == "" {                                                // If we don't find internal model header fall back on public one
		header = extractHeader(resources.SeldonModelHeader, md)
	}
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
	elapsedTime := time.Since(startTime).Seconds()
	if err != nil {
		go g.metrics.AddPipelineInferMetrics(resourceName, metrics.MethodTypeGrpc, elapsedTime, codes.FailedPrecondition.String())
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	meta := convertKafkaHeadersToGrpcMetadata(kafkaRequest.headers)
	meta[RequestIdHeader] = []string{kafkaRequest.key}
	err = grpc.SendHeader(ctx, meta)
	if err != nil {
		go g.metrics.AddPipelineInferMetrics(resourceName, metrics.MethodTypeGrpc, elapsedTime, codes.Internal.String())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	resProto := &v2.ModelInferResponse{}
	err = proto.Unmarshal(kafkaRequest.response, resProto)
	if err != nil {
		go g.metrics.AddPipelineInferMetrics(resourceName, metrics.MethodTypeGrpc, elapsedTime, codes.Internal.String())
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	go g.metrics.AddPipelineInferMetrics(resourceName, metrics.MethodTypeGrpc, elapsedTime, codes.OK.String())

	return resProto, nil
}
