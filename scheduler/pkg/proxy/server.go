/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
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
