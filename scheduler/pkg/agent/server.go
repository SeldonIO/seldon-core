/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package agent

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	seldontls "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	pbs "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/scheduler"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	grpcMaxConcurrentStreams           = 1_000_000
	pendingSyncsQueueSize          int = 10
	modelEventHandlerName              = "agent.server.models"
	modelScalingCoolingDownSeconds     = 60 // this is currently used in scale down events
	serverDrainingExtraWaitMillis      = 500
)

type modelRelocatedWaiter struct {
	serverReplicaModels map[string]map[string]struct{}
	mu                  sync.Mutex
	waiters             map[string]*sync.WaitGroup
}

func newModelRelocatedWaiter() *modelRelocatedWaiter {
	return &modelRelocatedWaiter{
		serverReplicaModels: map[string]map[string]struct{}{},
		mu:                  sync.Mutex{},
		waiters:             make(map[string]*sync.WaitGroup),
	}
}

func (w *modelRelocatedWaiter) registerServerReplica(serverName string, serverReplicaIdx int, models []string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(models) == 0 {
		return
	}

	key := w.getServerReplicaName(serverName, serverReplicaIdx)
	w.serverReplicaModels[key] = make(map[string]struct{})
	for _, model := range models {
		w.serverReplicaModels[key][model] = struct{}{}
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	w.waiters[key] = &wg
}

func (w *modelRelocatedWaiter) signalModel(model string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for serverReplica, models := range w.serverReplicaModels {
		delete(models, model)
		if len(models) == 0 {
			delete(w.serverReplicaModels, serverReplica)
			w.waiters[serverReplica].Done()
		}
	}
}

func (w *modelRelocatedWaiter) wait(serverName string, serverReplicaIdx int) {
	key := w.getServerReplicaName(serverName, serverReplicaIdx)
	if wg, ok := w.waiters[key]; ok {
		wg.Wait()
	}
}

func (w *modelRelocatedWaiter) getServerReplicaName(serverName string, serverReplicaIdx int) string {
	return serverName + "_" + strconv.Itoa(serverReplicaIdx)
}

type ServerKey struct {
	serverName string
	replicaIdx uint32
}

type Server struct {
	mutex sync.RWMutex
	pb.UnimplementedAgentServiceServer
	logger                    log.FieldLogger
	agents                    map[ServerKey]*AgentSubscriber
	store                     store.ModelStore
	scheduler                 scheduler.Scheduler
	certificateStore          *seldontls.CertificateStore
	waiter                    *modelRelocatedWaiter // waiter for when we want to drain a particular server replica
	autoscalingServiceEnabled bool
}

type SchedulerAgent interface {
	modelSync(modelName string) error
}

type AgentSubscriber struct {
	finished chan<- bool
	mutex    sync.Mutex // grpc streams are not thread safe for sendMsg https://github.com/grpc/grpc-go/issues/2355
	stream   pb.AgentService_SubscribeServer
}

func NewAgentServer(
	logger log.FieldLogger,
	store store.ModelStore,
	scheduler scheduler.Scheduler,
	hub *coordinator.EventHub,
	autoscalingServiceEnabled bool,
) *Server {
	s := &Server{
		logger:                    logger.WithField("source", "AgentServer"),
		agents:                    make(map[ServerKey]*AgentSubscriber),
		store:                     store,
		scheduler:                 scheduler,
		waiter:                    newModelRelocatedWaiter(),
		autoscalingServiceEnabled: autoscalingServiceEnabled,
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
	logger.Debugf("Received sync for model %s", event.String())

	// TODO - Should this spawn a goroutine?
	// Surely we're risking reordering of events, e.g. load/unload -> unload/load?
	go s.Sync(event.ModelName)
}

func (s *Server) startServer(port uint, secure bool) error {
	logger := s.logger.WithField("func", "startServer")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	opts := []grpc.ServerOption{}
	if secure {
		opts = append(opts, grpc.Creds(s.certificateStore.CreateServerTransportCredentials()))
	}
	opts = append(opts, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	opts = append(opts, grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()))
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterAgentServiceServer(grpcServer, s)
	s.logger.Printf("Agent server running on %d mtls:%v", port, secure)
	go func() {
		err := grpcServer.Serve(lis)
		if err != nil {
			logger.WithError(err).Fatalf("Agent server failed on port %d mtls:%v", port, secure)
		} else {
			logger.Infof("Agent serving stopped on port %d mtls:%v", port, secure)
		}
	}()
	return nil
}

func (s *Server) StartGrpcServer(allowPlainTxt bool, agentPort uint, agentTlsPort uint) error {
	logger := s.logger.WithField("func", "StartGrpcServer")
	var err error
	protocol := seldontls.GetSecurityProtocolFromEnv(seldontls.EnvSecurityPrefixControlPlane)
	if protocol == seldontls.SecurityProtocolSSL {
		s.certificateStore, err = seldontls.NewCertificateStore(seldontls.Prefix(seldontls.EnvSecurityPrefixControlPlaneServer),
			seldontls.ValidationPrefix(seldontls.EnvSecurityPrefixControlPlaneClient))
		if err != nil {
			return err
		}
	}
	if !allowPlainTxt && s.certificateStore == nil {
		return fmt.Errorf("One of plain txt or mTLS needs to be defined. But have plain text [%v] and no TLS", allowPlainTxt)
	}
	if allowPlainTxt {
		err := s.startServer(agentPort, false)
		if err != nil {
			return err
		}
	} else {
		logger.Info("Not starting scheduler plain text server")
	}
	if s.certificateStore != nil {
		err := s.startServer(agentTlsPort, true)
		if err != nil {
			return err
		}
	} else {
		logger.Info("Not starting scheduler mTLS server")
	}
	return nil
}

func (s *Server) Sync(modelName string) {
	logger := s.logger.WithField("func", "Sync")
	s.mutex.RLock()
	defer s.mutex.RUnlock()
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

	latestModel := model.GetLatest()

	// we signal a model when other replica is available in case we have servers draining
	// TODO: extract as helper func
	if latestModel != nil {
		available := latestModel.GetReplicaForState(store.Available)
		if len(available) > 0 {
			s.waiter.signalModel(modelName)
		}
	}

	// Handle any load requests for latest version - we don't want to load models from older versions
	if latestModel != nil {
		for _, replicaIdx := range latestModel.GetReplicaForState(store.LoadRequested) {
			logger.Infof("Sending load model request for %s", modelName)

			as, ok := s.agents[ServerKey{serverName: latestModel.Server(), replicaIdx: uint32(replicaIdx)}]

			if !ok {
				logger.Errorf("Failed to find server replica for %s:%d", latestModel.Server(), replicaIdx)
				continue
			}

			as.mutex.Lock()
			err = as.stream.Send(&pb.ModelOperationMessage{
				Operation:          pb.ModelOperationMessage_LOAD_MODEL,
				ModelVersion:       &pb.ModelVersion{Model: latestModel.GetModel(), Version: latestModel.GetVersion()},
				AutoscalingEnabled: AutoscalingEnabled(latestModel.GetModel()) && s.autoscalingServiceEnabled,
			})
			as.mutex.Unlock()
			if err != nil {
				logger.WithError(err).Errorf("stream message send failed for model %s and replicaidx %d", modelName, replicaIdx)
				if errState := s.store.UpdateModelState(
					latestModel.Key(), latestModel.GetVersion(), latestModel.Server(), replicaIdx, nil,
					store.LoadRequested, store.LoadFailed, err.Error()); errState != nil {
					logger.WithError(errState).Errorf("Sync set model state failed for model %s replicaidx %d", modelName, replicaIdx)
				}
				continue
			}
			err = s.store.UpdateModelState(latestModel.Key(), latestModel.GetVersion(), latestModel.Server(), replicaIdx, nil, store.LoadRequested, store.Loading, "")
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
			as.mutex.Lock()
			err = as.stream.Send(&pb.ModelOperationMessage{
				Operation:    pb.ModelOperationMessage_UNLOAD_MODEL,
				ModelVersion: &pb.ModelVersion{Model: modelVersion.GetModel(), Version: modelVersion.GetVersion()},
			})
			as.mutex.Unlock()
			if err != nil {
				logger.WithError(err).Errorf("stream message send failed for model %s and replicaidx %d", modelName, replicaIdx)
				if errState := s.store.UpdateModelState(
					latestModel.Key(), latestModel.GetVersion(), latestModel.Server(), replicaIdx, nil,
					store.UnloadRequested, store.UnloadFailed, err.Error()); errState != nil {
					logger.WithError(errState).Errorf("Sync set model state failed for model %s replicaidx %d", modelName, replicaIdx)
				}
				continue
			}
			err = s.store.UpdateModelState(modelVersion.Key(), modelVersion.GetVersion(), modelVersion.Server(), replicaIdx, nil, store.UnloadRequested, store.Unloading, "")
			if err != nil {
				logger.WithError(err).Errorf("Sync set model state failed for model %s replicaidx %d", modelName, replicaIdx)
				continue
			}
		}
	}
}

func (s *Server) AgentDrain(ctx context.Context, message *pb.AgentDrainRequest) (*pb.AgentDrainResponse, error) {
	logger := s.logger.WithField("func", "AgentDrain")
	logger.Infof("Draining server replica %s:%d", message.GetServerName(), message.GetReplicaIdx())
	s.drainServerReplicaImpl(message.GetServerName(), int(message.GetReplicaIdx()))
	return &pb.AgentDrainResponse{Success: true}, nil
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

func (s *Server) ModelScalingTrigger(stream pb.AgentService_ModelScalingTriggerServer) error {
	for {
		message, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.ModelScalingTriggerResponse{})
		}
		if err != nil {
			return err
		}
		logger := s.logger.WithField("func", "ModelScalingTrigger")
		logger.Debugf("Received scaling event %d from server %s:%d for model %s:%d",
			message.GetTrigger(), message.GetServerName(), message.GetReplicaIdx(), message.GetModelName(), message.GetModelVersion())

		// so far we do not care about order of scaling events; the first one should win
		go func() {
			if err := s.applyModelScaling(message); err != nil {
				logger.WithError(err).Debugf(
					"Could not scale model %s:%d, type: %d", message.GetModelName(), message.GetModelVersion(), message.GetTrigger())
			}
		}()
	}
}

func (s *Server) Subscribe(request *pb.AgentSubscribeRequest, stream pb.AgentService_SubscribeServer) error {
	logger := s.logger.WithField("func", "Subscribe")
	logger.Infof("Received subscribe request from %s:%d", request.ServerName, request.ReplicaIdx)

	fin := make(chan bool)

	s.mutex.Lock()
	s.agents[ServerKey{serverName: request.ServerName, replicaIdx: request.ReplicaIdx}] = &AgentSubscriber{
		finished: fin,
		stream:   stream,
	}
	s.mutex.Unlock()

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
			s.mutex.Lock()
			delete(s.agents, ServerKey{serverName: request.ServerName, replicaIdx: request.ReplicaIdx})
			s.mutex.Unlock()
			s.removeServerReplicaImpl(request.GetServerName(), int(request.GetReplicaIdx())) // this is non-blocking beyond rescheduling models on removed server
			return nil
		}
	}
}

func (s *Server) StopAgentStreams() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for _, subscription := range s.agents {
		close(subscription.finished)
	}
}

func (s *Server) syncMessage(request *pb.AgentSubscribeRequest, stream pb.AgentService_SubscribeServer) error {
	s.logger.Debugf("Add Server Replica %+v with config %+v", request, request.ReplicaConfig)
	err := s.store.AddServerReplica(request)
	if err != nil {
		return err
	}

	// we have to reschedule models that are loaded on the incoming agent
	// this is because we can have a network glitch that causes the communication between the agent and the scheduler
	// to drop and the scheduler loading the models on other servers.
	// we need then to reconcile this case
	for _, model := range request.LoadedModels {
		modelName := model.GetModel().GetMeta().GetName()
		if err := s.scheduler.Schedule(model.GetModel().GetMeta().GetName()); err != nil {
			s.logger.WithError(err).Warnf("Failed to reschedule model %s from agent starting with state", modelName)
		}
	}

	_, err = s.scheduler.ScheduleFailedModels()
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) removeServerReplicaImpl(serverName string, serverReplicaIdx int) {
	modelsChanged, err := s.store.RemoveServerReplica(serverName, serverReplicaIdx)
	if err != nil {
		s.logger.WithError(err).Errorf("Failed to remove replica and redeploy models for %s:%d", serverName, serverReplicaIdx)
	}
	s.logger.Debugf("Removing models %v from server %s:%d", modelsChanged, serverName, serverReplicaIdx)
	for _, modelName := range modelsChanged {
		err = s.scheduler.Schedule(modelName)
		if err != nil {
			s.logger.Debugf("Failed to reschedule model %s when server %s replica %d disconnected", modelName, serverName, serverReplicaIdx)
		}
	}
}

func (s *Server) drainServerReplicaImpl(serverName string, serverReplicaIdx int) {
	modelsChanged, err := s.store.DrainServerReplica(serverName, serverReplicaIdx)
	if err != nil {
		s.logger.WithError(err).Errorf("Failed to remove replica and redeploy models for %s:%d", serverName, serverReplicaIdx)
		return
	}

	s.waiter.registerServerReplica(serverName, serverReplicaIdx, modelsChanged)

	s.logger.Debugf("Draining models %v from server %s:%d", modelsChanged, serverName, serverReplicaIdx)
	for _, modelName := range modelsChanged {
		err = s.scheduler.Schedule(modelName)
		if err != nil {
			s.logger.Debugf("Failed to reschedule model %s when server %s replica %d draining", modelName, serverName, serverReplicaIdx)
			// we do not want to wait on this model to be ready, as it cant (for the time being)
			s.waiter.signalModel(modelName)
		}
	}
	s.waiter.wait(serverName, serverReplicaIdx)

	// as we update envoy in batches and envoy is eventual consistent, give it time to settle down
	time.Sleep(util.EnvoyUpdateDefaultBatchWaitMillis + (time.Millisecond * serverDrainingExtraWaitMillis))
	s.logger.Debugf("Finished draining models %v from server %s:%d", modelsChanged, serverName, serverReplicaIdx)
}

func (s *Server) applyModelScaling(message *pb.ModelScalingTriggerMessage) error {

	modelName := message.ModelName
	model, err := s.store.GetModel(modelName)
	if err != nil {
		return err
	}
	if model == nil {
		return fmt.Errorf("Model %s not found", modelName)
	}

	modelProto, err := createScalingPseudoRequest(message, model)
	if err != nil {
		return err
	}

	return s.updateAndSchedule(modelProto)
}

func (s *Server) updateAndSchedule(modelProtos *pbs.Model) error {
	modelName := modelProtos.GetMeta().GetName()
	if err := s.store.UpdateModel(&pbs.LoadModelRequest{
		Model: modelProtos,
	}); err != nil {
		return err
	}

	return s.scheduler.Schedule(modelName)
}

func createScalingPseudoRequest(message *pb.ModelScalingTriggerMessage, model *store.ModelSnapshot) (*pbs.Model, error) {
	modelName := message.ModelName

	lastModelVersion := model.GetLatest()
	if lastModelVersion == nil {
		return nil, fmt.Errorf("Model %s does not exist yet, possibly due to scheduler restarting", modelName)
	}
	lastAvailableModelVersion := model.GetLastAvailableModel()
	tryScaleDown := (message.Trigger == pb.ModelScalingTriggerMessage_SCALE_DOWN && !model.Deleted)
	tryScaleUp := (message.Trigger == pb.ModelScalingTriggerMessage_SCALE_UP && lastAvailableModelVersion != nil && lastAvailableModelVersion.GetVersion() == lastModelVersion.GetVersion())

	if !tryScaleUp && !tryScaleDown {
		return nil, fmt.Errorf("Cannot scale model version %s", modelName)
	}

	if lastModelVersion.GetVersion() != message.GetModelVersion() {
		return nil, fmt.Errorf(
			"Model version %s not matching (expected: %d - actual: %d)",
			modelName, lastModelVersion.GetVersion(), message.GetModelVersion())
	}

	modelProtos := lastModelVersion.GetModel() // this is a clone of the protos

	// if we are scaling up:
	// the model should be available
	// if we are scaling down:
	// we reduce the replicas by one and try our best
	// if we have a draining replica while scaling down, this should be still fine I think?
	numReplicas := int(lastModelVersion.GetDeploymentSpec().Replicas)

	if tryScaleDown {
		if !isModelStable(lastModelVersion) {
			return nil, fmt.Errorf("Model %s has changed status recently, skip scaling", modelName)
		}
	}

	if desiredNumReplicas, err := calculateDesiredNumReplicas(modelProtos, message.Trigger, numReplicas); err != nil {
		return nil, err
	} else {
		modelProtos.DeploymentSpec.Replicas = uint32(desiredNumReplicas)
	}
	return modelProtos, nil
}

func isModelStable(modelVersion *store.ModelVersion) bool {
	return modelVersion.ModelState().Timestamp.Before(time.Now().Add(-modelScalingCoolingDownSeconds * time.Second))
}

func calculateDesiredNumReplicas(model *pbs.Model, trigger pb.ModelScalingTriggerMessage_Trigger, numReplicas int) (int, error) {

	if trigger == pb.ModelScalingTriggerMessage_SCALE_UP {
		if err := checkModelScalingWithinRange(model, numReplicas+1); err != nil {
			return 0, err
		} else {
			return numReplicas + 1, nil
		}
	} else if trigger == pb.ModelScalingTriggerMessage_SCALE_DOWN {
		if err := checkModelScalingWithinRange(model, numReplicas-1); err != nil {
			return 0, err
		} else {
			return numReplicas - 1, nil
		}
	}
	return 0, fmt.Errorf("event not supported")
}

// we autoscale if at least min or max replicas is set and that we are within the range
// if a user therefore sets only the number of replicas then autoscaling will not be activated
// which is hidden in this logic unfortunately as we reject the scaling up / down event.
// a side effect is that we do not go below 1 replica of a model
func checkModelScalingWithinRange(model *pbs.Model, targetNumReplicas int) error {
	if !AutoscalingEnabled(model) {
		return fmt.Errorf("No autoscaling for model %s", model.GetMeta().GetName())
	}

	minReplicas := model.DeploymentSpec.GetMinReplicas()
	maxReplicas := model.DeploymentSpec.GetMaxReplicas()

	if targetNumReplicas < int(minReplicas) || (targetNumReplicas < 1) {
		return fmt.Errorf("Violating min replicas %d / %d for model %s", minReplicas, targetNumReplicas, model.GetMeta().GetName())
	}

	if targetNumReplicas > int(maxReplicas) && (maxReplicas > 0) {
		return fmt.Errorf("Violating max replicas %d / %d for model %s", maxReplicas, targetNumReplicas, model.GetMeta().GetName())
	}

	return nil
}

// if min and max replicas are not set, we do not allow autoscaling
// we check that they are not set if they are equal to zero as per
// `GetMinReplicas` and `GetMaxReplicas` definition
func AutoscalingEnabled(model *pbs.Model) bool {
	minReplicas := model.DeploymentSpec.GetMinReplicas()
	maxReplicas := model.DeploymentSpec.GetMaxReplicas()

	if (minReplicas == 0) && (maxReplicas == 0) {
		// no autoscaling
		return false
	} else {
		return true
	}
}
