/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package dataflow

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/chainer"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/health"
	kafka_config "github.com/seldonio/seldon-core/components/kafka/v2/pkg/config"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka"
	cr "github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/conflict-resolution"
	scaling_config "github.com/seldonio/seldon-core/scheduler/v2/pkg/scaling/config"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

const (
	grpcMaxConcurrentStreams     = 1_000_000
	pipelineEventHandlerName     = "kafka.dataflow.server.pipelines"
	pendingEventsQueueSize   int = 1000
)

type ChainerServer struct {
	logger               log.FieldLogger
	mu                   sync.Mutex
	streams              map[string]*ChainerSubscription
	eventHub             *coordinator.EventHub
	pipelineHandler      pipeline.PipelineHandler
	topicNamer           *kafka.TopicNamer
	loadBalancer         util.LoadBalancer
	conflictResolutioner *cr.ConflictResolutioner[pipeline.PipelineStatus]
	chainerMutex         sync.Map
	configUpdatesMutex   sync.Mutex
	scalingConfigUpdates chan scaling_config.ScalingConfig
	currentScalingConfig scaling_config.ScalingConfig
	done                 chan struct{}
	grpcServer           *grpc.Server
	muFailedCreate       sync.Mutex
	// TODO we should update PipelineHandler to store state for dataflow-engine, as we do for model-gw. That way we
	//  won't have to use failedCreatePipelines and failedDeletePipelines.
	// failedCreatePipelines keyed off pipeline UID + version
	failedCreatePipelines map[string]pipeline.PipelineVersion
	muFailedDelete        sync.Mutex
	// failedDeletePipelines keyed off pipeline UID + version
	failedDeletePipelines    map[string]pipeline.PipelineVersion
	muRetriedFailedPipelines sync.Mutex
	// retriedFailedPipelines keyed off pipeline UID + version. Tracks how many attempts have been made to create/terminate
	// a pipeline
	retriedFailedPipelines map[string]uint
	chainer.UnimplementedChainerServer
	health.UnimplementedHealthCheckServiceServer
}

type ChainerSubscription struct {
	name   string
	stream chainer.Chainer_SubscribePipelineUpdatesServer
	fin    chan bool
}

func NewChainerServer(
	logger log.FieldLogger,
	eventHub *coordinator.EventHub,
	pipelineHandler pipeline.PipelineHandler,
	namespace string,
	loadBalancer util.LoadBalancer,
	kafkaConfig *kafka_config.KafkaConfig,
	scalingConfigHdl *scaling_config.ScalingConfigHandler,
) (*ChainerServer, error) {
	conflictResolutioner := cr.NewConflictResolution[pipeline.PipelineStatus](logger)
	topicNamer, err := kafka.NewTopicNamer(namespace, kafkaConfig.TopicPrefix)
	if err != nil {
		return nil, err
	}
	c := &ChainerServer{
		logger:                 logger.WithField("source", "dataflow"),
		streams:                make(map[string]*ChainerSubscription),
		eventHub:               eventHub,
		pipelineHandler:        pipelineHandler,
		topicNamer:             topicNamer,
		loadBalancer:           loadBalancer,
		conflictResolutioner:   conflictResolutioner,
		chainerMutex:           sync.Map{},
		scalingConfigUpdates:   make(chan scaling_config.ScalingConfig),
		done:                   make(chan struct{}),
		failedCreatePipelines:  make(map[string]pipeline.PipelineVersion, 0),
		failedDeletePipelines:  make(map[string]pipeline.PipelineVersion, 0),
		retriedFailedPipelines: make(map[string]uint, 0),
	}

	eventHub.RegisterPipelineEventHandler(
		pipelineEventHandlerName,
		pendingEventsQueueSize,
		c.logger,
		c.handlePipelineEvent,
	)

	if scalingConfigHdl != nil {
		c.currentScalingConfig = scalingConfigHdl.GetConfiguration()
		scalingConfigHdl.AddListener(c.scalingConfigUpdates)
		go c.handleScalingConfigChanges()
	} else {
		c.currentScalingConfig = scaling_config.DefaultScalingConfig
	}

	c.configUpdatesMutex.Lock()
	scaling_config.LogWhenUsingDefaultScalingConfig(&c.currentScalingConfig, c.logger)
	c.configUpdatesMutex.Unlock()

	return c, nil
}

func (c *ChainerServer) Stop() {
	if c.grpcServer != nil {
		c.grpcServer.GracefulStop()
		c.logger.Info("Scheduler closing gRPC server managing connections from dataflow-engine replicas")
	}
	c.logger.Info("Stop watching for scaling config changes")
	close(c.done)
	c.StopSendPipelineEvents()
}

func (c *ChainerServer) StartGrpcServer(ctx context.Context, pollerFailedCreatePipelines, pollerFailedDeletePipelines time.Duration, maxRetry, agentPort uint) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", agentPort))
	if err != nil {
		log.Fatalf("failed to create listener: %v", err)
	}

	go c.pollerFailedTerminatingPipelines(ctx, pollerFailedDeletePipelines, maxRetry)
	go c.pollerFailedCreatingPipelines(ctx, pollerFailedCreatePipelines, maxRetry)

	kaep := util.GetServerKeepAliveEnforcementPolicy()

	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	grpcOptions = append(grpcOptions, grpc.KeepaliveEnforcementPolicy(kaep))
	grpcServer := grpc.NewServer(grpcOptions...)
	chainer.RegisterChainerServer(grpcServer, c)
	health.RegisterHealthCheckServiceServer(grpcServer, c)
	c.grpcServer = grpcServer

	c.logger.Printf("Chainer server running on %d", agentPort)
	return grpcServer.Serve(lis)
}

func (c *ChainerServer) mkPipelineRetryKey(uid string, version uint32) string {
	return fmt.Sprintf("%s_%d", uid, version)
}

func (c *ChainerServer) storeFailedCreate(m *chainer.PipelineUpdateMessage) {
	c.muFailedCreate.Lock()
	defer c.muFailedCreate.Unlock()
	c.failedCreatePipelines[c.mkPipelineRetryKey(m.Uid, m.Version)] = pipeline.PipelineVersion{
		Name:    m.Pipeline,
		Version: m.Version,
		UID:     m.Uid,
	}
}

func (c *ChainerServer) storeFailedDelete(m *chainer.PipelineUpdateMessage) {
	c.muFailedDelete.Lock()
	defer c.muFailedDelete.Unlock()
	c.failedDeletePipelines[c.mkPipelineRetryKey(m.Uid, m.Version)] = pipeline.PipelineVersion{
		Name:    m.Pipeline,
		Version: m.Version,
		UID:     m.Uid,
	}
}

func (c *ChainerServer) resetPipelineRetryCount(msg *chainer.PipelineUpdateMessage) {
	c.muRetriedFailedPipelines.Lock()
	defer c.muRetriedFailedPipelines.Unlock()
	c.retriedFailedPipelines[c.mkPipelineRetryKey(msg.Uid, msg.Version)] = 0
}

func (c *ChainerServer) PipelineUpdateEvent(ctx context.Context, message *chainer.PipelineUpdateStatusMessage) (*chainer.PipelineUpdateStatusResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	logger := c.logger.WithField("func", "PipelineUpdateEvent")
	var statusVal pipeline.PipelineStatus

	switch message.Update.Op {
	// create, delete, rebalance operation from the scheduler
	case chainer.PipelineUpdateMessage_Create:
		if message.Success {
			c.resetPipelineRetryCount(message.Update)
			statusVal = pipeline.PipelineReady
		} else {
			c.storeFailedCreate(message.Update)
			statusVal = pipeline.PipelineFailed
		}
	case chainer.PipelineUpdateMessage_Delete:
		if message.Success {
			c.resetPipelineRetryCount(message.Update)
			statusVal = pipeline.PipelineTerminated
		} else {
			c.storeFailedDelete(message.Update)
			statusVal = pipeline.PipelineFailedTerminating
		}
	// internal rebalancing operation
	case chainer.PipelineUpdateMessage_Rebalance:
		if message.Success {
			statusVal = pipeline.PipelineRebalancing
		} else {
			statusVal = pipeline.PipelineFailed
		}
	case chainer.PipelineUpdateMessage_Ready:
		if message.Success {
			statusVal = pipeline.PipelineReady
		} else {
			statusVal = pipeline.PipelineFailed
		}
	}

	pipelineName := message.Update.Pipeline
	pipelineVersion := message.Update.Version
	stream := message.Update.Stream
	logger.Debugf(
		"Received pipeline update event from %s for pipeline %s:%d with status %s",
		stream, pipelineName, pipelineVersion, statusVal.String(),
	)

	if cr.IsPipelineMessageOutdated(c.conflictResolutioner, message) {
		// Maybe in the future we can process the outdated message in case of an error
		logger.Debugf("Message for pipeline %s:%d is outdated, ignoring", pipelineName, pipelineVersion)
		return &chainer.PipelineUpdateStatusResponse{}, nil
	}

	c.conflictResolutioner.UpdateStatus(pipelineName, stream, statusVal)
	pipelineStatusVal, reason := cr.GetPipelineStatus(c.conflictResolutioner, pipelineName, message)
	if pipelineStatusVal == pipeline.PipelineTerminated {
		c.conflictResolutioner.Delete(pipelineName)
	}

	err := c.pipelineHandler.SetPipelineState(message.Update.Pipeline, message.Update.Version, message.Update.Uid, pipelineStatusVal, reason, util.SourceChainerServer)
	if err != nil {
		logger.WithError(err).Errorf("Failed to update pipeline status for %s:%d (%s)", message.Update.Pipeline, message.Update.Version, message.Update.Uid)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &chainer.PipelineUpdateStatusResponse{}, nil
}

func (c *ChainerServer) HealthCheck(_ context.Context, _ *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	if c.eventHub.IsClosed() {
		return nil, errors.New("event hub closed")
	}
	return &health.HealthCheckResponse{Ok: true}, nil
}

func (c *ChainerServer) SubscribePipelineUpdates(req *chainer.PipelineSubscriptionRequest, stream chainer.Chainer_SubscribePipelineUpdatesServer) error {
	logger := c.logger.WithField("func", "SubscribePipelineStatus")

	key := req.GetName()
	// this is forcing a serial order per dataflow-engine
	// in general this will make sure that a given dataflow-engine disconnects fully before another dataflow-engine is allowed to connect
	mu, _ := c.chainerMutex.LoadOrStore(key, &sync.Mutex{})
	mu.(*sync.Mutex).Lock()
	defer mu.(*sync.Mutex).Unlock()

	logger.Infof("Received pipeline updates subscribe request from %s", req.GetName())

	fin := make(chan bool)

	c.mu.Lock()
	c.streams[key] = &ChainerSubscription{
		name:   key,
		stream: stream,
		fin:    fin,
	}
	c.loadBalancer.AddServer(key)
	c.mu.Unlock()

	// Handle addition of new server
	//TODO delay this for x seconds after server start and do 1 rebalance after clients have joined
	// otherwise first server may be overloaded with all pipelines in system
	c.rebalance()

	ctx := stream.Context()
	// Keep this scope alive because once this scope exits - the stream is closed

	select {
	case <-fin:
		logger.Infof("Closing stream for %s", key)
	case <-ctx.Done():
		logger.Infof("Stream disconnected %s", key)
		c.mu.Lock()
		c.loadBalancer.RemoveServer(key)
		delete(c.streams, key)
		c.mu.Unlock()
		// Handle removal of server
		c.rebalance()
	}

	return nil
}

func (c *ChainerServer) StopSendPipelineEvents() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, subscription := range c.streams {
		close(subscription.fin)
	}
}

// Create the topics for pipeline inputs
// This will be pipeline inputs/outputs or model topics
func (c *ChainerServer) createPipelineTopicSources(inputs []string) []*chainer.PipelineTopic {
	var pipelineTopics []*chainer.PipelineTopic
	for _, inp := range inputs {
		// The pipeline name being referred to by the input specification
		pipelineName := c.topicNamer.GetPipelineNameFromInput(inp)
		inpReference := c.topicNamer.CreateStepReferenceFromPipelineInput(inp)
		// The topic being referred to: either pipeline or model
		source, tensor := c.topicNamer.GetModelOrPipelineTopicAndTensor(pipelineName, inpReference)
		pipelineTopics = append(pipelineTopics, &chainer.PipelineTopic{PipelineName: pipelineName, TopicName: source, Tensor: tensor})
	}
	return pipelineTopics
}

func (c *ChainerServer) createTopicSources(inputs []string, pipelineName string) []*chainer.PipelineTopic {
	var pipelineTopics []*chainer.PipelineTopic
	for _, inp := range inputs {
		source, tensor := c.topicNamer.GetModelOrPipelineTopicAndTensor(pipelineName, inp)
		pipelineTopics = append(pipelineTopics, &chainer.PipelineTopic{PipelineName: pipelineName, TopicName: source, Tensor: tensor})
	}
	if len(pipelineTopics) == 0 {
		pipelineTopics = append(pipelineTopics, &chainer.PipelineTopic{PipelineName: pipelineName, TopicName: c.topicNamer.GetPipelineTopicInputs(pipelineName), Tensor: nil})
	}
	return pipelineTopics
}

func (c *ChainerServer) createTriggerSources(inputs []string, pipelineName string) []*chainer.PipelineTopic {
	var sources []*chainer.PipelineTopic
	for _, inp := range inputs {
		source, tensor := c.topicNamer.GetModelOrPipelineTopicAndTensor(pipelineName, inp)
		sources = append(sources, &chainer.PipelineTopic{PipelineName: pipelineName, TopicName: source, Tensor: tensor})
	}
	return sources
}

func (c *ChainerServer) createInputStepUpdate(pv *pipeline.PipelineVersion) *chainer.PipelineStepUpdate {
	stepUpdate := chainer.PipelineStepUpdate{
		Sources:      c.createPipelineTopicSources(pv.Input.ExternalInputs),
		Sink:         &chainer.PipelineTopic{PipelineName: pv.Name, TopicName: c.topicNamer.GetPipelineTopicInputs(pv.Name), Tensor: nil},
		Triggers:     c.createPipelineTopicSources(pv.Input.ExternalTriggers),
		TensorMap:    c.topicNamer.GetFullyQualifiedPipelineTensorMap(pv.Input.TensorMap),
		JoinWindowMs: pv.Input.JoinWindowMs,
	}
	switch pv.Input.InputsJoinType {
	case pipeline.JoinInner:
		stepUpdate.InputJoinTy = chainer.PipelineStepUpdate_Inner
	case pipeline.JoinOuter:
		stepUpdate.InputJoinTy = chainer.PipelineStepUpdate_Outer
	case pipeline.JoinAny:
		stepUpdate.InputJoinTy = chainer.PipelineStepUpdate_Any
	}
	switch pv.Input.TriggersJoinType {
	case pipeline.JoinInner:
		stepUpdate.TriggersJoinTy = chainer.PipelineStepUpdate_Inner
	case pipeline.JoinOuter:
		stepUpdate.TriggersJoinTy = chainer.PipelineStepUpdate_Outer
	case pipeline.JoinAny:
		stepUpdate.TriggersJoinTy = chainer.PipelineStepUpdate_Any
	}
	c.logger.Infof("Adding input sources %v with tensorMap %v to %s", stepUpdate.Sources, stepUpdate.TensorMap, stepUpdate.Sink)
	return &stepUpdate
}

func (c *ChainerServer) createOutputStepUpdate(pv *pipeline.PipelineVersion) *chainer.PipelineStepUpdate {
	stepUpdate := chainer.PipelineStepUpdate{
		Sources:      c.createTopicSources(pv.Output.Steps, pv.Name),
		Sink:         &chainer.PipelineTopic{PipelineName: pv.Name, TopicName: c.topicNamer.GetPipelineTopicOutputs(pv.Name), Tensor: nil},
		TensorMap:    c.topicNamer.GetFullyQualifiedTensorMap(pv.Name, pv.Output.TensorMap),
		JoinWindowMs: &pv.Output.JoinWindowMs,
	}
	switch pv.Output.StepsJoinType {
	case pipeline.JoinInner:
		stepUpdate.InputJoinTy = chainer.PipelineStepUpdate_Inner
	case pipeline.JoinOuter:
		stepUpdate.InputJoinTy = chainer.PipelineStepUpdate_Outer
	case pipeline.JoinAny:
		stepUpdate.InputJoinTy = chainer.PipelineStepUpdate_Any
	}
	c.logger.Infof("Adding sources %v to %s", stepUpdate.Sources, stepUpdate.Sink)
	return &stepUpdate
}

func (c *ChainerServer) createStepUpdate(pv *pipeline.PipelineVersion, step *pipeline.PipelineStep) *chainer.PipelineStepUpdate {
	stepUpdate := chainer.PipelineStepUpdate{
		Sources:      c.createTopicSources(step.Inputs, pv.Name),
		Triggers:     c.createTriggerSources(step.Triggers, pv.Name),
		Sink:         &chainer.PipelineTopic{PipelineName: pv.Name, TopicName: c.topicNamer.GetModelTopicInputs(step.Name), Tensor: nil},
		TensorMap:    c.topicNamer.GetFullyQualifiedTensorMap(pv.Name, step.TensorMap),
		JoinWindowMs: step.JoinWindowMs,
	}
	switch step.InputsJoinType {
	case pipeline.JoinInner:
		stepUpdate.InputJoinTy = chainer.PipelineStepUpdate_Inner
	case pipeline.JoinOuter:
		stepUpdate.InputJoinTy = chainer.PipelineStepUpdate_Outer
	case pipeline.JoinAny:
		stepUpdate.InputJoinTy = chainer.PipelineStepUpdate_Any
	}
	switch step.TriggersJoinType {
	case pipeline.JoinInner:
		stepUpdate.TriggersJoinTy = chainer.PipelineStepUpdate_Inner
	case pipeline.JoinOuter:
		stepUpdate.TriggersJoinTy = chainer.PipelineStepUpdate_Outer
	case pipeline.JoinAny:
		stepUpdate.TriggersJoinTy = chainer.PipelineStepUpdate_Any
	}
	if step.Batch != nil {
		stepUpdate.Batch = &chainer.Batch{
			Size:     step.Batch.Size,
			WindowMs: step.Batch.WindowMs,
		}
	}
	c.logger.Infof("Adding sources %v to %s", stepUpdate.Sources, stepUpdate.Sink)
	return &stepUpdate
}

func (c *ChainerServer) createPipelineCreationMessage(pv *pipeline.PipelineVersion) *chainer.PipelineUpdateMessage {
	var stepUpdates []*chainer.PipelineStepUpdate
	for _, step := range pv.Steps {
		stepUpdates = append(stepUpdates, c.createStepUpdate(pv, step))
	}
	if pv.Input != nil {
		stepUpdates = append(stepUpdates, c.createInputStepUpdate(pv))
	}
	if pv.Output != nil {
		stepUpdates = append(stepUpdates, c.createOutputStepUpdate(pv))
	}
	//Append an error step to send any errors to pipeline output
	stepUpdates = append(stepUpdates, &chainer.PipelineStepUpdate{
		Sources:     []*chainer.PipelineTopic{{PipelineName: pv.Name, TopicName: c.topicNamer.GetModelErrorTopic(), Tensor: nil}},
		Sink:        &chainer.PipelineTopic{PipelineName: pv.Name, TopicName: c.topicNamer.GetPipelineTopicOutputs(pv.Name), Tensor: nil},
		InputJoinTy: chainer.PipelineStepUpdate_Inner,
	})
	return &chainer.PipelineUpdateMessage{
		Pipeline:            pv.Name,
		Version:             pv.Version,
		Uid:                 pv.UID,
		Updates:             stepUpdates,
		Op:                  chainer.PipelineUpdateMessage_Create,
		PipelineOutputTopic: c.topicNamer.GetPipelineTopicOutputs(pv.Name),
		PipelineErrorTopic:  c.topicNamer.GetModelErrorTopic(),
		AllowCycles:         pv.AllowCycles,
		MaxStepRevisits:     pv.MaxStepRevisits,
	}
}

func (c *ChainerServer) createPipelineDeletionMessage(pv *pipeline.PipelineVersion, keepTopics bool) *chainer.PipelineUpdateMessage {
	message := chainer.PipelineUpdateMessage{
		Pipeline: pv.Name,
		Version:  pv.Version,
		Uid:      pv.UID,
		Op:       chainer.PipelineUpdateMessage_Delete,
	}
	if !keepTopics && pv.DataflowSepec != nil && pv.DataflowSepec.CleanTopicsOnDelete {
		// Both topics are always created. The input topic is created
		// when creating the topics for each step. The output topic
		// is created when creating the error topic.
		message.Updates = []*chainer.PipelineStepUpdate{
			{
				Sources: []*chainer.PipelineTopic{{PipelineName: pv.Name, TopicName: c.topicNamer.GetPipelineTopicInputs(pv.Name), Tensor: nil}},
				Sink:    &chainer.PipelineTopic{PipelineName: pv.Name, TopicName: c.topicNamer.GetPipelineTopicOutputs(pv.Name), Tensor: nil},
			},
		}
	}
	return &message
}

func (c *ChainerServer) sendPipelineMsgToSelectedServers(msg *chainer.PipelineUpdateMessage, pv *pipeline.PipelineVersion) {
	logger := c.logger.WithField("func", "sendPipelineMsg")
	servers := c.loadBalancer.GetServersForKey(pv.UID)

	cr.CreateNewPipelineIteration(c.conflictResolutioner, pv.Name, servers)
	msg.Timestamp = c.conflictResolutioner.VectorClock[pv.Name]

	for _, serverId := range servers {
		if subscription, ok := c.streams[serverId]; ok {
			select {
			case <-subscription.stream.Context().Done():
				logger.WithError(subscription.stream.Context().Err()).Errorf("Failed to send msg to pipeline %s - stream ctx cancelled", pv.String())
			default:
				if err := subscription.stream.Send(msg); err != nil {
					logger.WithError(err).Errorf("Failed to send msg to pipeline %s", pv.String())
				}
			}
		} else {
			logger.Errorf("Failed to get pipeline subscription with key %s", serverId)
		}
	}
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func (c *ChainerServer) pollerFailedTerminatingPipelines(ctx context.Context, tick time.Duration, maxRetry uint) {
	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	logger := c.logger.WithField("func", "pollerFailedTerminatingPipelines")
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// check for any pipelines which failed to create and retry
			logger.Debug("Checking for pipelines which failed to terminate")
			c.muFailedDelete.Lock()
			if len(c.failedDeletePipelines) == 0 {
				c.muFailedDelete.Unlock()
				logger.Debug("No pipelines found that failed to terminate")
				continue
			}

			c.mu.Lock()
			for _, p := range c.failedDeletePipelines {
				key := c.mkPipelineRetryKey(p.UID, p.Version)
				c.muRetriedFailedPipelines.Lock()
				c.retriedFailedPipelines[key]++

				if c.retriedFailedPipelines[key] > maxRetry {
					c.muRetriedFailedPipelines.Unlock()
					logger.Warnf("Failed to terminate pipeline %s reached max retries", p.Name)
					delete(c.failedDeletePipelines, key)
					continue
				}
				c.muRetriedFailedPipelines.Unlock()

				logger.Debugf("Attempting to terminate pipeline which failed to terminate %s", p.Name)
				pv, err := c.pipelineHandler.GetPipelineVersion(p.Name, p.Version, p.UID)
				if err != nil {
					notFound := &pipeline.PipelineNotFoundErr{}
					uidMisMatch := &pipeline.PipelineVersionUidMismatchErr{}
					verNotFound := &pipeline.PipelineVersionNotFoundErr{}

					if errors.As(err, &notFound) || errors.As(err, &uidMisMatch) || errors.As(err, &verNotFound) {
						delete(c.failedDeletePipelines, key)
						logger.Debugf("Pipeline %s not found, removing from poller list", p.Name)
						continue
					}

					logger.WithError(err).Errorf("Failed to get pipeline %s", p.Name)
					continue
				}
				logger.Debugf("Found pipeline %s attempting to terminate", p.Name)

				// note we are forcing keeping topics here, so there may be unwanted orphaned topics left in Kafka even
				// though customer deleted pipeline and set pipeline config to delete topics. This is because we don't
				// know if the termination request was initiated by customer i.e. deleted Pipeline CR, or from a rebalance
				// of pipelines across dataflow-engine replicas (in which case we force keeping topics)
				c.terminatePipeline(pv, true)
				// remove from list as we've successfully retried termination
				delete(c.failedDeletePipelines, key)
			}

			c.mu.Unlock()
			c.muFailedDelete.Unlock()
		}
	}
}

func (c *ChainerServer) pollerFailedCreatingPipelines(ctx context.Context, tick time.Duration, maxRetry uint) {
	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	logger := c.logger.WithField("func", "pollerFailedCreatingPipelines")
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// check for any pipelines which failed to create and retry
			logger.Debug("Checking for pipelines which failed to create")
			c.muFailedCreate.Lock()
			if len(c.failedCreatePipelines) == 0 {
				c.muFailedCreate.Unlock()
				logger.Debug("No pipelines found that failed to create")
				continue
			}

			c.mu.Lock()
			for _, p := range c.failedCreatePipelines {
				key := c.mkPipelineRetryKey(p.UID, p.Version)
				c.muRetriedFailedPipelines.Lock()
				c.retriedFailedPipelines[key]++
				logger.Debugf("Attempting to create failed pipeline %s", p.Name)

				if c.retriedFailedPipelines[key] > maxRetry {
					c.muRetriedFailedPipelines.Unlock()
					logger.Warnf("Failed to create pipeline %s reached max retries", p.Name)
					delete(c.failedCreatePipelines, key)
					continue
				}
				c.muRetriedFailedPipelines.Unlock()

				// we only want to create this pipeline if it's the latest version, it could have failed to create
				// and customer has since updated the pipeline and has created successfully, we'd then end up
				// overwriting the new pipeline
				if isLatest, err := c.pipelineHandler.IsLatestVersion(p.Name, p.Version, p.UID); err != nil {
					logger.WithError(err).Errorf("Failed checking pipeline %s is latest version before creating", p.Name)
					delete(c.failedCreatePipelines, key)
					continue
				} else if !isLatest {
					logger.Debugf("Pipeline %s not the latest, ignoring", p.Name)
					delete(c.failedCreatePipelines, key)
					continue
				}

				if err := c.rebalancePipeline(p.Name, p.Version, p.UID); err != nil {
					notFound := &pipeline.PipelineNotFoundErr{}
					uidMisMatch := &pipeline.PipelineVersionUidMismatchErr{}
					verNotFound := &pipeline.PipelineVersionNotFoundErr{}

					if errors.As(err, &notFound) || errors.As(err, &uidMisMatch) || errors.As(err, &verNotFound) {
						delete(c.failedCreatePipelines, key)
						logger.Debugf("Pipeline %s not found, removing from poller list", p.Name)
						continue
					}

					// don't remove from map as we want to retry on next tick
					logger.WithError(err).Errorf("Failed to create pipeline %s", p.Name)
					continue
				}

				// remove from list as we've successfully retried creating
				delete(c.failedCreatePipelines, key)
			}

			c.mu.Unlock()
			c.muFailedCreate.Unlock()
		}
	}
}

func (c *ChainerServer) rebalancePipeline(pipelineName string, pipelineVersion uint32, pipelineUID string) error {
	logger := c.logger.WithField("func", "rebalance")
	pv, err := c.pipelineHandler.GetPipelineVersion(pipelineName, pipelineVersion, pipelineUID)
	if err != nil {
		logger.WithError(err).Errorf("Failed to get pipeline %s UID %d version %s", pipelineName, pipelineVersion, pipelineUID)
		return err
	}

	c.logger.Debugf("Rebalancing pipeline %s:%d with state %s", pipelineName, pipelineVersion, pv.State.Status.String())
	if len(c.streams) == 0 {
		pipelineState := pipeline.PipelineCreate
		// if no dataflow engines available then we think we can terminate pipelines.
		if pv.State.Status == pipeline.PipelineTerminating {
			pipelineState = pipeline.PipelineTerminated
		}

		c.logger.Debugf("No dataflow engines available to handle pipeline %s, setting state to %s", pv.String(), pipelineState.String())
		if err := c.pipelineHandler.SetPipelineState(
			pv.Name,
			pv.Version,
			pv.UID,
			pipelineState,
			"no dataflow engines available to handle pipeline",
			util.SourceChainerServer,
		); err != nil {
			logger.WithError(err).Errorf("Failed to set pipeline state to creating for %s", pv.String())
			return fmt.Errorf("failed setting pipeline state: %w", err)
		}

		return nil
	}

	var msg *chainer.PipelineUpdateMessage
	servers := c.loadBalancer.GetServersForKey(pv.UID)
	cr.CreateNewPipelineIteration(c.conflictResolutioner, pv.Name, servers)

	var errs error
	for server, subscription := range c.streams {
		if contains(servers, server) {
			// we do not need to set pipeline state to creating if it is already in terminating state, and we need to delete it
			if pv.State.Status == pipeline.PipelineTerminating {
				msg = c.createPipelineDeletionMessage(pv, false)
			} else {
				msg = c.createPipelineCreationMessage(pv)
				pipelineState := pipeline.PipelineCreating
				if err := c.pipelineHandler.SetPipelineState(pv.Name, pv.Version, pv.UID, pipelineState, "Rebalance", util.SourceChainerServer); err != nil {
					logger.WithError(err).Errorf("Failed to set pipeline state to creating for %s", pv.String())
				}
			}
			msg.Timestamp = c.conflictResolutioner.GetTimestamp(pv.Name)

			select {
			case <-subscription.stream.Context().Done():
				err := subscription.stream.Context().Err()
				logger.WithError(err).Errorf("Failed to send create rebalance msg to pipeline %s stream ctx cancelled", pv.String())
				errs = errors.Join(errs, err)
			default:
				if err := subscription.stream.Send(msg); err != nil {
					logger.WithError(err).Errorf("Failed to send create rebalance msg to pipeline %s", pv.String())
					errs = errors.Join(errs, err)
				}
			}
			continue
		}

		msg = c.createPipelineDeletionMessage(pv, true)
		msg.Timestamp = c.conflictResolutioner.GetTimestamp(pv.Name)

		select {
		case <-subscription.stream.Context().Done():
			err := subscription.stream.Context().Err()
			logger.WithError(err).Errorf("Failed to send delete rebalance msg to pipeline %s stream ctx cancelled", pv.String())
			errs = errors.Join(errs, err)
		default:
			if err := subscription.stream.Send(msg); err != nil {
				logger.WithError(err).Errorf("Failed to send delete rebalance msg to pipeline %s", pv.String())
				errs = errors.Join(errs, err)
			}
		}
	}

	return errs
}

func (c *ChainerServer) rebalance() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, event := range c.pipelineHandler.GetAllRunningPipelineVersions() {
		if err := c.rebalancePipeline(event.PipelineName, event.PipelineVersion, event.UID); err != nil {
			c.logger.WithError(err).Errorf("Failed to rebalance pipeline %s", event.PipelineName)
		}
	}
}

func (c *ChainerServer) handleScalingConfigChanges() {
	logger := c.logger.WithField("func", "handleScalingConfigChanges")
	for {
		select {
		case newConfig := <-c.scalingConfigUpdates:
			if newConfig.Pipelines == nil {
				continue
			}
			c.configUpdatesMutex.Lock()
			if newConfig.Pipelines.MaxShardCountMultiplier != c.currentScalingConfig.Pipelines.MaxShardCountMultiplier {
				logger.Info("Updating mapping of Pipelines onto dataflow-engine replicas following scaling config change")
				// lock Mutex to avoid updating load balancer if a concurrent rebalance is in progress
				c.mu.Lock()
				c.currentScalingConfig = newConfig
				scaling_config.LogWhenUsingDefaultScalingConfig(&c.currentScalingConfig, logger)
				c.loadBalancer.UpdatePartitions(newConfig.Pipelines.MaxShardCountMultiplier)
				c.mu.Unlock()
				// rebalance all pipelines onto available dataflow-engine replicas according to new config
				c.rebalance()
			}
			c.configUpdatesMutex.Unlock()
		case <-c.done:
			return
		}
	}
}

func (c *ChainerServer) handlePipelineEvent(event coordinator.PipelineEventMsg) {
	logger := c.logger.WithField("func", "handlePipelineEvent")
	if event.ExperimentUpdate {
		return
	}

	// don't consider events from pipeline status or chainer server
	var pipelineEventSources = map[string]struct{}{
		util.SourcePipelineStatusEvent: {},
		util.SourceChainerServer:       {},
	}
	if _, ok := pipelineEventSources[event.Source]; ok {
		return
	}

	go func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		pv, err := c.pipelineHandler.GetPipelineVersion(event.PipelineName, event.PipelineVersion, event.UID)
		if err != nil {
			logger.WithError(err).Errorf("Failed to get pipeline from event %s", event.String())
			return
		}
		logger.Debugf("Received event %s with state %s", event.String(), pv.State.Status.String())

		// Handle case where we have no subscribers
		if len(c.streams) == 0 {
			if pv.State.Status == pipeline.PipelineTerminated {
				return
			}
			errMsg := "no dataflow engines available to handle pipeline"
			logger.WithField("pipeline", event.PipelineName).Warn(errMsg)

			status := pv.State.Status
			// if no dataflow engines available then we think we can terminate pipelines.
			// TODO: however it might be a networking glitch and we need to handle this better in future
			if pv.State.Status == pipeline.PipelineTerminating || pv.State.Status == pipeline.PipelineTerminate {
				status = pipeline.PipelineTerminated
			}
			err := c.pipelineHandler.SetPipelineState(pv.Name, pv.Version, pv.UID, status, errMsg, util.SourceChainerServer)
			if err != nil {
				logger.
					WithError(err).
					WithField("pipeline", pv.String()).
					WithField("status", status).
					Error("failed to set pipeline state")
			}

			return
		}
		switch pv.State.Status {
		case pipeline.PipelineCreate:
			err := c.pipelineHandler.SetPipelineState(pv.Name, pv.Version, pv.UID, pipeline.PipelineCreating, "", util.SourceChainerServer)
			if err != nil {
				logger.WithError(err).Errorf("Failed to set pipeline state to creating for %s", pv.String())
			}

			msg := c.createPipelineCreationMessage(pv)
			c.sendPipelineMsgToSelectedServers(msg, pv)

		case pipeline.PipelineTerminate:
			c.terminatePipeline(pv, event.KeepTopics)
		}
	}()
}

func (c *ChainerServer) terminatePipeline(pv *pipeline.PipelineVersion, keepTopics bool) {
	err := c.pipelineHandler.SetPipelineState(pv.Name, pv.Version, pv.UID, pipeline.PipelineTerminating, "", util.SourceChainerServer)
	if err != nil {
		c.logger.WithError(err).Errorf("Failed to set pipeline state to terminating for %s", pv.Name)
	}
	msg := c.createPipelineDeletionMessage(pv, keepTopics) // note pv is a copy and does not include the new change to terminating state
	c.sendPipelineMsgToSelectedServers(msg, pv)
}
