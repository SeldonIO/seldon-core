package server

import (
	"context"
	"fmt"
	"net"
	"sync"

	"google.golang.org/protobuf/types/known/timestamppb"

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
	mutext       sync.RWMutex
	EventStream
}

type EventStream struct {
	streams   map[pb.Scheduler_SubscribeModelStatusServer]*Subscription
	chanEvent chan *store.ModelSnapshot
}

type Subscription struct {
	name   string
	stream pb.Scheduler_SubscribeModelStatusServer
	fin    chan bool
}

func (s *SchedulerServer) StartGrpcServer(schedulerPort uint) error {
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

func NewSchedulerServer(logger log.FieldLogger, schedStore store.SchedulerStore, scheduler scheduler2.Scheduler, agentHandler agent.AgentHandler) *SchedulerServer {

	s := &SchedulerServer{
		logger:       logger.WithField("source", "SchedulerServer"),
		store:        schedStore,
		scheduler:    scheduler,
		agentHandler: agentHandler,
		EventStream: EventStream{
			streams:   make(map[pb.Scheduler_SubscribeModelStatusServer]*Subscription),
			chanEvent: make(chan *store.ModelSnapshot, 1),
		},
	}
	schedStore.AddListener(s.EventStream.chanEvent)
	return s
}

func (s *SchedulerServer) LoadModel(ctx context.Context, req *pb.LoadModelRequest) (*pb.LoadModelResponse, error) {
	logger := s.logger.WithField("func", "LoadModel")
	logger.Debugf("Load model %+v k8s meta %+v", req.GetModel().GetMeta(), req.GetModel().GetMeta().GetKubernetesMeta())
	s.store.UpdateModel(req)
	err := s.scheduler.Schedule(req.GetModel().GetMeta().GetName())
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	s.agentHandler.SendAgentSync(req.GetModel().GetMeta().GetName())
	return &pb.LoadModelResponse{}, nil
}

func (s *SchedulerServer) UnloadModel(ctx context.Context, req *pb.UnloadModelRequest) (*pb.UnloadModelResponse, error) {
	logger := s.logger.WithField("func", "UnloadModel")
	logger.Debugf("Unload model %s", req.GetModel().Name)
	err := s.store.RemoveModel(req)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	err = s.scheduler.Schedule(req.GetModel().Name)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	s.agentHandler.SendAgentSync(req.GetModel().GetName())
	return &pb.UnloadModelResponse{}, nil
}

func createModelVersionStatus(mv *store.ModelVersion) *pb.ModelVersionStatus {
	stateMap := make(map[int32]*pb.ModelReplicaStatus)
	for k, v := range mv.ReplicaState() {
		stateMap[int32(k)] = &pb.ModelReplicaStatus{
			State:               pb.ModelReplicaStatus_ModelReplicaState(pb.ModelReplicaStatus_ModelReplicaState_value[v.State.String()]),
			Reason:              v.Reason,
			LastChangeTimestamp: timestamppb.New(v.Timestamp),
		}
	}
	modelState := mv.ModelState()
	mvs := &pb.ModelVersionStatus{
		Version:           mv.GetVersion(),
		ServerName:        mv.Server(),
		ModelReplicaState: stateMap,
		State: &pb.ModelStatus{
			State:               pb.ModelStatus_ModelState(pb.ModelStatus_ModelState_value[modelState.State.String()]),
			Reason:              modelState.Reason,
			LastChangeTimestamp: timestamppb.New(modelState.Timestamp),
			AvailableReplicas:   modelState.AvailableReplicas,
			UnavailableReplicas: modelState.UnavailableReplicas,
		},
	}
	if mv.GetMeta().KubernetesMeta != nil {
		mvs.KubernetesMeta = mv.GetModel().GetMeta().GetKubernetesMeta()
	}
	return mvs
}

func (s *SchedulerServer) modelStatusImpl(ctx context.Context, model *store.ModelSnapshot, allVersions bool) (*pb.ModelStatusResponse, error) {
	var modelVersionStatuses []*pb.ModelVersionStatus
	if !allVersions {
		latestModel := model.GetLatest()
		if latestModel == nil {
			return nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("Failed to find model %s", model.Name))
		}
		modelVersionStatuses = append(modelVersionStatuses, createModelVersionStatus(latestModel))
	} else {
		for _, mv := range model.Versions {
			modelVersionStatuses = append(modelVersionStatuses, createModelVersionStatus(mv))
		}
	}
	msr := &pb.ModelStatusResponse{
		ModelName: model.Name,
		Versions:  modelVersionStatuses,
	}
	return msr, nil
}

func (s *SchedulerServer) ModelStatus(ctx context.Context, req *pb.ModelStatusRequest) (*pb.ModelStatusResponse, error) {
	model, err := s.store.GetModel(req.Model.Name)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	if model == nil || len(model.Versions) == 0 {
		return nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("Failed to find model %s", req.Model.Name))
	}
	return s.modelStatusImpl(ctx, model, req.AllVersions)
}

func (s *SchedulerServer) ServerStatus(ctx context.Context, reference *pb.ServerReference) (*pb.ServerStatusResponse, error) {
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
			ReplicaIdx:           uint32(replica.GetReplicaIdx()),
			Memory:               replica.GetMemory(),
			AvailableMemoryBytes: replica.GetAvailableMemory(),
		})
	}
	return ss, nil
}
