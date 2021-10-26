package scheduler

import (
	"context"
	"errors"
	"fmt"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
)

var (
	ErrAddServerEmptyServerName = status.Errorf(codes.FailedPrecondition, "Empty server name passed")
)

type SchedulerServer struct {
	pb.UnimplementedSchedulerServer
	logger log.FieldLogger
	store       store.SchedulerStore
}

func (s SchedulerServer) SubscribeModelEvents(req *pb.ModelSubscriptionRequest, server pb.Scheduler_SubscribeModelEventsServer) error {
	panic("implement me")
}

func(s SchedulerServer) StartGrpcServer(schedulerPort uint) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", schedulerPort))
	if err != nil {
		log.Fatalf("failed to create listener: %v", err)
	}
	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterSchedulerServer(grpcServer, s)
	s.logger.Printf("Scheduler server running on %d", schedulerPort)
	return  grpcServer.Serve(lis)
}

func NewScheduler(logger log.FieldLogger, store store.SchedulerStore) *SchedulerServer {

	s := &SchedulerServer{
		logger:         logger,
		store:          store,
	}
	return s
}

func (s SchedulerServer) LoadModel(ctx context.Context, req *pb.LoadModelRequest) (*pb.LoadModelResponse, error) {
	// find modelAssignment assignment
	modelKey := req.Model.Name
	model, err := s.store.GetModel(modelKey)
	if err != nil {
		if errors.Is(err, store.ModelNotFoundErr) {
			err := s.store.CreateModel(modelKey, req.Model)
			if err != nil {
				return nil, status.Errorf(codes.FailedPrecondition, err.Error())
			}
			model, err = s.store.GetModel(modelKey)
			if err != nil {
				return nil, status.Errorf(codes.FailedPrecondition, err.Error())
			}
		} else {
			return nil, status.Errorf(codes.FailedPrecondition, err.Error())
		}
	}

	var server *store.Server
	if !model.HasServer() {
		if req.Model.Server != nil {
			server, err = s.store.GetServer(*req.Model.Server)
			if err != nil {
				return nil, status.Errorf(codes.FailedPrecondition, err.Error())
			}
		}
	} else {
		server, err = s.store.GetServer(model.Server())
		if err != nil {
			return nil, status.Errorf(codes.FailedPrecondition, err.Error())
		}
	}

	if server != nil {
		err := s.store.UpdateModelOnServer(model.Key(), server.Key())
		if err != nil {
			return nil, status.Errorf(codes.FailedPrecondition, err.Error())
		}
	} else {
		err := s.store.ScheduleModelToServer(modelKey)
		if err != nil {
			return nil, status.Errorf(codes.FailedPrecondition, err.Error())
		}
	}


	return &pb.LoadModelResponse{}, nil
}

func (s SchedulerServer) UnloadModel(ctx context.Context, reference *pb.ModelReference) (*pb.UnloadModelResponse, error) {
	err := s.store.RemoveModel(reference.Name)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	return &pb.UnloadModelResponse{}, nil
}

func (s SchedulerServer) ModelStatus(ctx context.Context, reference *pb.ModelReference) (*pb.ModelStatusResponse, error) {
	model, err := s.store.GetModel(reference.Name)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	return &pb.ModelStatusResponse{
		ModelName: reference.Name,
		ServerName: model.Server(),
		State: model.ReplicaState(),
	}, nil
}

func (s SchedulerServer) ServerStatus(ctx context.Context, reference *pb.ServerReference) (*pb.ServerStatusResponse, error) {
	server, err := s.store.GetServer(reference.Name)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	ss := &pb.ServerStatusResponse{
		ServerName: reference.Name,
	}

	for ridx := 0; ridx <  server.NumReplicas(); ridx++{
		ss.Resources = append(ss.Resources, &pb.ServerResources{
			Memory: server.GetMemory(ridx),
			AvailableMemory: server.GetAvailableMemory(ridx),
		})
	}
	return ss, nil
}
