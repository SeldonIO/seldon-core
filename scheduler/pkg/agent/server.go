package agent

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/seldonio/seldon-core/scheduler/pkg/coordinator"

	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	"github.com/seldonio/seldon-core/scheduler/pkg/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	grpcMaxConcurrentStreams     = 1_000_000
	pendingSyncsQueueSize    int = 10
	modelEventHandlerName        = "agent.server.models"
)

type ServerKey struct {
	serverName string
	replicaIdx uint32
}

type Server struct {
	mutext sync.RWMutex
	pb.UnimplementedAgentServiceServer
	logger    log.FieldLogger
	agents    map[ServerKey]*AgentSubscriber
	store     store.ModelStore
	scheduler scheduler.Scheduler
}

type SchedulerAgent interface {
	modelSync(modelName string) error
}

type AgentSubscriber struct {
	finished chan<- bool
	mutext   sync.Mutex // grpc streams are not thread safe for sendMsg https://github.com/grpc/grpc-go/issues/2355
	stream   pb.AgentService_SubscribeServer
}

func NewAgentServer(
	logger log.FieldLogger,
	store store.ModelStore,
	scheduler scheduler.Scheduler,
	hub *coordinator.EventHub,
) *Server {
	s := &Server{
		logger:    logger.WithField("source", "AgentServer"),
		agents:    make(map[ServerKey]*AgentSubscriber),
		store:     store,
		scheduler: scheduler,
	}

	hub.RegisterModelEventHandler(
		modelEventHandlerName,
		pendingSyncsQueueSize,
		s.logger,
		s.handleSyncs,
	)

	return s
}

func (s *Server) handleSyncs(event coordinator.ModelEventMsg) {
	logger := s.logger.WithField("func", "handleSyncs")
	logger.Infof("Received sync for model %s", event.String())

	// TODO - Should this spawn a goroutine?
	// Surely we're risking reordering of events, e.g. load/unload -> unload/load?
	go s.Sync(event.ModelName)
}

func (s *Server) StartGrpcServer(agentPort uint) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", agentPort))
	if err != nil {
		log.Fatalf("failed to create listener: %v", err)
	}
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	grpcServer := grpc.NewServer(grpcOptions...)
	pb.RegisterAgentServiceServer(grpcServer, s)
	s.logger.Printf("Agent server running on %d", agentPort)
	return grpcServer.Serve(lis)
}

func (s *Server) Sync(modelName string) {
	logger := s.logger.WithField("func", "Sync")
	s.mutext.RLock()
	defer s.mutext.RUnlock()
	s.store.LockModel(modelName)
	defer s.store.UnlockModel(modelName)

	model, err := s.store.GetModel(modelName)
	if err != nil {
		logger.WithError(err).Error("Sync failed")
		return
	}
	if model == nil {
		logger.Errorf("Model %s not found", modelName)
		return
	}

	// Handle any load requests for latest version - we don't want to load models from older versions
	latestModel := model.GetLatest()
	if latestModel != nil {
		for _, replicaIdx := range latestModel.GetReplicaForState(store.LoadRequested) {
			logger.Infof("Sending load model request for %s", modelName)

			as, ok := s.agents[ServerKey{serverName: latestModel.Server(), replicaIdx: uint32(replicaIdx)}]

			if !ok {
				logger.Errorf("Failed to find server replica for %s:%d", latestModel.Server(), replicaIdx)
				continue
			}

			as.mutext.Lock()
			err = as.stream.Send(&pb.ModelOperationMessage{
				Operation:    pb.ModelOperationMessage_LOAD_MODEL,
				ModelVersion: &pb.ModelVersion{Model: latestModel.GetModel(), Version: latestModel.GetVersion()},
			})
			as.mutext.Unlock()
			if err != nil {
				logger.WithError(err).Errorf("stream message send failed for model %s and replicaidx %d", modelName, replicaIdx)
				continue
			}
			err := s.store.UpdateModelState(latestModel.Key(), latestModel.GetVersion(), latestModel.Server(), replicaIdx, nil, store.LoadRequested, store.Loading, "")
			if err != nil {
				logger.WithError(err).Errorf("Sync set model state failed for model %s replicaidx %d", modelName, replicaIdx)
				continue
			}
		}
	}

	// Loop through all versions and unload any requested - any version of a model might have an unload request
	for _, modelVersion := range model.Versions {
		for _, replicaIdx := range modelVersion.GetReplicaForState(store.UnloadRequested) {
			s.logger.Infof("Sending unload model request for %s:%d", modelName, modelVersion.GetVersion())
			as, ok := s.agents[ServerKey{serverName: modelVersion.Server(), replicaIdx: uint32(replicaIdx)}]
			if !ok {
				logger.Errorf("Failed to find server replica for %s:%d", modelVersion.Server(), replicaIdx)
				continue
			}
			as.mutext.Lock()
			err = as.stream.Send(&pb.ModelOperationMessage{
				Operation:    pb.ModelOperationMessage_UNLOAD_MODEL,
				ModelVersion: &pb.ModelVersion{Model: modelVersion.GetModel(), Version: modelVersion.GetVersion()},
			})
			as.mutext.Unlock()
			if err != nil {
				logger.WithError(err).Errorf("stream message send failed for model %s and replicaidx %d", modelName, replicaIdx)
				continue
			}
			err := s.store.UpdateModelState(modelVersion.Key(), modelVersion.GetVersion(), modelVersion.Server(), replicaIdx, nil, store.UnloadRequested, store.Unloading, "")
			if err != nil {
				logger.WithError(err).Errorf("Sync set model state failed for model %s replicaidx %d", modelName, replicaIdx)
				continue
			}
		}
	}
}

func (s *Server) AgentEvent(ctx context.Context, message *pb.ModelEventMessage) (*pb.ModelEventResponse, error) {
	logger := s.logger.WithField("func", "AgentEvent")
	var desiredState store.ModelReplicaState
	var expectedState store.ModelReplicaState
	switch message.Event {
	case pb.ModelEventMessage_LOADED:
		expectedState = store.Loading
		desiredState = store.Loaded
	case pb.ModelEventMessage_UNLOADED:
		expectedState = store.Unloading
		desiredState = store.Unloaded
	case pb.ModelEventMessage_LOAD_FAILED,
		pb.ModelEventMessage_LOAD_FAIL_MEMORY:
		expectedState = store.Loading
		desiredState = store.LoadFailed
	case pb.ModelEventMessage_UNLOAD_FAILED:
		expectedState = store.Unloading
		desiredState = store.UnloadFailed
	default:
		desiredState = store.ModelReplicaStateUnknown
	}
	logger.Infof("Updating state for model %s to %s", message.ModelName, desiredState.String())
	s.store.LockModel(message.ModelName)
	defer s.store.UnlockModel(message.ModelName)
	err := s.store.UpdateModelState(message.ModelName, message.GetModelVersion(), message.ServerName, int(message.ReplicaIdx), &message.AvailableMemoryBytes, expectedState, desiredState, message.GetMessage())
	if err != nil {
		logger.WithError(err).Infof("Failed Updating state for model %s", message.ModelName)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.ModelEventResponse{}, nil
}

func (s *Server) Subscribe(request *pb.AgentSubscribeRequest, stream pb.AgentService_SubscribeServer) error {
	logger := s.logger.WithField("func", "Subscribe")
	logger.Infof("Received subscribe request from %s:%d", request.ServerName, request.ReplicaIdx)

	fin := make(chan bool)

	s.mutext.Lock()
	s.agents[ServerKey{serverName: request.ServerName, replicaIdx: request.ReplicaIdx}] = &AgentSubscriber{
		finished: fin,
		stream:   stream,
	}
	s.mutext.Unlock()

	err := s.syncMessage(request, stream)
	if err != nil {
		return err
	}

	ctx := stream.Context()
	// Keep this scope alive because once this scope exits - the stream is closed
	for {
		select {
		case <-fin:
			logger.Infof("Closing stream for replica: %s:%d", request.ServerName, request.ReplicaIdx)
			return nil
		case <-ctx.Done():
			logger.Infof("Client replica %s:%d has disconnected", request.ServerName, request.ReplicaIdx)
			s.mutext.Lock()
			delete(s.agents, ServerKey{serverName: request.ServerName, replicaIdx: request.ReplicaIdx})
			s.mutext.Unlock()
			modelsChanged, err := s.store.RemoveServerReplica(request.ServerName, int(request.ReplicaIdx))
			if err != nil {
				logger.WithError(err).Errorf("Failed to remove replica and redeploy models for %s:%d", request.ServerName, request.ReplicaIdx)
			}
			s.logger.Debugf("Models changed by disconnect %v", modelsChanged)
			for _, modelName := range modelsChanged {
				err = s.scheduler.Schedule(modelName)
				if err != nil {
					logger.Debugf("Failed to reschedule model %s when server %s replica %d disconnected", modelName, request.ServerName, request.ReplicaIdx)
				}
			}
			return nil
		}
	}
}

func (s *Server) syncMessage(request *pb.AgentSubscribeRequest, stream pb.AgentService_SubscribeServer) error {
	s.mutext.Lock()
	defer s.mutext.Unlock()

	s.logger.Debugf("Add Server Replica %+v with config %+v", request, request.ReplicaConfig)
	err := s.store.AddServerReplica(request)
	if err != nil {
		return err
	}
	_, err = s.scheduler.ScheduleFailedModels()
	if err != nil {
		return err
	}
	return nil
}
