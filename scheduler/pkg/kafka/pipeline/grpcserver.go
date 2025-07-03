/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package pipeline

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	v2 "github.com/seldonio/seldon-core/apis/go/v2/mlops/v2_dataplane"

	status2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/pipeline/status"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/metrics"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

type GatewayGrpcServer struct {
	v2.UnimplementedGRPCInferenceServiceServer
	port                 int
	grpcServer           *grpc.Server
	gateway              PipelineInferer
	logger               log.FieldLogger
	metrics              metrics.PipelineMetricsHandler
	tlsOptions           *util.TLSOptions
	pipelineReadyChecker status2.PipelineReadyChecker
}

func NewGatewayGrpcServer(port int,
	logger log.FieldLogger,
	gateway PipelineInferer,
	metricsHandler metrics.PipelineMetricsHandler,
	tlsOptions *util.TLSOptions,
	piplineReadyChecker status2.PipelineReadyChecker) *GatewayGrpcServer {
	return &GatewayGrpcServer{
		port:                 port,
		gateway:              gateway,
		logger:               logger.WithField("source", "GatewayGrpcServer"),
		metrics:              metricsHandler,
		tlsOptions:           tlsOptions,
		pipelineReadyChecker: piplineReadyChecker,
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
	if g.tlsOptions.TLS {
		opts = append(opts, grpc.Creds(g.tlsOptions.Cert.CreateServerTransportCredentials()))
	}
	opts = append(opts, grpc.MaxConcurrentStreams(maxConcurrentStreams))
	opts = append(opts, grpc.MaxRecvMsgSize(util.GRPCMaxMsgSizeBytes))
	opts = append(opts, grpc.MaxSendMsgSize(util.GRPCMaxMsgSizeBytes))
	opts = append(opts, grpc.StatsHandler(otelgrpc.NewServerHandler()))
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

// Get or create a requestId
func (g *GatewayGrpcServer) getRequestId(md metadata.MD) string {
	requestId := extractHeader(util.RequestIdHeader, md)
	if requestId == "" {
		requestId = util.CreateRequestId()
	}
	return requestId
}

func (g *GatewayGrpcServer) ModelInfer(ctx context.Context, r *v2.ModelInferRequest) (*v2.ModelInferResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.FailedPrecondition, "failed to find any metadata - required %s or %s", util.SeldonModelHeader, util.SeldonInternalModelHeader)
	}
	g.logger.Debugf("Seldon model header %v and seldon internal model header %v", md[util.SeldonModelHeader], md[util.SeldonInternalModelHeader])
	header := extractHeader(util.SeldonInternalModelHeader, md) // Internal model header has precedence
	if header == "" {                                           // If we don't find internal model header fall back on public one
		header = extractHeader(util.SeldonModelHeader, md)
	}
	resourceName, isModel, err := createResourceNameFromHeader(header)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "failed to find valid header %s, found %s", util.SeldonModelHeader, resourceName)
	}

	startTime := time.Now()
	b, err := proto.Marshal(r)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "%s", err.Error())
	}
	kafkaRequest, err := g.gateway.Infer(ctx, resourceName, isModel, b, convertGrpcMetadataToKafkaHeaders(md), g.getRequestId(md))
	elapsedTime := time.Since(startTime).Seconds()
	if err != nil {
		go g.metrics.AddPipelineInferMetrics(resourceName, metrics.MethodTypeGrpc, elapsedTime, codes.FailedPrecondition.String())
		return nil, status.Errorf(codes.FailedPrecondition, "%s", err.Error())
	} else if kafkaRequest.isError {
		go g.metrics.AddPipelineInferMetrics(resourceName, metrics.MethodTypeGrpc, elapsedTime, codes.Unknown.String())
		return nil, status.Errorf(codes.Unknown, "%s", string(createResponseErrorPayload(kafkaRequest.errorModel, kafkaRequest.response)))
	}
	meta := convertKafkaHeadersToGrpcMetadata(kafkaRequest.headers)
	meta[util.RequestIdHeader] = []string{kafkaRequest.key}
	err = grpc.SendHeader(ctx, meta)
	if err != nil {
		go g.metrics.AddPipelineInferMetrics(resourceName, metrics.MethodTypeGrpc, elapsedTime, codes.Internal.String())
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	resProto := &v2.ModelInferResponse{}
	err = proto.Unmarshal(kafkaRequest.response, resProto)
	if err != nil {
		go g.metrics.AddPipelineInferMetrics(resourceName, metrics.MethodTypeGrpc, elapsedTime, codes.Internal.String())
		return nil, status.Errorf(codes.FailedPrecondition, "%s", err.Error())
	}
	go g.metrics.AddPipelineInferMetrics(resourceName, metrics.MethodTypeGrpc, elapsedTime, codes.OK.String())

	return resProto, nil
}

// This is presently used for pipeline ready use cases but the v2 protocol only has the concept of model ready calls
func (g *GatewayGrpcServer) ModelReady(ctx context.Context, req *v2.ModelReadyRequest) (*v2.ModelReadyResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.FailedPrecondition, "failed to find any metadata - required %s or %s", util.SeldonModelHeader, util.SeldonInternalModelHeader)
	}
	ready, err := g.pipelineReadyChecker.CheckPipelineReady(ctx, req.GetName(), g.getRequestId(md))
	if err != nil {
		if errors.Is(err, status2.PipelineNotFoundErr) {
			return nil, status.Errorf(codes.NotFound, "Pipeline not found")
		} else {
			return nil, status.Errorf(codes.Internal, "%s", err.Error())
		}
	}
	return &v2.ModelReadyResponse{Ready: ready}, nil
}
