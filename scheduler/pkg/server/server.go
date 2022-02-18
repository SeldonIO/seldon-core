package server

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/seldonio/seldon-core/scheduler/pkg/coordinator"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	scheduler2 "github.com/seldonio/seldon-core/scheduler/pkg/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	grpcMaxConcurrentStreams = 1_000_000
)

var (
	ErrAddServerEmptyServerName = status.Errorf(codes.FailedPrecondition, "Empty server name passed")
)

type SchedulerServer struct {
	pb.UnimplementedSchedulerServer
	logger            log.FieldLogger
	store             store.SchedulerStore
	scheduler         scheduler2.Scheduler
	mutext            sync.RWMutex
	modelEventStream  ModelEventStream
	serverEventStream ServerEventStream
}

type ModelEventStream struct {
	streams   map[pb.Scheduler_SubscribeModelStatusServer]*ModelSubscription
	chanEvent chan coordinator.ModelEventMsg
}

type ServerEventStream struct {
	streams   map[pb.Scheduler_SubscribeServerStatusServer]*ServerSubscription
	chanEvent chan coordinator.ModelEventMsg
}

type ModelSubscription struct {
	name   string
	stream pb.Scheduler_SubscribeModelStatusServer
	fin    chan bool
}

type ServerSubscription struct {
	name   string
	stream pb.Scheduler_SubscribeServerStatusServer
	fin    chan bool
}

func (s *SchedulerServer) StartGrpcServer(schedulerPort uint) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", schedulerPort))
	if err != nil {
		log.Fatalf("failed to create listener: %v", err)
	}
	opts := []grpc.ServerOption{}
	opts = append(opts, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterSchedulerServer(grpcServer, s)
	s.logger.Printf("Scheduler server running on %d", schedulerPort)
	return grpcServer.Serve(lis)
}

func NewSchedulerServer(logger log.FieldLogger, schedStore store.SchedulerStore, scheduler scheduler2.Scheduler, eventHub *coordinator.ModelEventHub) *SchedulerServer {

	s := &SchedulerServer{
		logger:    logger.WithField("source", "SchedulerServer"),
		store:     schedStore,
		scheduler: scheduler,
		modelEventStream: ModelEventStream{
			streams:   make(map[pb.Scheduler_SubscribeModelStatusServer]*ModelSubscription),
			chanEvent: make(chan coordinator.ModelEventMsg, 1),
		},
		serverEventStream: ServerEventStream{
			streams:   make(map[pb.Scheduler_SubscribeServerStatusServer]*ServerSubscription),
			chanEvent: make(chan coordinator.ModelEventMsg, 1),
		},
	}
	eventHub.AddListener(s.modelEventStream.chanEvent)
	eventHub.AddListener(s.serverEventStream.chanEvent)
	return s
}

func (s *SchedulerServer) ServerNotify(ctx context.Context, req *pb.ServerNotifyRequest) (*pb.ServerNotifyResponse, error) {
	logger := s.logger.WithField("func", "ServerNotify")
	logger.Infof("Server notification %s expectedReplicas %d shared %v", req.GetName(), req.GetExpectedReplicas(), req.GetShared())
	err := s.store.ServerNotify(req)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	if req.ExpectedReplicas == 0 {
		go s.rescheduleModels(req.GetName())
	}
	return &pb.ServerNotifyResponse{}, nil
}

func (s *SchedulerServer) rescheduleModels(serverKey string) {
	logger := s.logger.WithField("func", "rescheduleModels")
	server, err := s.store.GetServer(serverKey)
	if err != nil {
		logger.WithError(err).Errorf("Failed to get server %s", serverKey)
		return
	}
	models := make(map[string]bool)
	for _, replica := range server.Replicas {
		for _, model := range replica.GetLoadedModelVersions() {
			models[model.Name] = true
		}
	}
	for model := range models {
		err := s.scheduler.Schedule(model)
		if err != nil {
			logger.WithError(err).Errorf("Failed to reschedule model %s for server %s", model, serverKey)
		}
	}
}

func (s *SchedulerServer) LoadModel(ctx context.Context, req *pb.LoadModelRequest) (*pb.LoadModelResponse, error) {
	logger := s.logger.WithField("func", "LoadModel")
	logger.Debugf("Load model %+v k8s meta %+v", req.GetModel().GetMeta(), req.GetModel().GetMeta().GetKubernetesMeta())
	err := s.store.UpdateModel(req)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	err = s.scheduler.Schedule(req.GetModel().GetMeta().GetName())
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}
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
		Deleted:   model.Deleted,
	}
	return msr, nil
}

func (s *SchedulerServer) ModelStatus(ctx context.Context, req *pb.ModelStatusRequest) (*pb.ModelStatusResponse, error) {
	s.store.LockModel(req.Model.Name)
	defer s.store.UnlockModel(req.Model.Name)

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
		ServerName:        reference.Name,
		AvailableReplicas: int32(len(server.Replicas)),
		ExpectedReplicas:  int32(server.ExpectedReplicas),
		KubernetesMeta:    server.KubernetesMeta,
	}

	var totalModels int32
	for _, replica := range server.Replicas {
		numLoadedModelsOnReplica := int32(replica.GetNumLoadedModels())
		ss.Resources = append(ss.Resources, &pb.ServerReplicaResources{
			ReplicaIdx:           uint32(replica.GetReplicaIdx()),
			TotalMemoryBytes:     replica.GetMemory(),
			AvailableMemoryBytes: replica.GetAvailableMemory(),
			NumLoadedModels:      numLoadedModelsOnReplica,
		})
		totalModels = totalModels + numLoadedModelsOnReplica
	}
	ss.NumLoadedModelReplicas = totalModels
	return ss, nil
}
