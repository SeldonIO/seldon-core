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

package proxy

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	pba "github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
)

type AgentServer struct {
	pba.UnimplementedAgentServiceServer
	logger logrus.FieldLogger
	lock   sync.RWMutex
	agents map[ServerKey]*AgentSubscriber
}

type ServerKey struct {
	serverName string
	replicaIdx uint32
}

type AgentSubscriber struct {
	finished chan<- bool
	stream   pba.AgentService_SubscribeServer
}

func NewAgentServer(l logrus.FieldLogger) *AgentServer {
	return &AgentServer{
		logger: l,
		lock:   sync.RWMutex{},
		agents: make(map[ServerKey]*AgentSubscriber),
	}
}

func (a *AgentServer) Start(agentPort uint) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", agentPort))
	if err != nil {
		log.Fatalf("failed to create listener: %v", err)
	}

	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)
	pba.RegisterAgentServiceServer(grpcServer, a)
	a.logger.Printf("Agent server running on %d", agentPort)

	return grpcServer.Serve(lis)
}

func (a *AgentServer) AgentEvent(ctx context.Context, e *pba.ModelEventMessage) (*pba.ModelEventResponse, error) {
	l := a.logger.WithField("func", "AgentEvent")
	l.Debugf("received event %s from agent %s:%d", e.Event.String(), e.ServerName, e.ReplicaIdx)

	return &pba.ModelEventResponse{}, nil
}

func (a *AgentServer) Subscribe(request *pba.AgentSubscribeRequest, stream pba.AgentService_SubscribeServer) error {
	logger := a.logger.WithField("func", "Subscribe")
	logger.Infof("Received subscribe request from %s:%d", request.ServerName, request.ReplicaIdx)

	fin := make(chan bool)

	a.lock.Lock()
	a.agents[ServerKey{serverName: request.ServerName, replicaIdx: request.ReplicaIdx}] = &AgentSubscriber{
		finished: fin,
		stream:   stream,
	}
	a.lock.Unlock()

	ctx := stream.Context()
	// Keep this scope alive because once this scope exits - the stream is closed
	for {
		select {
		case <-fin:
			logger.Infof("Closing stream for replica: %s:%d", request.ServerName, request.ReplicaIdx)
			return nil
		case <-ctx.Done():
			logger.Infof("Client replica %s:%d has disconnected", request.ServerName, request.ReplicaIdx)

			a.lock.Lock()
			delete(a.agents, ServerKey{serverName: request.ServerName, replicaIdx: request.ReplicaIdx})
			a.lock.Unlock()

			return nil
		}
	}
}

func (a *AgentServer) HandleModelEvent(event ModelEvent) error {
	a.lock.RLock()
	defer a.lock.RUnlock()

	var lastError error
	for _, subscriber := range a.agents {
		err := subscriber.stream.Send(event.ModelOperationMessage)
		if err != nil {
			lastError = err
		}
	}

	return lastError
}
