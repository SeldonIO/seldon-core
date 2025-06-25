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

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
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

func (ad *agentDebug) SetState(sm interface{}) {
	ad.stateManager = sm.(*LocalStateManager)
}

func (ad *agentDebug) Start() error {
	if ad.stateManager == nil {
		return fmt.Errorf("set state before starting the debug service")
	}
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", ad.port))
	if err != nil {
		ad.logger.Errorf("Unable to start gRPC listening server on port %d", ad.port)
		return err
	}

	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)
	pbad.RegisterAgentDebugServiceServer(grpcServer, ad)

	ad.logger.Infof("Starting gRPC listening server on port %d", ad.port)
	ad.grpcServer = grpcServer
	go func() {
		ad.mu.Lock()
		ad.serverReady = true
		ad.mu.Unlock()
		err := ad.grpcServer.Serve(l)
		ad.logger.WithError(err).Info("Client debug service stopped")
		ad.mu.Lock()
		ad.serverReady = false
		ad.mu.Unlock()
	}()
	return nil
}

func (ad *agentDebug) Stop() error {
	ad.logger.Info("Start graceful shutdown")
	ad.mu.Lock()
	defer ad.mu.Unlock()
	if ad.grpcServer != nil {
		ad.grpcServer.GracefulStop()
	}
	ad.serverReady = false
	ad.logger.Info("Finished graceful shutdown")
	return nil
}

func (ad *agentDebug) Ready() bool {
	ad.mu.RLock()
	defer ad.mu.RUnlock()
	return ad.serverReady
}

func (ad *agentDebug) Name() string {
	return "AgentDebug GRPC service"
}

func (cd *agentDebug) GetType() interfaces.SubServiceType {
	return interfaces.OptionalService
}

func (ad *agentDebug) ReplicaStatus(ctx context.Context, r *pbad.ReplicaStatusRequest) (*pbad.ReplicaStatusResponse, error) {
	numModels := ad.stateManager.modelVersions.numModels()
	models := make([]*pbad.ModelReplicaState, numModels)
	i := 0
	// TODO: make read loadedModels thread safe
	for _, name := range ad.stateManager.modelVersions.modelNames() {
		ts, err := ad.stateManager.cache.Get(name)
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
		AvailableMemoryBytes: uint64(ad.stateManager.GetAvailableMemoryBytes()),
		Models:               models}, nil
}
