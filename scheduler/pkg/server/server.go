package server

import (
	"context"
	"fmt"
	"net"

	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/agent"
	scheduler2 "github.com/seldonio/seldon-core/scheduler/pkg/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrAddServerEmptyServerName = status.Errorf(codes.FailedPrecondition, "Empty server name passed")
)

type SchedulerServer struct {
	pb.UnimplementedSchedulerServer
	logger       log.FieldLogger
	store        store.SchedulerStore
	scheduler    scheduler2.Scheduler
	agentHandler agent.AgentHandler
}

func (s SchedulerServer) SubscribeModelEvents(req *pb.ModelSubscriptionRequest, server pb.Scheduler_SubscribeModelEventsServer) error {
	panic("implement me")
}

func (s SchedulerServer) StartGrpcServer(schedulerPort uint) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", schedulerPort))
	if err != nil {
		log.Fatalf("failed to create listener: %v", err)
	}
	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterSchedulerServer(grpcServer, s)
	s.logger.Printf("Scheduler server running on %d", schedulerPort)
	return grpcServer.Serve(lis)
}

func NewSchedulerServer(logger log.FieldLogger, store store.SchedulerStore, scheduler scheduler2.Scheduler, agentHandler agent.AgentHandler) *SchedulerServer {

	s := &SchedulerServer{
		logger:       logger,
		store:        store,
		scheduler:    scheduler,
		agentHandler: agentHandler,
	}
	return s
}

func (s SchedulerServer) LoadModel(ctx context.Context, req *pb.LoadModelRequest) (*pb.LoadModelResponse, error) {
	logger := s.logger.WithField("func", "LoadModel")
	logger.Debugf("Load model %s", req.GetModel().GetName())
	exists := s.store.ExistsModelVersion(req.GetModel().Name, req.GetModel().Version)
	if exists { //TODO check the model details match what we have otherwise error
		logger.Infof("Model %s:%s already exists",req.GetModel().Name, req.Model.Version)
		return &pb.LoadModelResponse{}, nil
	}
	err := s.store.UpdateModel(req.GetModel())
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	err = s.scheduler.Schedule(req.GetModel().GetName())
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	s.agentHandler.SendAgentSync(req.GetModel().GetName())
	return &pb.LoadModelResponse{}, nil
}

func (s SchedulerServer) UnloadModel(ctx context.Context, reference *pb.ModelReference) (*pb.UnloadModelResponse, error) {
	logger := s.logger.WithField("func", "UnloadModel")
	logger.Debugf("Unload model %s", reference.Name)
	err := s.store.RemoveModel(reference.Name)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	err = s.scheduler.Schedule(reference.Name)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	s.agentHandler.SendAgentSync(reference.GetName())
	return &pb.UnloadModelResponse{}, nil
}

func (s SchedulerServer) ModelStatus(ctx context.Context, reference *pb.ModelReference) (*pb.ModelStatusResponse, error) {
	model, err := s.store.GetModel(reference.Name)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	if model == nil {
		return nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("Failed to find model %s", reference.Name))
	}
	latestModel := model.GetLatest()
	if latestModel == nil {
		return nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("Failed to find model %s", reference.Name))
	}
	return &pb.ModelStatusResponse{
		ModelName:         reference.Name,
		Version:           latestModel.GetVersion(),
		ServerName:        latestModel.Server(),
		ModelReplicaState: latestModel.ReplicaState(),
	}, nil
}

func (s SchedulerServer) ServerStatus(ctx context.Context, reference *pb.ServerReference) (*pb.ServerStatusResponse, error) {
	server, err := s.store.GetServer(reference.Name)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	if server == nil {
		return nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("Failed to find server %s", reference.Name))
	}
	ss := &pb.ServerStatusResponse{
		ServerName: reference.Name,
	}

	for _, replica := range server.Replicas {
		ss.Resources = append(ss.Resources, &pb.ServerResources{
			Memory:               replica.GetMemory(),
			AvailableMemoryBytes: replica.GetAvailableMemory(),
		})
	}
	return ss, nil
}
