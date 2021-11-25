package agent

import (
	"context"
	"fmt"
	"net"
	"sync"

	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/processor"
	"github.com/seldonio/seldon-core/scheduler/pkg/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ServerKey struct {
	serverName string
	replicaIdx uint32
}

type AgentHandler interface {
	SendAgentSync(modelName string)
}

type Server struct {
	mutext sync.RWMutex
	pb.UnimplementedAgentServiceServer
	logger       log.FieldLogger
	agents       map[ServerKey]*AgentSubscriber
	store        store.SchedulerStore
	envoyHandler processor.EnvoyHandler
	source       chan string
	scheduler    scheduler.Scheduler
}

type SchedulerAgent interface {
	Sync(modelName string) error
}

type AgentSubscriber struct {
	finished chan<- bool
	//mutext   sync.Mutex // grpc streams are not thread safe for sendMsg https://github.com/grpc/grpc-go/issues/2355
	stream pb.AgentService_SubscribeServer
}

func NewAgentServer(logger log.FieldLogger,
	store store.SchedulerStore,
	envoyHandler processor.EnvoyHandler,
	scheduler scheduler.Scheduler) *Server {
	return &Server{
		logger:       logger.WithField("source", "AgentServer"),
		agents:       make(map[ServerKey]*AgentSubscriber),
		store:        store,
		envoyHandler: envoyHandler,
		source:       make(chan string, 1),
		scheduler:    scheduler,
	}
}

func (s *Server) SendAgentSync(modelName string) {
	s.source <- modelName
}

func (s *Server) StopAgentSync() {
	close(s.source)
}

func (s *Server) ListenForSyncs() {
	for modelName := range s.source {
		s.logger.Infof("Received sync for model %s", modelName)
		go s.Sync(modelName)
	}
}

func (s *Server) StartGrpcServer(agentPort uint) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", agentPort))
	if err != nil {
		log.Fatalf("failed to create listener: %v", err)
	}
	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterAgentServiceServer(grpcServer, s)
	s.logger.Printf("Agent server running on %d", agentPort)
	return grpcServer.Serve(lis)
}

func (s *Server) Sync(modelName string) {
	logger := s.logger.WithField("func", "Sync")
	s.mutext.RLock()
	defer s.mutext.RUnlock()

	model, err := s.store.GetModel(modelName)
	if err != nil {
		logger.WithError(err).Error("Sync failed")
		return
	}
	if model == nil {
		logger.Errorf("Model %s not found", modelName)
		return
	}

	// Handle any load requests for latest version
	latestModel := model.GetLatest()
	if latestModel != nil {
		for _, replicaIdx := range latestModel.GetReplicaForState(store.LoadRequested) {
			logger.Infof("Sending load model request for %s", modelName)

			as, ok := s.agents[ServerKey{serverName: latestModel.Server(), replicaIdx: uint32(replicaIdx)}]

			if !ok {
				logger.Errorf("Failed to find server replica for %s:%d", latestModel.Server(), replicaIdx)
				continue
			}

			err = as.stream.Send(&pb.ModelOperationMessage{
				Operation: pb.ModelOperationMessage_LOAD_MODEL,
				Details:   latestModel.Details(),
			})
			if err != nil {
				logger.WithError(err).Errorf("stream message send failed for model %s and replicaidx %d", modelName, replicaIdx)
				continue
			}
			err := s.store.UpdateModelState(latestModel.Key(), latestModel.GetVersion(), latestModel.Server(), replicaIdx, nil, store.Loading)
			if err != nil {
				logger.WithError(err).Errorf("Sync set model state failed for model %s replicaidx %d", modelName, replicaIdx)
				continue
			}
		}
	}

	// Loop through all versions and unload any requested
	for _, modelVersion := range model.Versions {
		for _, replicaIdx := range modelVersion.GetReplicaForState(store.UnloadRequested) {
			s.logger.Infof("Sending unload model request for %s", modelName)
			as, ok := s.agents[ServerKey{serverName: modelVersion.Server(), replicaIdx: uint32(replicaIdx)}]
			if !ok {
				logger.Errorf("Failed to find server replica for %s:%d", modelVersion.Server(), replicaIdx)
				continue
			}
			err = as.stream.Send(&pb.ModelOperationMessage{
				Operation: pb.ModelOperationMessage_UNLOAD_MODEL,
				Details:   modelVersion.Details(),
			})
			if err != nil {
				logger.WithError(err).Errorf("stream message send failed for model %s and replicaidx %d", modelName, replicaIdx)
				continue
			}
			err := s.store.UpdateModelState(modelVersion.Key(), modelVersion.GetVersion(), modelVersion.Server(), replicaIdx, nil, store.Unloading)
			if err != nil {
				logger.WithError(err).Errorf("Sync set model state failed for model %s replicaidx %d", modelName, replicaIdx)
				continue
			}
		}
	}
}

func (s *Server) AgentEvent(ctx context.Context, message *pb.ModelEventMessage) (*pb.ModelEventResponse, error) {
	logger := s.logger.WithField("func", "AgentEvent")
	var state store.ModelReplicaState
	switch message.Event {
	case pb.ModelEventMessage_LOADED:
		state = store.Loaded
	case pb.ModelEventMessage_UNLOADED:
		state = store.Unloaded
	case pb.ModelEventMessage_LOAD_FAILED,
		pb.ModelEventMessage_LOAD_FAIL_MEMORY:
		state = store.LoadFailed
	default:
		state = store.ModelReplicaStateUnknown
	}
	logger.Infof("Updating state for model %s to %s", message.ModelName, state.String())
	err := s.store.UpdateModelState(message.ModelName, message.GetModelVersion(), message.ServerName, int(message.ReplicaIdx), &message.AvailableMemoryBytes, state)
	if err != nil {
		logger.Infof("Failed Updating state for model %s: err:%s", message.ModelName, err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	s.envoyHandler.SendEnvoySync(message.ModelName)
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
				} else {
					s.SendAgentSync(modelName)
				}

			}
			return nil
		}
	}
}

func (s *Server) syncMessage(request *pb.AgentSubscribeRequest, stream pb.AgentService_SubscribeServer) error {
	s.mutext.Lock()
	defer s.mutext.Unlock()

	err := s.store.AddServerReplica(request)
	if err != nil {
		return err
	}
	updatedModels, err := s.scheduler.ScheduleFailedModels()
	if err != nil {
		return err
	}
	for _, updatedModels := range updatedModels {
		s.SendAgentSync(updatedModels)
	}
	return nil
}
