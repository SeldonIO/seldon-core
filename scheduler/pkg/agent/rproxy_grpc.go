/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package agent

// reverse proxy for grpc infer and metadata endpoints

import (
	"context"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"sync"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	v2 "github.com/seldonio/seldon-core/apis/go/v2/mlops/v2_dataplane"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/modelscaling"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/modelserver_controlplane/oip"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/metrics"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

const (
	ReverseGRPCProxyPort          = 9998
	grpcProxyMaxConcurrentStreams = 1000
	maxConnsPerHostGRPC           = 10
)

type reverseGRPCProxy struct {
	v2.UnimplementedGRPCInferenceServiceServer
	logger                     log.FieldLogger
	stateManager               *LocalStateManager
	grpcServer                 *grpc.Server
	serverReady                bool
	backendGRPCServerHost      string
	backendGRPCServerPort      uint
	v2GRPCClientPool           []v2.GRPCInferenceServiceClient
	port                       uint // service port
	mu                         sync.RWMutex
	metrics                    metrics.AgentMetricsHandler
	callOptions                []grpc.CallOption
	tlsOptions                 util.TLSOptions
	modelScalingStatsCollector *modelscaling.DataPlaneStatsCollector
}

func NewReverseGRPCProxy(
	metricsHandler metrics.AgentMetricsHandler,
	logger log.FieldLogger, backendGRPCServerHost string,
	backendGRPCServerPort uint,
	servicePort uint,
	modelScalingStatsCollector *modelscaling.DataPlaneStatsCollector,
) *reverseGRPCProxy {
	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}
	return &reverseGRPCProxy{
		logger:                     logger.WithField("Source", "GRPCProxy"),
		backendGRPCServerHost:      backendGRPCServerHost,
		backendGRPCServerPort:      backendGRPCServerPort,
		port:                       servicePort,
		metrics:                    metricsHandler,
		callOptions:                opts,
		modelScalingStatsCollector: modelScalingStatsCollector,
	}
}

func (rp *reverseGRPCProxy) SetState(sm interface{}) {
	rp.stateManager = sm.(*LocalStateManager)
}

func (rp *reverseGRPCProxy) Start() error {
	var err error
	rp.tlsOptions, err = util.CreateUpstreamDataplaneServerTLSOptions()
	if err != nil {
		return err
	}
	if rp.stateManager == nil {
		rp.logger.Error("Set state before starting reverse proxy service")
		return fmt.Errorf("State not set, aborting")
	}

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", rp.port))
	if err != nil {
		rp.logger.Errorf("unable to start gRPC listening server on port %d", rp.port)
		return err
	}

	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(math.MaxInt32),
		grpc.MaxSendMsgSize(math.MaxInt32),
	}
	if rp.tlsOptions.TLS {
		opts = append(opts, grpc.Creds(rp.tlsOptions.Cert.CreateServerTransportCredentials()))
	}
	opts = append(opts, grpc.MaxConcurrentStreams(grpcProxyMaxConcurrentStreams))
	opts = append(opts, grpc.MaxRecvMsgSize(util.GRPCMaxMsgSizeBytes))
	opts = append(opts, grpc.MaxSendMsgSize(util.GRPCMaxMsgSizeBytes))
	opts = append(opts, grpc.StatsHandler(otelgrpc.NewServerHandler()))
	opts = append(opts, grpc.UnaryInterceptor(rp.metrics.UnaryServerInterceptor()))
	grpcServer := grpc.NewServer(opts...)
	v2.RegisterGRPCInferenceServiceServer(grpcServer, rp)

	rp.logger.Infof("Setting grpc v2 client pool on port %d", rp.backendGRPCServerPort)

	conns, clients, err := rp.createV2CRPCClients(rp.backendGRPCServerHost, int(rp.backendGRPCServerPort), maxConnsPerHostGRPC)
	if err != nil {
		return err
	}

	rp.v2GRPCClientPool = clients

	rp.logger.Infof("Starting gRPC listening server on port %d", rp.port)
	rp.grpcServer = grpcServer
	go func() {
		defer closeV2CRPCConnections(conns)
		rp.mu.Lock()
		rp.serverReady = true
		rp.mu.Unlock()
		err := rp.grpcServer.Serve(l)
		rp.logger.WithError(err).Infof("Reverse GRPC proxy stopped with error")
		rp.mu.Lock()
		rp.serverReady = false
		rp.mu.Unlock()
	}()
	return nil
}

func (rp *reverseGRPCProxy) Stop() error {
	rp.logger.Info("Start graceful shutdown")
	// Shutdown is graceful
	rp.mu.Lock()
	defer rp.mu.Unlock()
	if rp.grpcServer != nil {
		rp.grpcServer.GracefulStop()
	}
	rp.serverReady = false
	rp.logger.Info("Finished graceful shutdown")
	return nil
}

func (rp *reverseGRPCProxy) Ready() bool {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	return rp.serverReady
}

func (rp *reverseGRPCProxy) Name() string {
	return "Reverse GRPC Proxy"
}

func (rp *reverseGRPCProxy) GetType() interfaces.SubServiceType {
	return interfaces.CriticalDataPlaneService
}

func (rp *reverseGRPCProxy) extractModelNamesFromContext(ctx context.Context) (string, string, error) {
	var internalModelName, externalModelName string
	var inHeader bool
	if internalModelName, externalModelName, inHeader = extractModelNamesFromHeaders(ctx); inHeader {
		rp.logger.Debugf("Extracted model name %s:%s %s:%s", util.SeldonInternalModelHeader, internalModelName, util.SeldonModelHeader, externalModelName)
		return internalModelName, externalModelName, nil
	} else {
		msg := fmt.Sprintf("Failed to extract model name %s:[%s] %s:[%s]", util.SeldonInternalModelHeader, internalModelName, util.SeldonModelHeader, externalModelName)
		rp.logger.Error(msg)
		return "", "", status.Error(codes.FailedPrecondition, msg)
	}
}

func (rp *reverseGRPCProxy) createOutgoingCtxWithRequestId(ctx context.Context) (context.Context, string) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	var requestId string
	requestIds := md.Get(util.RequestIdHeader)
	rp.logger.Debugf("Request ids from incoming meta %s", requestIds)
	if len(requestIds) == 0 {
		requestId = util.CreateRequestId()
	} else {
		requestId = requestIds[0]
	}
	ctxNew := metadata.NewOutgoingContext(ctx, md)
	return metadata.AppendToOutgoingContext(ctxNew, util.RequestIdHeader, requestId), requestId
}

func (rp *reverseGRPCProxy) setTrailer(ctx context.Context, trailer metadata.MD, requestId string) {
	logger := rp.logger.WithField("func", "SetTrailer")
	if trailer == nil {
		trailer = metadata.MD{}
	}
	trailer.Set(util.RequestIdHeader, requestId)
	errTrailer := grpc.SetTrailer(ctx, trailer) // pass on any trailers set by inference server such as MLServer
	if errTrailer != nil {
		logger.WithError(errTrailer).Error("Failed to set trailers")
	}
}

// to sync between scalingMetricsSetup and scalingMetricsTearDown calls running in go routines
func (rp *reverseGRPCProxy) syncScalingMetrics(internalModelName string) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		if err := rp.modelScalingStatsCollector.ScalingMetricsSetup(&wg, internalModelName); err != nil {
			rp.logger.WithError(err).Warnf("cannot collect scaling stats for model %s", internalModelName)
		}
	}()
	defer func() {
		go func() {
			if err := rp.modelScalingStatsCollector.ScalingMetricsTearDown(&wg, internalModelName); err != nil {
				rp.logger.WithError(err).Warnf("cannot collect scaling stats for model %s", internalModelName)
			}
		}()
	}()

}

func (rp *reverseGRPCProxy) ModelInfer(ctx context.Context, r *v2.ModelInferRequest) (*v2.ModelInferResponse, error) {
	logger := rp.logger.WithField("func", "ModelInfer")
	internalModelName, externalModelName, err := rp.extractModelNamesFromContext(ctx)
	if err != nil {
		return nil, err
	}
	r.ModelName = internalModelName
	r.ModelVersion = ""

	// handle scaling metrics
	rp.syncScalingMetrics(internalModelName)

	startTime := time.Now()
	err = rp.ensureLoadModel(r.ModelName)
	if err != nil {
		elapsedTime := time.Since(startTime).Seconds()
		go rp.metrics.AddModelInferMetrics(externalModelName, internalModelName, metrics.MethodTypeGrpc, elapsedTime, codes.NotFound.String())
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Model %s not found (err: %s)", r.ModelName, err))
	}

	// Create an outgoing context for the proxy call to service from incoming context
	outgoingCtx, requestId := rp.createOutgoingCtxWithRequestId(ctx)

	var trailer metadata.MD
	opts := append(rp.callOptions, grpc.Trailer(&trailer))
	resp, err := rp.getV2GRPCClient().ModelInfer(outgoingCtx, r, opts...)
	if retryForLazyReload(err) {
		if v2Err := rp.stateManager.v2Client.LoadModel(internalModelName); v2Err != nil {
			logger.WithError(v2Err).Warnf("error loading model %s", internalModelName)
		}
		resp, err = rp.getV2GRPCClient().ModelInfer(outgoingCtx, r, opts...)
	}

	rp.setTrailer(ctx, trailer, requestId)

	grpcStatus, _ := status.FromError(err)
	elapsedTime := time.Since(startTime).Seconds()
	go rp.metrics.AddModelInferMetrics(externalModelName, internalModelName, metrics.MethodTypeGrpc, elapsedTime, grpcStatus.Code().String())
	return resp, err
}

type InferMessage interface {
	*v2.ModelInferRequest | *v2.ModelInferResponse
}

// forwardStream forwards messages from recv to send, modifying them with modify
// before sending. Two instances of forwardStream are launched, one for each
// direction of the stream, from client <> reverse proxy and  reverse proxy <> (model) server. If one
// of the streams stops due to an error, it cancels the context and returns.
// This will stop the other stream as well. The error is sent to the errChan
// and it is returned in the ModelStreamInfer. Note that EOF is not considered
// an error, but it is just a signal that the stream is done. The CloseSend
// function is called on either error or EOF to properly signal the end of the
// stream so the resources can be properly released.
//
// A goroutine is launched to wait for the termination of the two streams. If
// no error occurs, the goroutine closes the errChan, releasing the ModelStreamInfer
// with a nil error. If an error occurs, the goroutine sends the error to the
// errChan. In ModelStreamInfer, only the first error is waited for and returned.
// Note that we need to launch a goroutine to wait for the termination of the
// two streams, because otherwise the ModelStreamInfer function might wait indefinitely
// (e.g., a forwardStream might be blocked on a recv operation and never call wg.Done()).
func forwardStream[T InferMessage](
	ctx context.Context,
	recv func() (T, error),
	send func(T) error,
	modify func(T) T,
	errChan chan<- error,
	wg *sync.WaitGroup,
	cancel context.CancelFunc,
	closeSend func() error,
) {
	defer func() {
		if err := closeSend(); err != nil {
			errChan <- fmt.Errorf("failed to close send stream: %v", err)
			cancel()
		}
	}()
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				errChan <- fmt.Errorf("stream recv error: %v", err)
				cancel()
				return
			}

			msg = modify(msg)
			if err := send(msg); err != nil {
				errChan <- fmt.Errorf("stream send error: %v", err)
				cancel()
				return
			}
		}
	}
}

func (rp *reverseGRPCProxy) ModelStreamInfer(stream v2.GRPCInferenceService_ModelStreamInferServer) error {
	ctx := stream.Context()
	logger := rp.logger.WithField("func", "ModelStreamInfer")
	internalModelName, externalModelName, err := rp.extractModelNamesFromContext(ctx)
	if err != nil {
		return err
	}

	// handle scaling metrics
	rp.syncScalingMetrics(internalModelName)

	startTime := time.Now()
	// TODO: check the model is still loaded while the stream is going, not just at the start of the stream
	err = rp.ensureLoadModel(internalModelName)
	if err != nil {
		elapsedTime := time.Since(startTime).Seconds()
		go rp.metrics.AddModelInferMetrics(externalModelName, internalModelName, metrics.MethodTypeGrpc, elapsedTime, codes.NotFound.String())
		return status.Error(codes.NotFound, fmt.Sprintf("Model %s not found (err: %s)", internalModelName, err))
	}

	// Create an outgoing context for the proxy call to service from incoming context
	outgoingCtx, requestId := rp.createOutgoingCtxWithRequestId(ctx)

	var trailer metadata.MD
	opts := append(rp.callOptions, grpc.Trailer(&trailer), grpc_retry.Disable())
	clientStream, err := rp.getV2GRPCClient().ModelStreamInfer(outgoingCtx, opts...)
	if err != nil {
		logger.WithError(err).Error("Failed to create stream")
		return err
	}

	errChan := make(chan error, 4) // Buffered channel to collect errors

	// Add cancel to context to cancel goroutines when error occurs
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	wg := sync.WaitGroup{}

	// Launch goroutines for forwarding
	wg.Add(2)
	go forwardStream(
		ctxWithCancel,
		stream.Recv,
		clientStream.Send,
		func(req *v2.ModelInferRequest) *v2.ModelInferRequest {
			req.ModelName = internalModelName
			req.ModelVersion = ""
			return req
		},
		errChan,
		&wg,
		cancel,
		clientStream.CloseSend, // Pass CloseSend function here
	)
	go forwardStream(
		ctxWithCancel,
		clientStream.Recv,
		stream.Send,
		func(resp *v2.ModelInferResponse) *v2.ModelInferResponse { return resp }, // No modification
		errChan,
		&wg,
		cancel,
		func() error { return nil }, // No need to close stream on receiving
	)

	go func() {
		wg.Wait()
		close(errChan) // Close the error channel after all goroutines are done
	}()

	rp.setTrailer(ctx, trailer, requestId)

	err = <-errChan // Wait for the first error
	if err != nil {
		logger.WithError(err).Error("Error in stream forwarding")
	}

	grpcStatus, _ := status.FromError(err)
	elapsedTime := time.Since(startTime).Seconds()
	go rp.metrics.AddModelInferMetrics(externalModelName, internalModelName, metrics.MethodTypeGrpc, elapsedTime, grpcStatus.Code().String())
	return err
}

func (rp *reverseGRPCProxy) ModelMetadata(ctx context.Context, r *v2.ModelMetadataRequest) (*v2.ModelMetadataResponse, error) {
	internalModelName, _, err := rp.extractModelNamesFromContext(ctx)
	if err != nil {
		return nil, err
	}
	r.Name = internalModelName
	r.Version = ""

	if err := rp.ensureLoadModel(r.Name); err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Model %s not found (err: %s)", r.Name, err))
	}

	resp, err := rp.getV2GRPCClient().ModelMetadata(ctx, r)
	if retryForLazyReload(err) {
		if v2Err := rp.stateManager.v2Client.LoadModel(internalModelName); v2Err != nil {
			rp.logger.WithError(v2Err).Warnf("error loading model %s", internalModelName)
		}
		resp, err = rp.getV2GRPCClient().ModelMetadata(ctx, r)
	}
	return resp, err
}

func (rp *reverseGRPCProxy) ModelReady(ctx context.Context, r *v2.ModelReadyRequest) (*v2.ModelReadyResponse, error) {
	internalModelName, _, err := rp.extractModelNamesFromContext(ctx)
	if err != nil {
		return nil, err
	}
	r.Name = internalModelName
	r.Version = ""

	if err := rp.ensureLoadModel(r.Name); err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Model %s not found (err: %s)", r.Name, err))
	}

	resp, err := rp.getV2GRPCClient().ModelReady(ctx, r)
	if retryForLazyReload(err) {
		if v2Err := rp.stateManager.v2Client.LoadModel(internalModelName); v2Err != nil {
			rp.logger.WithError(v2Err).Warnf("error loading model %s", internalModelName)
		}
		resp, err = rp.getV2GRPCClient().ModelReady(ctx, r)
	}
	return resp, err
}

func (rp *reverseGRPCProxy) ensureLoadModel(modelId string) error {
	return rp.stateManager.EnsureLoadModel(modelId)
}

func (rp *reverseGRPCProxy) getV2GRPCClient() v2.GRPCInferenceServiceClient {
	i := rand.Intn(len(rp.v2GRPCClientPool))
	return rp.v2GRPCClientPool[i]
}

func (rp *reverseGRPCProxy) createV2CRPCClients(backendGRPCServerHost string, backendGRPCServerPort int, size int) ([]*grpc.ClientConn, []v2.GRPCInferenceServiceClient, error) {
	conns := make([]*grpc.ClientConn, size)
	clients := make([]v2.GRPCInferenceServiceClient, size)
	for i := 0; i < size; i++ {
		conn, err := oip.CreateV2GrpcConnection(
			oip.GetV2ConfigWithDefaults(backendGRPCServerHost, backendGRPCServerPort))

		if err != nil {
			// TODO: this could fail in later iterations, so close earlier connections
			conn.Close()
			return nil, nil, err
		}

		conns[i] = conn
		clients[i] = v2.NewGRPCInferenceServiceClient(conn)
	}
	return conns, clients, nil
}

func closeV2CRPCConnections(conns []*grpc.ClientConn) {
	for i := 0; i < len(conns); i++ {
		// TODO: handle errors in closing connections?
		_ = conns[i].Close()
	}
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

func extractModelNamesFromHeaders(ctx context.Context) (string, string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		internalModelName := extractHeader(util.SeldonInternalModelHeader, md)
		externalModelName, _, err := util.GetOrignalModelNameAndVersion(internalModelName)
		if err != nil {
			externalModelName = extractHeader(util.SeldonModelHeader, md)
		}
		return internalModelName, externalModelName, internalModelName != "" && externalModelName != ""
	}
	return "", "", false
}

func getGrpcErrCode(err error) codes.Code {
	if err != nil {
		if e, ok := status.FromError(err); ok {
			return e.Code()
		}
	}
	return codes.Unknown
}

func retryForLazyReload(err error) bool {
	// we do lazy load in case of 404 and unavailable (triton), the idea being that if ml server restarts, state with agent is inconsistent.
	return getGrpcErrCode(err) == codes.NotFound || getGrpcErrCode(err) == codes.Unavailable
}
