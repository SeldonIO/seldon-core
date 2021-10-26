package agent

import (
	"context"
	"errors"
	"fmt"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/processor"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
	"sync"
)

type ServerKey struct {
	serverName string
	replicaIdx uint32
}

type Server struct {
	mutext sync.RWMutex
	pb.UnimplementedAgentServiceServer
	logger log.FieldLogger
	agents map[ServerKey]*AgentSubscriber
	store store.SchedulerStore
	EnvoyProcessor *processor.IncrementalProcessor

	source chan string
}

type SchedulerAgent interface {
	Sync(modelName string) error
}

type AgentSubscriber struct {
	finished chan<- bool
	mutext sync.Mutex // grpc streams are not thread safe for sendMsg https://github.com/grpc/grpc-go/issues/2355
	stream pb.AgentService_SubscribeServer
}

func NewAgentServer(logger log.FieldLogger, store store.SchedulerStore, envoyProcessor *processor.IncrementalProcessor, source chan string) *Server {
	return &Server{
		logger: logger.WithField("Source","AgentServer"),
		agents: make(map[ServerKey]*AgentSubscriber),
		store: store,
		EnvoyProcessor: envoyProcessor,
		source: source,
	}
}

func (s *Server) ListenForSyncs() {
	for msg := range s.source {
		s.logger.Infof("Received sync for model %s",msg)
		go s.Sync(msg)
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
	return  grpcServer.Serve(lis)
}

func (s *Server) Sync(modelName string)  {
	s.mutext.RLock()
	defer s.mutext.RUnlock()

	model, err := s.store.GetModel(modelName)
	if err != nil {
		if errors.Is(err, store.ModelNotFoundErr) {
			s.logger.Infof("Model not found in sync for %s", modelName)
			return
		}
		s.logger.WithError(err).Error("Sync failed")
		return
	}
	serverName := model.Server()
	if serverName == "" {
		return
	}

	for _,replicaIdx := range model.GetReplicaForState(store.LoadRequested) {
		s.logger.Infof("Sending load model request for %s", modelName)

		as, ok := s.agents[ServerKey{serverName: model.Server(), replicaIdx: uint32(replicaIdx)}]

		if !ok {
			s.logger.Errorf("Failed to find server replica for %s:%d",model.Server(), replicaIdx)
			continue
		}
		s.logger.Infof("1 About to call set model state for model %s",modelName)
		err := s.store.SetModelState(model.Key(), model.Server(), replicaIdx, store.Loading, nil)
		s.logger.Infof("1 Finished to call set model state for model %s",modelName)
		if err != nil {
			s.logger.WithError(err).Errorf("Sync set model state failed for model %s replicaidx %d",modelName,replicaIdx)
			continue
		}
		err = as.stream.Send(&pb.ModelOperationMessage{
			Operation: pb.ModelOperationMessage_LOAD_MODEL,
			Details: model.Details(),
		})
		s.logger.WithError(err).Errorf("stream message send failed for model %s and replicaidx %d",modelName, replicaIdx)
	}

	for _,replicaIdx := range model.GetReplicaForState(store.UnloadRequested) {
		s.logger.Infof("Sending unload model request for %s", modelName)
		as, ok := s.agents[ServerKey{serverName: model.Server(), replicaIdx: uint32(replicaIdx)}]
		if !ok {
			s.logger.Errorf("Failed to find server replica for %s:%d",model.Server(), replicaIdx)
			continue
		}
		err := s.store.SetModelState(model.Key(), model.Server(), replicaIdx, store.Unloading, nil)
		if err != nil {
			s.logger.WithError(err).Errorf("Sync set model state failed for model %s replicaidx %d",modelName,replicaIdx)
			continue
		}
		err = as.stream.Send(&pb.ModelOperationMessage{
			Operation: pb.ModelOperationMessage_UNLOAD_MODEL,
			Details: model.Details(),
		})
		s.logger.WithError(err).Errorf("stream message send failed for model %s and replicaidx %d",modelName, replicaIdx)
	}

}



func (s *Server) AgentEvent(ctx context.Context, message *pb.ModelEventMessage) (*pb.ModelEventResponse, error) {
	//TODO finish
	var state store.ModelState
	switch message.Event {
	case pb.ModelEventMessage_LOADED:
		state = store.Loaded
	case pb.ModelEventMessage_UNLOADED:
		state = store.Unloaded
	case pb.ModelEventMessage_LOAD_FAILED,
		pb.ModelEventMessage_LOAD_FAIL_MEMORY:
		state = store.LoadFailed
	default:
		state = store.Unknown
	}
	s.logger.Infof("Updating state for model %s",message.ModelName)
	err := s.store.SetModelState(message.ModelName, message.ServerName, int(message.ReplicaIdx), state, &message.AvailableMemory)
	if err != nil {
		s.logger.Infof("Failed Updating state for model %s",message.ModelName)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.ModelEventResponse{}, nil
}

func (s *Server) Subscribe(request *pb.AgentSubscribeRequest, stream pb.AgentService_SubscribeServer) error {
	s.logger.Infof("Received subscribe request from %s:%d",request.ServerName, request.ReplicaIdx)

	fin := make(chan bool)

	s.mutext.Lock()
	s.agents[ServerKey{serverName: request.ServerName, replicaIdx: request.ReplicaIdx}] = &AgentSubscriber{
		finished: fin,
		stream: stream,
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
			log.Printf("Closing stream for replica: %s:%d", request.ServerName, request.ReplicaIdx)
			return nil
		case <- ctx.Done():
			log.Printf("Client replica %s:%d has disconnected", request.ServerName, request.ReplicaIdx)
			s.mutext.Lock()
			delete(s.agents,ServerKey{serverName: request.ServerName, replicaIdx: request.ReplicaIdx})
			s.mutext.Unlock()
			err := s.store.RemoveServerReplicaAndRedeployModels(request.ServerName, int(request.ReplicaIdx))
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to remove replica and redeploy models for %s:%d",request.ServerName, request.ReplicaIdx)
			}
			return nil
		}
	}
}

func (s *Server) syncMessage(request *pb.AgentSubscribeRequest, stream pb.AgentService_SubscribeServer) error {
	s.mutext.Lock()
	defer s.mutext.Unlock()

	err := s.store.UpdateServerReplica(request)
	if err != nil {
		return err
	}
	serverReplica, err := s.store.GetServerReplica(request.GetServerName(), int(request.GetReplicaIdx()))
	if err != nil {
		return err
	}

	// Send our state to agent
	for _, modelName := range serverReplica.GetLoadedModels() {
		model, err := s.store.GetModel(modelName)
		if err != nil || model.GetModelReplicaState(int(request.ReplicaIdx)) == store.UnloadRequested {
			err := stream.Send(&pb.ModelOperationMessage{
				Operation: pb.ModelOperationMessage_UNLOAD_MODEL,
				Details:   model.Details(),
			})
			if err != nil {
				return err
			}
			err2 := s.store.SetModelState(model.Key(), model.Server(), int(request.ReplicaIdx), store.Unloading, nil)
			if err2 != nil {
				return err2
			}
		}
	}
	return nil
}



