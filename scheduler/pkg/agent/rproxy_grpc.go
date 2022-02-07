package agent

// reverse proxy for grpc infer and metadata endpoints

import (
	"context"
	"fmt"
	"net"
	"sync"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	v2 "github.com/seldonio/seldon-core/scheduler/apis/mlops/v2_dataplane"
)

const (
	ReverseGRPCProxyPort          = 9998
	ModelNameKeyInHeader          = "seldon-model"
	grpcProxyMaxConcurrentStreams = 1_000_000
)

type reverseGRPCProxy struct {
	v2.UnimplementedGRPCInferenceServiceServer
	logger                log.FieldLogger
	stateManager          *LocalStateManager
	grpcServer            *grpc.Server
	serverReady           bool
	backendGRPCServerHost string
	backendGRPCServerPort uint
	v2GRPCClient          v2.GRPCInferenceServiceClient
	port                  uint // service port
	mu                    sync.RWMutex
}

func NewReverseGRPCProxy(logger log.FieldLogger, backendGRPCServerHost string, backendGRPCServerPort uint, servicePort uint) *reverseGRPCProxy {
	return &reverseGRPCProxy{
		logger:                logger,
		backendGRPCServerHost: backendGRPCServerHost,
		backendGRPCServerPort: backendGRPCServerPort,
		port:                  servicePort,
	}
}

func (rp *reverseGRPCProxy) SetState(sm *LocalStateManager) {
	rp.stateManager = sm
}

func (rp *reverseGRPCProxy) Start() error {
	if rp.stateManager == nil {
		rp.logger.Error("Set state before starting reverse proxy service")
		return fmt.Errorf("State not set, aborting")
	}

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", rp.port))
	if err != nil {
		rp.logger.Errorf("unable to start gRPC listening server on port %d", rp.port)
		return err
	}

	opts := []grpc.ServerOption{}
	opts = append(opts, grpc.MaxConcurrentStreams(grpcProxyMaxConcurrentStreams))
	grpcServer := grpc.NewServer(opts...)
	v2.RegisterGRPCInferenceServiceServer(grpcServer, rp)

	// TODO: add this to V2Client
	rp.logger.Infof("Setting grpc v2 client on port %d", rp.backendGRPCServerPort)

	conn, err := getConnection(rp.backendGRPCServerHost, int(rp.backendGRPCServerPort))

	if err != nil {
		rp.logger.Error("Cannot dial to backend server (%s)", err)
		conn.Close()
		return err
	}
	rp.v2GRPCClient = v2.NewGRPCInferenceServiceClient(conn)

	rp.logger.Infof("Starting gRPC listening server on port %d", rp.port)
	rp.grpcServer = grpcServer
	go func() {
		defer conn.Close()
		rp.mu.Lock()
		rp.serverReady = true
		rp.mu.Unlock()
		err := rp.grpcServer.Serve(l)
		rp.logger.Infof("Reverse GRPC proxy stopped with error: %s", err)
		rp.mu.Lock()
		rp.serverReady = false
		rp.mu.Unlock()
	}()
	return nil
}

func (rp *reverseGRPCProxy) Stop() error {
	// Shutdown is graceful
	rp.mu.Lock()
	defer rp.mu.Unlock()
	rp.grpcServer.GracefulStop()
	rp.serverReady = false
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

func (rp *reverseGRPCProxy) ModelInfer(ctx context.Context, r *v2.ModelInferRequest) (*v2.ModelInferResponse, error) {

	if modelId, inHeader := extractModelNameFromHeader(ctx); inHeader {
		r.ModelName = modelId
		rp.logger.Debugf("Model name set from header: %s", r.ModelName)
	}

	if err := rp.ensureLoadModel(r.ModelName); err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Model %s not found (err: %s)", r.ModelName, err))
	}

	return rp.v2GRPCClient.ModelInfer(ctx, r)
}

func (rp *reverseGRPCProxy) ModelMetadata(ctx context.Context, r *v2.ModelMetadataRequest) (*v2.ModelMetadataResponse, error) {

	if modelId, inHeader := extractModelNameFromHeader(ctx); inHeader {
		r.Name = modelId
		rp.logger.Debugf("Model name set from header: %s", r.Name)
	}

	if err := rp.ensureLoadModel(r.Name); err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Model %s not found (err: %s)", r.Name, err))
	}

	return rp.v2GRPCClient.ModelMetadata(ctx, r)
}

func (rp *reverseGRPCProxy) ModelReady(ctx context.Context, r *v2.ModelReadyRequest) (*v2.ModelReadyResponse, error) {

	if modelId, inHeader := extractModelNameFromHeader(ctx); inHeader {
		r.Name = modelId
		rp.logger.Debugf("Model name set from header: %s", r.Name)
	}

	if err := rp.ensureLoadModel(r.Name); err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Model %s not found (err: %s)", r.Name, err))
	}

	return rp.v2GRPCClient.ModelReady(ctx, r)
}

func (rp *reverseGRPCProxy) ensureLoadModel(modelId string) error {
	return rp.stateManager.EnsureLoadModel(modelId)
}

func extractModelNameFromHeader(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		modelNameFromHeader, ok := md[ModelNameKeyInHeader]
		if ok {
			if len(modelNameFromHeader) > 0 {
				// note if there are more than one elements we just return the first one
				return modelNameFromHeader[0], true
			}
		}
	}
	return "", false
}
