package agent

// grpc endpoints for testing and debugging purposes
// they should not be used in production settings

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	pbad "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent_debug"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	GRPCDebugServicePort = 7777
)

type ClientDebug struct {
	pbad.UnimplementedAgentDebugServiceServer
	logger       log.FieldLogger
	stateManager *LocalStateManager
	grpcServer   *grpc.Server
	serverReady  bool
	port         uint
	mu           sync.RWMutex
}

func NewClientDebug(logger log.FieldLogger, port uint) *ClientDebug {
	return &ClientDebug{
		logger: logger,
		port:   port,
	}
}

func (cd *ClientDebug) SetState(sm *LocalStateManager) {
	cd.stateManager = sm
}

func (cd *ClientDebug) Start() error {
	if cd.stateManager == nil {
		return fmt.Errorf("set state before starting the debug service")
	}
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", cd.port))
	if err != nil {
		cd.logger.Errorf("Unable to start gRPC listening server on port %d", cd.port)
		return err
	}

	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)
	pbad.RegisterAgentDebugServiceServer(grpcServer, cd)

	cd.logger.Infof("Starting gRPC listening server on port %d", cd.port)
	cd.grpcServer = grpcServer
	go func() {
		cd.mu.Lock()
		cd.serverReady = true
		cd.mu.Unlock()
		err := cd.grpcServer.Serve(l)
		cd.logger.Infof("Client debug service stopped (%s)", err)
		cd.mu.Lock()
		cd.serverReady = false
		cd.mu.Unlock()
	}()
	return nil
}

func (cd *ClientDebug) Stop() error {
	cd.mu.Lock()
	defer cd.mu.Unlock()
	cd.grpcServer.GracefulStop()
	cd.serverReady = false
	return nil
}

func (cd *ClientDebug) Ready() bool {
	cd.mu.RLock()
	defer cd.mu.RUnlock()
	return cd.serverReady
}

func (rp *ClientDebug) Name() string {
	return "ClientDebug GRPC service"
}

func (cd *ClientDebug) ReplicaStatus(ctx context.Context, r *pbad.ReplicaStatusRequest) (*pbad.ReplicaStatusResponse, error) {
	numModels := cd.stateManager.modelVersions.numModels()
	models := make([]*pbad.ModelReplicaState, numModels)
	i := 0
	// TODO: make read loadedModels thread safe
	for _, name := range cd.stateManager.modelVersions.modelNames() {
		ts, err := cd.stateManager.cache.Get(name)
		state := pbad.ModelReplicaState_Evicted
		tspb := timestamppb.New(time.Time{})
		if err == nil {
			state = pbad.ModelReplicaState_InMemory
			tspb = timestamppb.New(time.Unix(-ts/1000000000, 0)) // we store in cache negative unix ts
		}
		models[i] = &pbad.ModelReplicaState{
			State:        state,
			Name:         name,
			LastAccessed: tspb,
		}
		i++
	}
	return &pbad.ReplicaStatusResponse{
		AvailableMemoryBytes: uint64(cd.stateManager.GetAvailableMemoryBytes()),
		Models:               models}, nil
}
