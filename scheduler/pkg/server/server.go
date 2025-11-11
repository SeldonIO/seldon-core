/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/health"
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	seldontls "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	cr "github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/conflict-resolution"
	scaling_config "github.com/seldonio/seldon-core/scheduler/v2/pkg/scaling/config"
	scheduler2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/scheduler"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/experiment"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/synchroniser"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

const (
	grpcMaxConcurrentStreams              = 1_000_000
	pendingEventsQueueSize            int = 1000
	modelEventHandlerName                 = "scheduler.server.models"
	serverEventHandlerName                = "scheduler.server.servers"
	serverModelEventHandlerName           = "scheduler.server.servers.models"
	experimentEventHandlerName            = "scheduler.server.experiments"
	pipelineEventHandlerName              = "scheduler.server.pipelines"
	defaultBatchWait                      = 250 * time.Millisecond
	sendTimeout                           = 30 * time.Second // Timeout for sending events to subscribers via grpc `sendMsg`
	modelGatewayConsumerNamePrefix        = "seldon-modelgateway"
	pipelineGatewayConsumerNamePrefix     = "seldon-pipelinegateway"
	EnvModelGatewayMaxNumConsumers        = "MODELGATEWAY_MAX_NUM_CONSUMERS"
	EnvPipelineGatewayMaxNumConsumers     = "PIPELINEGATEWAY_MAX_NUM_CONSUMERS"
	DefaultMaxNumConsumers                = 100
)

var ErrAddServerEmptyServerName = status.Errorf(codes.FailedPrecondition, "Empty server name passed")

type SchedulerServer struct {
	pb.UnimplementedSchedulerServer
	health.UnimplementedHealthCheckServiceServer
	logger                 log.FieldLogger
	modelStore             store.ModelStore
	experimentServer       experiment.ExperimentServer
	pipelineHandler        pipeline.PipelineHandler
	scheduler              scheduler2.Scheduler
	modelEventStream       ModelEventStream
	serverEventStream      ServerEventStream
	experimentEventStream  ExperimentEventStream
	pipelineEventStream    PipelineEventStream
	controlPlaneStream     ControlPlaneStream
	timeout                time.Duration
	synchroniser           synchroniser.Synchroniser
	config                 SchedulerServerConfig
	modelGwLoadBalancer    *util.RingLoadBalancer
	pipelineGWLoadBalancer *util.RingLoadBalancer
	scalingConfigUpdates   chan scaling_config.ScalingConfig
	currentScalingConfig   *scaling_config.ScalingConfig
	mu                     sync.Mutex
	done                   chan struct{}
	grpcServer             *grpc.Server
	consumerGroupConfig    *ConsumerGroupConfig
	eventHub               *coordinator.EventHub
	tlsOptions             seldontls.TLSOptions
}

type SchedulerServerConfig struct {
	PackThreshold            float64
	AutoScalingServerEnabled bool
}

type ModelEventStream struct {
	mu                   sync.Mutex
	streams              map[pb.Scheduler_SubscribeModelStatusServer]*ModelSubscription
	conflictResolutioner *cr.ConflictResolutioner[store.ModelState]
}

type ServerEventStream struct {
	mu            sync.Mutex
	streams       map[pb.Scheduler_SubscribeServerStatusServer]*ServerSubscription
	batchWait     time.Duration
	trigger       *time.Timer
	pendingEvents map[string]struct{}
	pendingLock   sync.Mutex
}

type ExperimentEventStream struct {
	mu      sync.Mutex
	streams map[pb.Scheduler_SubscribeExperimentStatusServer]*ExperimentSubscription
}

type PipelineEventStream struct {
	mu                   sync.Mutex
	streams              map[pb.Scheduler_SubscribePipelineStatusServer]*PipelineSubscription
	namesToIps           map[string]string // Maps pipeline names to their IPs
	conflictResolutioner *cr.ConflictResolutioner[pipeline.PipelineStatus]
}

type ControlPlaneStream struct {
	mu      sync.Mutex
	streams map[pb.Scheduler_SubscribeControlPlaneServer]*ControlPlaneSubsription
}

type ModelSubscription struct {
	name           string
	stream         pb.Scheduler_SubscribeModelStatusServer
	fin            chan bool
	isModelGateway bool // Indicates if this subscription is for the model gateway
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
	name              string
	ip                string
	isPipelineGateway bool // Indicates if this subscription is for the pipeline gateway
	stream            pb.Scheduler_SubscribePipelineStatusServer
	fin               chan bool
}

type ControlPlaneSubsription struct {
	name   string
	stream pb.Scheduler_SubscribeControlPlaneServer
	fin    chan bool
}

type ConsumerGroupConfig struct {
	namespace                      string
	consumerGroupIdPrefix          string
	modelGatewayMaxNumConsumers    int
	pipelineGatewayMaxNumConsumers int
}

func NewConsumerGroupConfig(namespace, consumerGroupIdPrefix string, modelGatewayMaxNumConsumers int, pipelineGatewayMaxNumConsumers int) *ConsumerGroupConfig {
	if namespace == "" {
		namespace = "default"
	}
	if modelGatewayMaxNumConsumers <= 0 {
		modelGatewayMaxNumConsumers = DefaultMaxNumConsumers
	}
	if pipelineGatewayMaxNumConsumers <= 0 {
		pipelineGatewayMaxNumConsumers = DefaultMaxNumConsumers
	}
	return &ConsumerGroupConfig{
		namespace:                      namespace,
		consumerGroupIdPrefix:          consumerGroupIdPrefix,
		modelGatewayMaxNumConsumers:    modelGatewayMaxNumConsumers,
		pipelineGatewayMaxNumConsumers: pipelineGatewayMaxNumConsumers,
	}
}

func (s *SchedulerServer) startServer(ctx context.Context, port uint, secure bool, pollerTickCreate, pollerTickDelete time.Duration) error {
	logger := s.logger.WithField("func", "startServer")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	kaep := util.GetServerKeepAliveEnforcementPolicy()

	opts := []grpc.ServerOption{}
	if secure {
		opts = append(opts, grpc.Creds(s.tlsOptions.Cert.CreateServerTransportCredentials()))
	}
	opts = append(opts, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	opts = append(opts, grpc.StatsHandler(otelgrpc.NewServerHandler()))
	opts = append(opts, grpc.KeepaliveEnforcementPolicy(kaep))
	grpcServer := grpc.NewServer(opts...)
	s.grpcServer = grpcServer
	pb.RegisterSchedulerServer(grpcServer, s)
	health.RegisterHealthCheckServiceServer(grpcServer, s)

	s.logger.Printf("Scheduler server running on %d mtls:%v", port, secure)
	go func() {
		err := grpcServer.Serve(lis)
		logger.WithError(err).Fatalf("Scheduler mTLS server failed on port %d mtls:%v", port, secure)
	}()

	s.startPollers(ctx, pollerTickCreate, pollerTickDelete)
	return nil
}

func (s *SchedulerServer) startPollers(ctx context.Context, pollerTickCreate, pollerTickDelete time.Duration) {
	go s.pollerRetryFailedCreateModels(ctx, pollerTickCreate)
	go s.pollerRetryFailedDeleteModels(ctx, pollerTickDelete)
	go s.pollerRetryFailedCreatePipelines(ctx, pollerTickCreate)
	go s.pollerRetryFailedDeletePipelines(ctx, pollerTickDelete)
}

func (s *SchedulerServer) StartGrpcServers(ctx context.Context, allowPlainTxt bool, schedulerPort uint, schedulerTlsPort uint, pollerTickCreate, pollerTickDelete time.Duration) error {
	logger := s.logger.WithField("func", "StartGrpcServers")

	if !allowPlainTxt && s.tlsOptions.Cert == nil {
		return fmt.Errorf("one of plain txt or mTLS needs to be defined. But have plain text [%v] and no TLS", allowPlainTxt)
	}
	if allowPlainTxt {
		err := s.startServer(ctx, schedulerPort, false, pollerTickCreate, pollerTickDelete)
		if err != nil {
			return err
		}
	} else {
		logger.Info("Not starting scheduler plain text server")
	}
	if s.tlsOptions.Cert != nil {
		err := s.startServer(ctx, schedulerTlsPort, true, pollerTickCreate, pollerTickDelete)
		if err != nil {
			return err
		}
	} else {
		logger.Info("Not starting scheduler mTLS server")
	}
	return nil
}

func getEnVar(logger *log.Entry, key string, defaultValue int) int {
	valStr := os.Getenv(key)
	if valStr != "" {
		val, err := strconv.ParseInt(valStr, 10, 64)
		if err != nil {
			logger.WithError(err).Fatalf("Failed to parse %s", key)
		}
		logger.Infof("Got %s = %d", key, val)
		return int(val)
	}
	logger.Infof("Returning default %s = %d", key, defaultValue)
	return defaultValue
}

func NewSchedulerServer(
	logger log.FieldLogger,
	modelStore store.ModelStore,
	experiementServer experiment.ExperimentServer,
	pipelineHandler pipeline.PipelineHandler,
	scheduler scheduler2.Scheduler,
	eventHub *coordinator.EventHub,
	synchroniser synchroniser.Synchroniser,
	config SchedulerServerConfig,
	namespace string,
	consumerGroupIdPrefix string,
	modelGwLoadBalancer *util.RingLoadBalancer,
	pipelineGWLoadBalancer *util.RingLoadBalancer,
	scalingConfigHdl *scaling_config.ScalingConfigHandler,
	tlsOptions seldontls.TLSOptions,
) *SchedulerServer {
	loggerWithField := logger.WithField("source", "SchedulerServer")
	modelGatewayMaxNumConsumers := getEnVar(loggerWithField, EnvModelGatewayMaxNumConsumers, DefaultMaxNumConsumers)
	pipelineGatewayMaxNumConsumers := getEnVar(loggerWithField, EnvPipelineGatewayMaxNumConsumers, DefaultMaxNumConsumers)
	consumerGroupConfig := NewConsumerGroupConfig(
		namespace,
		consumerGroupIdPrefix,
		modelGatewayMaxNumConsumers,
		pipelineGatewayMaxNumConsumers,
	)

	s := &SchedulerServer{
		logger:           loggerWithField,
		modelStore:       modelStore,
		experimentServer: experiementServer,
		pipelineHandler:  pipelineHandler,
		scheduler:        scheduler,
		modelEventStream: ModelEventStream{
			streams:              make(map[pb.Scheduler_SubscribeModelStatusServer]*ModelSubscription),
			conflictResolutioner: cr.NewConflictResolution[store.ModelState](logger),
		},
		serverEventStream: ServerEventStream{
			streams:       make(map[pb.Scheduler_SubscribeServerStatusServer]*ServerSubscription),
			batchWait:     defaultBatchWait,
			trigger:       nil,
			pendingEvents: map[string]struct{}{},
		},
		pipelineEventStream: PipelineEventStream{
			streams:              make(map[pb.Scheduler_SubscribePipelineStatusServer]*PipelineSubscription),
			namesToIps:           make(map[string]string),
			conflictResolutioner: cr.NewConflictResolution[pipeline.PipelineStatus](loggerWithField),
		},
		experimentEventStream: ExperimentEventStream{
			streams: make(map[pb.Scheduler_SubscribeExperimentStatusServer]*ExperimentSubscription),
		},
		controlPlaneStream: ControlPlaneStream{
			streams: make(map[pb.Scheduler_SubscribeControlPlaneServer]*ControlPlaneSubsription),
		},
		timeout:                sendTimeout,
		synchroniser:           synchroniser,
		config:                 config,
		modelGwLoadBalancer:    modelGwLoadBalancer,
		pipelineGWLoadBalancer: pipelineGWLoadBalancer,
		scalingConfigUpdates:   make(chan scaling_config.ScalingConfig),
		done:                   make(chan struct{}),
		consumerGroupConfig:    consumerGroupConfig,
		eventHub:               eventHub,
		tlsOptions:             tlsOptions,
	}

	eventHub.RegisterModelEventHandler(
		modelEventHandlerName,
		pendingEventsQueueSize,
		s.logger,
		s.handleModelEvent,
	)
	eventHub.RegisterModelEventHandler(
		serverModelEventHandlerName,
		pendingEventsQueueSize,
		s.logger,
		s.handleModelEventForServerStatus,
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
	eventHub.RegisterServerEventHandler(
		serverEventHandlerName,
		pendingEventsQueueSize,
		s.logger,
		s.handleServerEvents,
	)

	if scalingConfigHdl != nil {
		initScalingConfig := scalingConfigHdl.GetConfiguration()
		s.currentScalingConfig = &initScalingConfig
		scalingConfigHdl.AddListener(s.scalingConfigUpdates)
		go s.handleScalingConfigChanges()
	} else {
		s.currentScalingConfig = &scaling_config.DefaultScalingConfig
	}

	s.mu.Lock()
	scaling_config.LogWhenUsingDefaultScalingConfig(s.currentScalingConfig, loggerWithField)
	s.mu.Unlock()

	return s
}

func (s *SchedulerServer) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
		s.logger.Info("Scheduler closing gRPC server managing connections from controller and gateways")
	}
	close(s.done)
}

func (s *SchedulerServer) handleScalingConfigChanges() {
	logger := s.logger.WithField("func", "handleScalingConfigChanges")
	for {
		select {
		case newScalingConfig := <-s.scalingConfigUpdates:
			if newScalingConfig.Pipelines == nil {
				continue
			}
			s.mu.Lock()
			if newScalingConfig.Pipelines.MaxShardCountMultiplier != s.currentScalingConfig.Pipelines.MaxShardCountMultiplier {
				logger.Info("Updating mapping of Pipelines and Models onto pipeline-gateway and model-gateway replicas following scaling config change")
				wg := sync.WaitGroup{}
				wg.Add(2)
				s.currentScalingConfig = &newScalingConfig
				scaling_config.LogWhenUsingDefaultScalingConfig(s.currentScalingConfig, logger)
				go func() {
					// lock Mutex to avoid updating load balancer if a concurrent rebalance is in progress
					s.pipelineEventStream.mu.Lock()
					s.pipelineGWLoadBalancer.UpdatePartitions(newScalingConfig.Pipelines.MaxShardCountMultiplier)
					s.pipelineEventStream.mu.Unlock()

					// There is a chance that another concurrent rebalance will start here (applying the
					// updated partitions), but it means we'll just do one extra rebalance that will
					// distribute the pipelines in the exact same way (given no other infra changes)
					// Given that config changes should be infrequent, this should be ok.

					// rebalance all pipelines onto available pipeline-gw replicas according to new config
					s.pipelineGwRebalance()
					wg.Done()
				}()

				go func() {
					s.modelEventStream.mu.Lock()
					s.modelGwLoadBalancer.UpdatePartitions(newScalingConfig.Pipelines.MaxShardCountMultiplier)
					s.modelEventStream.mu.Unlock()
					s.modelGwRebalance()
					wg.Done()
				}()
				wg.Wait()
			}
			s.mu.Unlock()
		case <-s.done:
			return
		}
	}
}

func (s *SchedulerServer) ServerNotify(ctx context.Context, req *pb.ServerNotifyRequest) (*pb.ServerNotifyResponse, error) {
	logger := s.logger.WithField("func", "ServerNotify")
	logger.Info("Received ServerNotify request", "req", req)

	// numExpectedReplicas is only used when we are doing the first sync
	numExpectedReplicas := uint(0)
	for _, server := range req.GetServers() {
		logger.Infof("Server notification %s expectedReplicas %d shared %v", server.GetName(), server.GetExpectedReplicas(), server.GetShared())
		err := s.modelStore.ServerNotify(server)
		if err != nil {
			return nil, status.Errorf(codes.FailedPrecondition, "%s", err.Error())
		}
		if server.ExpectedReplicas == 0 {
			go s.rescheduleModels(server.GetName())
		}
		numExpectedReplicas += uint(server.ExpectedReplicas)
	}
	if req.IsFirstSync && !s.synchroniser.IsReady() {
		s.synchroniser.Signals(numExpectedReplicas)
		logger.Infof("Signalling synchroniser with %d expected server agents to connect", numExpectedReplicas)
	}
	return &pb.ServerNotifyResponse{}, nil
}

func (s *SchedulerServer) HealthCheck(_ context.Context, _ *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	if s.eventHub.IsClosed() {
		return nil, errors.New("event hub closed")
	}
	return &health.HealthCheckResponse{Ok: true}, nil
}

func (s *SchedulerServer) rescheduleModels(serverKey string) {
	logger := s.logger.WithField("func", "rescheduleModels")
	server, err := s.modelStore.GetServer(serverKey, false, true)
	if err != nil {
		logger.WithError(err).Errorf("Failed to get server %s", serverKey)
		return
	}
	models := make(map[string]bool)
	for _, replica := range server.Replicas {
		for _, model := range replica.GetLoadedOrLoadingModelVersions() {
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
		return nil, status.Errorf(codes.FailedPrecondition, "%s", err.Error())
	}
	go func() {
		err := s.scheduler.Schedule(req.GetModel().GetMeta().GetName())
		if err != nil {
			logger.WithError(err).Warnf("Failed to schedule model %s", req.GetModel().GetMeta().GetName())
		}
	}()
	return &pb.LoadModelResponse{}, nil
}

func (s *SchedulerServer) UnloadModel(ctx context.Context, req *pb.UnloadModelRequest) (*pb.UnloadModelResponse, error) {
	logger := s.logger.WithField("func", "UnloadModel")
	logger.Debugf("Unload model %s", req.GetModel().Name)
	err := s.modelStore.RemoveModel(req)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "%s", err.Error())
	}
	go func() {
		err := s.scheduler.Schedule(req.GetModel().Name)
		if err != nil {
			logger.WithError(err).Warnf("Failed to schedule model %s (for unload)", req.GetModel().GetName())
		}
	}()
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
			ModelGwState:        pb.ModelStatus_ModelState(pb.ModelStatus_ModelState_value[modelState.ModelGwState.String()]),
			Reason:              modelState.Reason,
			ModelGwReason:       modelState.ModelGwReason,
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
			return nil, status.Errorf(codes.FailedPrecondition, "Failed to find model %s", model.Name)
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
			return status.Errorf(codes.FailedPrecondition, "%s", err.Error())
		}

		for _, m := range models {
			resp, err := s.modelStatusImpl(m, req.AllVersions)
			if err != nil {
				return status.Errorf(codes.FailedPrecondition, "%s", err.Error())
			}

			select {
			case <-stream.Context().Done():
				return status.Errorf(codes.Canceled, "stream ctx cancelled: %s", stream.Context().Err())
			default:
				err = stream.Send(resp)
				if err != nil {
					return status.Errorf(codes.Internal, "%s", err.Error())
				}
			}
		}
		return nil
	} else {
		// Single model requested
		s.modelStore.LockModel(req.Model.Name)
		defer s.modelStore.UnlockModel(req.Model.Name)

		model, err := s.modelStore.GetModel(req.Model.Name)
		if err != nil {
			return status.Errorf(codes.FailedPrecondition, "%s", err.Error())
		}
		if model == nil || len(model.Versions) == 0 {
			return status.Errorf(
				codes.FailedPrecondition,
				"Failed to find model %s", req.Model.Name,
			)
		}

		resp, err := s.modelStatusImpl(model, req.AllVersions)
		if err != nil {
			return status.Errorf(codes.Internal, "%s", err.Error())
		}

		select {
		case <-stream.Context().Done():
			return status.Errorf(codes.Canceled, "stream ctx cancelled: %s", stream.Context().Err())
		default:
			err = stream.Send(resp)
			if err != nil {
				return status.Errorf(codes.Internal, "%s", err.Error())
			}
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
		servers, err := s.modelStore.GetServers(true, true)
		if err != nil {
			return status.Errorf(codes.FailedPrecondition, "%s", err.Error())
		}

		for _, s := range servers {
			resp := createServerStatusUpdateResponse(s)
			select {
			case <-stream.Context().Done():
				return status.Errorf(codes.Canceled, "stream ctx cancelled: %s", stream.Context().Err())
			default:
				err := stream.Send(resp)
				if err != nil {
					return status.Errorf(codes.Internal, "%s", err.Error())
				}
			}
		}
		return nil
	} else {
		// Single server requested
		server, err := s.modelStore.GetServer(req.GetName(), true, true)
		if err != nil {
			return status.Errorf(codes.FailedPrecondition, "%s", err.Error())
		}
		resp := createServerStatusUpdateResponse(server)

		select {
		case <-stream.Context().Done():
			return status.Errorf(codes.Canceled, "stream ctx cancelled: %s", stream.Context().Err())
		default:
			err = stream.Send(resp)
			if err != nil {
				return status.Errorf(codes.Internal, "%s", err.Error())
			}
		}

		return nil
	}
}

func createServerStatusUpdateResponse(s *store.ServerSnapshot) *pb.ServerStatusResponse {
	// note we dont count draining replicas in available replicas

	resp := &pb.ServerStatusResponse{
		Type:             pb.ServerStatusResponse_StatusUpdate,
		ServerName:       s.Name,
		ExpectedReplicas: int32(s.ExpectedReplicas),
		KubernetesMeta:   s.KubernetesMeta,
	}

	totalModels := int32(0)
	numAvailableServerReplicas := int32(0)
	for _, replica := range s.Replicas {
		if !replica.GetIsDraining() {
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
			totalModels += numLoadedModelsOnReplica
			numAvailableServerReplicas++
		}
	}
	resp.NumLoadedModelReplicas = totalModels
	resp.AvailableReplicas = numAvailableServerReplicas

	return resp
}

func createServerScaleResponse(s *store.ServerSnapshot, expectedReplicas uint32) *pb.ServerStatusResponse {
	// we dont care about populating the other fields as they should not be used by the controller, reconsider if this changes

	resp := &pb.ServerStatusResponse{
		Type:             pb.ServerStatusResponse_ScalingRequest,
		ServerName:       s.Name,
		ExpectedReplicas: int32(expectedReplicas),
		KubernetesMeta:   s.KubernetesMeta,
	}

	return resp
}

func (s *SchedulerServer) StartExperiment(ctx context.Context, req *pb.StartExperimentRequest) (*pb.StartExperimentResponse, error) {
	err := s.experimentServer.StartExperiment(experiment.CreateExperimentFromRequest(req.Experiment))
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "%s", err.Error())
	}
	return &pb.StartExperimentResponse{}, nil
}

func (s *SchedulerServer) StopExperiment(ctx context.Context, req *pb.StopExperimentRequest) (*pb.StopExperimentResponse, error) {
	err := s.experimentServer.StopExperiment(req.GetName())
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "%s", err.Error())
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
			return status.Errorf(codes.FailedPrecondition, "%s", err.Error())
		}

		for _, e := range experiments {
			resp := createExperimentStatus(e)
			select {
			case <-stream.Context().Done():
				return status.Errorf(codes.Canceled, "stream ctx cancelled: %s", stream.Context().Err())
			default:
				err = stream.Send(resp)
				if err != nil {
					return status.Errorf(codes.Internal, "%s", err.Error())
				}
			}

		}
		return nil
	} else {
		// Single experiment requested
		exp, err := s.experimentServer.GetExperiment(req.GetName())
		if err != nil {
			return status.Errorf(codes.FailedPrecondition, "%s", err.Error())
		}

		resp := createExperimentStatus(exp)

		select {
		case <-stream.Context().Done():
			return status.Errorf(codes.Canceled, "stream ctx cancelled: %s", stream.Context().Err())
		default:
			err = stream.Send(resp)
			if err != nil {
				return status.Errorf(codes.Internal, "%s", err.Error())
			}
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
		return nil, status.Errorf(codes.FailedPrecondition, "%s", err.Error())
	}
	return &pb.LoadPipelineResponse{}, nil
}

func (s *SchedulerServer) UnloadPipeline(ctx context.Context, req *pb.UnloadPipelineRequest) (*pb.UnloadPipelineResponse, error) {
	err := s.pipelineHandler.RemovePipeline(req.GetName())
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "%s", err.Error())
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
		return s.sendCurrentPipelineStatuses(stream, req.AllVersions)
	} else {
		// Single pipeline requested
		p, err := s.pipelineHandler.GetPipeline(req.GetName())
		if err != nil {
			return status.Errorf(codes.FailedPrecondition, "%s", err.Error())
		}
		resp := createPipelineStatus(p, req.GetAllVersions())

		select {
		case <-stream.Context().Done():
			return status.Errorf(codes.Canceled, "stream ctx cancelled: %s", stream.Context().Err())
		default:
			err = stream.Send(resp)
			if err != nil {
				return status.Errorf(codes.Internal, "%s", err.Error())
			}
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
