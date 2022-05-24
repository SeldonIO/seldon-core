package server

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/seldonio/seldon-core/scheduler/pkg/store/pipeline"

	"github.com/seldonio/seldon-core/scheduler/pkg/store/experiment"

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
	grpcMaxConcurrentStreams       = 1_000_000
	pendingEventsQueueSize     int = 10
	modelEventHandlerName          = "scheduler.server.models"
	serverEventHandlerName         = "scheduler.server.servers"
	experimentEventHandlerName     = "scheduler.server.experiments"
	pipelineEventHandlerName       = "scheduler.server.pipelines"
)

var (
	ErrAddServerEmptyServerName = status.Errorf(codes.FailedPrecondition, "Empty server name passed")
)

type SchedulerServer struct {
	pb.UnimplementedSchedulerServer
	logger                log.FieldLogger
	modelStore            store.ModelStore
	experimentServer      experiment.ExperimentServer
	pipelineHandler       pipeline.PipelineHandler
	scheduler             scheduler2.Scheduler
	modelEventStream      ModelEventStream
	serverEventStream     ServerEventStream
	experimentEventStream ExperimentEventStream
	pipelineEventStream   PipelineEventStream
}

type ModelEventStream struct {
	mu      sync.Mutex
	streams map[pb.Scheduler_SubscribeModelStatusServer]*ModelSubscription
}

type ServerEventStream struct {
	mu      sync.Mutex
	streams map[pb.Scheduler_SubscribeServerStatusServer]*ServerSubscription
}

type ExperimentEventStream struct {
	mu      sync.Mutex
	streams map[pb.Scheduler_SubscribeExperimentStatusServer]*ExperimentSubscription
}

type PipelineEventStream struct {
	mu      sync.Mutex
	streams map[pb.Scheduler_SubscribePipelineStatusServer]*PipelineSubscription
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

type ExperimentSubscription struct {
	name   string
	stream pb.Scheduler_SubscribeExperimentStatusServer
	fin    chan bool
}

type PipelineSubscription struct {
	name   string
	stream pb.Scheduler_SubscribePipelineStatusServer
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

func NewSchedulerServer(
	logger log.FieldLogger,
	modelStore store.ModelStore,
	experiementServer experiment.ExperimentServer,
	pipelineHandler pipeline.PipelineHandler,
	scheduler scheduler2.Scheduler,
	eventHub *coordinator.EventHub,
) *SchedulerServer {
	s := &SchedulerServer{
		logger:           logger.WithField("source", "SchedulerServer"),
		modelStore:       modelStore,
		experimentServer: experiementServer,
		pipelineHandler:  pipelineHandler,
		scheduler:        scheduler,
		modelEventStream: ModelEventStream{
			streams: make(map[pb.Scheduler_SubscribeModelStatusServer]*ModelSubscription),
		},
		serverEventStream: ServerEventStream{
			streams: make(map[pb.Scheduler_SubscribeServerStatusServer]*ServerSubscription),
		},
		pipelineEventStream: PipelineEventStream{
			streams: make(map[pb.Scheduler_SubscribePipelineStatusServer]*PipelineSubscription),
		},
		experimentEventStream: ExperimentEventStream{
			streams: make(map[pb.Scheduler_SubscribeExperimentStatusServer]*ExperimentSubscription),
		},
	}

	eventHub.RegisterModelEventHandler(
		modelEventHandlerName,
		pendingEventsQueueSize,
		s.logger,
		s.handleModelEvent,
	)
	eventHub.RegisterModelEventHandler(
		serverEventHandlerName,
		pendingEventsQueueSize,
		s.logger,
		s.handleServerEvent,
	)
	eventHub.RegisterExperimentEventHandler(
		experimentEventHandlerName,
		pendingEventsQueueSize,
		s.logger,
		s.handleExperimentEvents,
	)

	eventHub.RegisterPipelineEventHandler(
		pipelineEventHandlerName,
		pendingEventsQueueSize,
		s.logger,
		s.handlePipelineEvents,
	)

	return s
}

func (s *SchedulerServer) ServerNotify(ctx context.Context, req *pb.ServerNotifyRequest) (*pb.ServerNotifyResponse, error) {
	logger := s.logger.WithField("func", "ServerNotify")
	logger.Infof("Server notification %s expectedReplicas %d shared %v", req.GetName(), req.GetExpectedReplicas(), req.GetShared())
	err := s.modelStore.ServerNotify(req)
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
	server, err := s.modelStore.GetServer(serverKey)
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
	err := s.modelStore.UpdateModel(req)
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
	err := s.modelStore.RemoveModel(req)
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
		ModelDefn: mv.GetModel(),
	}
	if mv.GetMeta().KubernetesMeta != nil {
		mvs.KubernetesMeta = mv.GetModel().GetMeta().GetKubernetesMeta()
	}
	return mvs
}

func (s *SchedulerServer) modelStatusImpl(model *store.ModelSnapshot, allVersions bool) (*pb.ModelStatusResponse, error) {
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

func (s *SchedulerServer) ModelStatus(
	req *pb.ModelStatusRequest,
	stream pb.Scheduler_ModelStatusServer,
) error {
	logger := s.logger.WithField("func", "ModelStatus")
	logger.Infof("received status request from %s", req.SubscriberName)

	if req.Model == nil {
		// All models requested
		models, err := s.modelStore.GetModels()
		if err != nil {
			return status.Errorf(codes.FailedPrecondition, err.Error())
		}

		for _, m := range models {
			resp, err := s.modelStatusImpl(m, req.AllVersions)
			if err != nil {
				return status.Errorf(codes.FailedPrecondition, err.Error())
			}

			err = stream.Send(resp)
			if err != nil {
				return status.Errorf(codes.Internal, err.Error())
			}
		}
		return nil
	} else {
		// Single model requested
		s.modelStore.LockModel(req.Model.Name)
		defer s.modelStore.UnlockModel(req.Model.Name)

		model, err := s.modelStore.GetModel(req.Model.Name)
		if err != nil {
			return status.Errorf(codes.FailedPrecondition, err.Error())
		}
		if model == nil || len(model.Versions) == 0 {
			return status.Errorf(
				codes.FailedPrecondition,
				fmt.Sprintf("Failed to find model %s", req.Model.Name),
			)
		}

		resp, err := s.modelStatusImpl(model, req.AllVersions)
		if err != nil {
			return status.Errorf(codes.Internal, err.Error())
		}
		err = stream.Send(resp)
		if err != nil {
			return status.Errorf(codes.Internal, err.Error())
		}
		return nil
	}
}

func (s *SchedulerServer) ServerStatus(
	req *pb.ServerStatusRequest,
	stream pb.Scheduler_ServerStatusServer,
) error {
	logger := s.logger.WithField("func", "ServerStatus")
	logger.Infof("received status request from %s", req.SubscriberName)

	if req.Name == nil {
		// All servers requested
		servers, err := s.modelStore.GetServers()
		if err != nil {
			return status.Errorf(codes.FailedPrecondition, err.Error())
		}

		for _, s := range servers {
			resp := createServerStatusResponse(s)
			err := stream.Send(resp)
			if err != nil {
				return status.Errorf(codes.Internal, err.Error())
			}
		}
		return nil
	} else {
		// Single server requested
		server, err := s.modelStore.GetServer(req.GetName())
		if err != nil {
			return status.Errorf(codes.FailedPrecondition, err.Error())
		}
		if server == nil {
			return status.Errorf(codes.FailedPrecondition, fmt.Sprintf("Failed to find server %s", req.GetName()))
		}
		resp := createServerStatusResponse(server)
		err = stream.Send(resp)
		if err != nil {
			return status.Errorf(codes.Internal, err.Error())
		}
		return nil
	}
}

func createServerStatusResponse(s *store.ServerSnapshot) *pb.ServerStatusResponse {
	resp := &pb.ServerStatusResponse{
		ServerName:        s.Name,
		AvailableReplicas: int32(len(s.Replicas)),
		ExpectedReplicas:  int32(s.ExpectedReplicas),
		KubernetesMeta:    s.KubernetesMeta,
	}

	var totalModels int32
	for _, replica := range s.Replicas {
		numLoadedModelsOnReplica := int32(replica.GetNumLoadedModels())
		resp.Resources = append(
			resp.Resources,
			&pb.ServerReplicaResources{
				ReplicaIdx:           uint32(replica.GetReplicaIdx()),
				TotalMemoryBytes:     replica.GetMemory(),
				AvailableMemoryBytes: replica.GetAvailableMemory(),
				NumLoadedModels:      numLoadedModelsOnReplica,
				OverCommitPercentage: replica.GetOverCommitPercentage(),
			},
		)
		totalModels = totalModels + numLoadedModelsOnReplica
	}
	resp.NumLoadedModelReplicas = totalModels

	return resp
}

func (s *SchedulerServer) StartExperiment(ctx context.Context, req *pb.StartExperimentRequest) (*pb.StartExperimentResponse, error) {
	err := s.experimentServer.StartExperiment(experiment.CreateExperimentFromRequest(req.Experiment))
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	return &pb.StartExperimentResponse{}, nil
}

func (s *SchedulerServer) StopExperiment(ctx context.Context, req *pb.StopExperimentRequest) (*pb.StopExperimentResponse, error) {
	err := s.experimentServer.StopExperiment(req.GetName())
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	return &pb.StopExperimentResponse{}, nil
}

func (s *SchedulerServer) ExperimentStatus(
	req *pb.ExperimentStatusRequest,
	stream pb.Scheduler_ExperimentStatusServer,
) error {
	logger := s.logger.WithField("func", "ExperimentStatus")
	logger.Infof("received status request from %s", req.SubscriberName)

	if req.Name == nil {
		// All experiments requested
		experiments, err := s.experimentServer.GetExperiments()
		if err != nil {
			return status.Errorf(codes.FailedPrecondition, err.Error())
		}

		for _, e := range experiments {
			resp := createExperimentStatus(e)
			err = stream.Send(resp)
			if err != nil {
				return status.Errorf(codes.Internal, err.Error())
			}
		}
		return nil
	} else {
		// Single experiment requested
		exp, err := s.experimentServer.GetExperiment(req.GetName())
		if err != nil {
			return status.Errorf(codes.FailedPrecondition, err.Error())
		}

		resp := createExperimentStatus(exp)
		err = stream.Send(resp)
		if err != nil {
			return status.Errorf(codes.Internal, err.Error())
		}
		return nil
	}
}

func createExperimentStatus(e *experiment.Experiment) *pb.ExperimentStatusResponse {
	response := &pb.ExperimentStatusResponse{
		ExperimentName:    e.Name,
		Active:            e.Active,
		StatusDescription: e.StatusDescription,
		CandidatesReady:   e.AreCandidatesReady(),
		MirrorReady:       e.IsMirrorReady(),
	}
	if e.KubernetesMeta != nil {
		response.KubernetesMeta = &pb.KubernetesMeta{
			Namespace:  e.KubernetesMeta.Namespace,
			Generation: e.KubernetesMeta.Generation,
		}
	}
	return response
}

func (s *SchedulerServer) LoadPipeline(ctx context.Context, req *pb.LoadPipelineRequest) (*pb.LoadPipelineResponse, error) {
	err := s.pipelineHandler.AddPipeline(req.Pipeline)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	return &pb.LoadPipelineResponse{}, nil
}

func (s *SchedulerServer) UnloadPipeline(ctx context.Context, req *pb.UnloadPipelineRequest) (*pb.UnloadPipelineResponse, error) {
	err := s.pipelineHandler.RemovePipeline(req.GetName())
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	return &pb.UnloadPipelineResponse{}, nil
}

func createPipelineStatus(p *pipeline.Pipeline, allVersions bool) *pb.PipelineStatusResponse {
	var pipelineVersions []*pb.PipelineWithState
	pipelineLastVersion := p.GetLatestPipelineVersion()
	if !allVersions {
		pipelineWithState := pipeline.CreatePipelineWithState(pipelineLastVersion)
		pipelineVersions = append(pipelineVersions, pipelineWithState)
	} else {
		for _, pv := range p.Versions {
			pipelineWithState := pipeline.CreatePipelineWithState(pv)
			pipelineVersions = append(pipelineVersions, pipelineWithState)
		}
	}
	return &pb.PipelineStatusResponse{
		PipelineName: pipelineLastVersion.Name,
		Versions:     pipelineVersions,
	}
}

func (s *SchedulerServer) PipelineStatus(
	req *pb.PipelineStatusRequest,
	stream pb.Scheduler_PipelineStatusServer,
) error {
	logger := s.logger.WithField("func", "PipelineStatus")
	logger.Infof("received status request from %s", req.SubscriberName)

	if req.Name == nil {

		pipelines, err := s.pipelineHandler.GetPipelines()
		if err != nil {
			return status.Errorf(codes.FailedPrecondition, err.Error())
		}
		for _, p := range pipelines {
			resp := createPipelineStatus(p, req.GetAllVersions())
			err = stream.Send(resp)
			if err != nil {
				return status.Errorf(codes.Internal, err.Error())
			}
		}
		return nil
	} else {
		// Single pipeline requested
		p, err := s.pipelineHandler.GetPipeline(req.GetName())
		if err != nil {
			return status.Errorf(codes.FailedPrecondition, err.Error())
		}
		resp := createPipelineStatus(p, req.GetAllVersions())
		err = stream.Send(resp)
		if err != nil {
			return status.Errorf(codes.Internal, err.Error())
		}
		return nil
	}
}

func (s *SchedulerServer) SchedulerStatus(ctx context.Context, req *pb.SchedulerStatusRequest) (*pb.SchedulerStatusResponse, error) {
	logger := s.logger.WithField("func", "SchedulerStatus")
	logger.Infof("received status request from %s", req.SubscriberName)

	return &pb.SchedulerStatusResponse{
		ApplicationVersion: "0.0.1",
	}, nil
}
