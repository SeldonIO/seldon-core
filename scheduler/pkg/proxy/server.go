package proxy

import (
	"context"
	"fmt"
	"net"
	"sync"

	pba "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/proxy"
	pbs "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
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
			Operation:    pba.ModelOperationMessage_LOAD_MODEL,
			ModelVersion: &pba.ModelVersion{Model: r.GetRequest().GetModel(), Version: r.GetVersion()},
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
