/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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

	pbad "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent_debug"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
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
