/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package agent

// grpc endpoints for testing and debugging purposes
// they should not be used in production settings

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbad "github.com/seldonio/seldon-core/apis/go/v2/mlops/agent_debug"
)

const (
	GRPCDebugServicePort = 7777
)

type agentDebug struct {
	pbad.UnimplementedAgentDebugServiceServer
	logger       log.FieldLogger
	stateManager *LocalStateManager
	grpcServer   *grpc.Server
	serverReady  bool
	port         uint
	mu           sync.RWMutex
}

func NewAgentDebug(logger log.FieldLogger, port uint) *agentDebug {
	return &agentDebug{
		logger: logger.WithField("source", "AgentDebug"),
		port:   port,
	}
}

func (cd *agentDebug) SetState(sm interface{}) {
	cd.stateManager = sm.(*LocalStateManager)
}

func (cd *agentDebug) Start() error {
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
		cd.logger.WithError(err).Info("Client debug service stopped")
		cd.mu.Lock()
		cd.serverReady = false
		cd.mu.Unlock()
	}()
	return nil
}

func (cd *agentDebug) Stop() error {
	cd.logger.Info("Start graceful shutdown")
	cd.mu.Lock()
	defer cd.mu.Unlock()
	if cd.grpcServer != nil {
		cd.grpcServer.GracefulStop()
	}
	cd.serverReady = false
	cd.logger.Info("Finished graceful shutdown")
	return nil
}

func (cd *agentDebug) Ready() bool {
	cd.mu.RLock()
	defer cd.mu.RUnlock()
	return cd.serverReady
}

func (rp *agentDebug) Name() string {
	return "AgentDebug GRPC service"
}

func (cd *agentDebug) ReplicaStatus(ctx context.Context, r *pbad.ReplicaStatusRequest) (*pbad.ReplicaStatusResponse, error) {
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
