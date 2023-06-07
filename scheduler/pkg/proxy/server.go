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
	"net"
	"sync"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	pba "github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/proxy"
	pbs "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent"
)

type ProxyServer struct {
	pb.UnimplementedSchedulerProxyServer
	logger      log.FieldLogger
	lock        sync.Mutex
	modelEvents chan<- ModelEvent
}

func NewProxyServer(logger log.FieldLogger, es chan<- ModelEvent) *ProxyServer {
	return &ProxyServer{
		logger:      logger,
		lock:        sync.Mutex{},
		modelEvents: es,
	}
}

func (p *ProxyServer) Start(port uint) error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		p.logger.Errorf("unable to start gRPC listening server on port %d", port)
		return err
	}

	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterSchedulerProxyServer(grpcServer, p)

	p.logger.Infof("starting gRPC listening server on port %d", port)
	return grpcServer.Serve(l)
}

func (p *ProxyServer) LoadModel(ctx context.Context, r *pb.LoadModelRequest) (*pb.LoadModelResponse, error) {
	m := ModelEvent{
		ModelOperationMessage: &pba.ModelOperationMessage{
			Operation:          pba.ModelOperationMessage_LOAD_MODEL,
			ModelVersion:       &pba.ModelVersion{Model: r.GetRequest().GetModel(), Version: r.GetVersion()},
			AutoscalingEnabled: agent.AutoscalingEnabled(r.GetRequest().GetModel()),
		},
	}
	p.modelEvents <- m

	return &pb.LoadModelResponse{}, nil
}

func (p *ProxyServer) UnloadModel(ctx context.Context, r *pb.UnloadModelRequest) (*pb.UnloadModelResponse, error) {
	m := ModelEvent{
		ModelOperationMessage: &pba.ModelOperationMessage{
			Operation: pba.ModelOperationMessage_UNLOAD_MODEL,
			ModelVersion: &pba.ModelVersion{
				Model:   &pbs.Model{Meta: &pbs.MetaData{Name: r.GetModel().GetName()}},
				Version: r.GetVersion()},
		},
	}
	p.modelEvents <- m
	return &pb.UnloadModelResponse{}, nil
}
