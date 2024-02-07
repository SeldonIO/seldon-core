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
	"fmt"
	"net"
	"sync"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/chainer"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/config"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

const (
	grpcMaxConcurrentStreams     = 1_000_000
	pipelineEventHandlerName     = "kafka.dataflow.server.pipelines"
	pendingEventsQueueSize   int = 10
	sourceChainerServer          = "chainer-server"
)

type ChainerServer struct {
	logger          log.FieldLogger
	mu              sync.Mutex
	streams         map[string]*ChainerSubscription
	eventHub        *coordinator.EventHub
	pipelineHandler pipeline.PipelineHandler
	topicNamer      *kafka.TopicNamer
	loadBalancer    util.LoadBalancer
	chainer.UnimplementedChainerServer
}

type ChainerSubscription struct {
	name   string
	stream chainer.Chainer_SubscribePipelineUpdatesServer
	fin    chan bool
}

func NewChainerServer(logger log.FieldLogger, eventHub *coordinator.EventHub, pipelineHandler pipeline.PipelineHandler,
	namespace string, loadBalancer util.LoadBalancer, kafkaConfig *config.KafkaConfig) (*ChainerServer, error) {
	topicNamer, err := kafka.NewTopicNamer(namespace, kafkaConfig.TopicPrefix)
	if err != nil {
		return nil, err
	}
	c := &ChainerServer{
		logger:          logger.WithField("source", "dataflow"),
		streams:         make(map[string]*ChainerSubscription),
		eventHub:        eventHub,
		pipelineHandler: pipelineHandler,
		topicNamer:      topicNamer,
		loadBalancer:    loadBalancer,
	}

	eventHub.RegisterPipelineEventHandler(
		pipelineEventHandlerName,
		pendingEventsQueueSize,
		c.logger,
		c.handlePipelineEvent,
	)
	return c, nil
}

func (c *ChainerServer) StartGrpcServer(agentPort uint) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", agentPort))
	if err != nil {
		log.Fatalf("failed to create listener: %v", err)
	}
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	grpcServer := grpc.NewServer(grpcOptions...)
	chainer.RegisterChainerServer(grpcServer, c)
	c.logger.Printf("Chainer server running on %d", agentPort)
	return grpcServer.Serve(lis)
}

func (c *ChainerServer) PipelineUpdateEvent(ctx context.Context, message *chainer.PipelineUpdateStatusMessage) (*chainer.PipelineUpdateStatusResponse, error) {
	logger := c.logger.WithField("func", "PipelineUpdateEvent")
	var statusVal pipeline.PipelineStatus
	switch message.Update.Op {
	case chainer.PipelineUpdateMessage_Create:
		if message.Success {
			statusVal = pipeline.PipelineReady
		} else {
			statusVal = pipeline.PipelineFailed
		}
	case chainer.PipelineUpdateMessage_Delete:
		if message.Success {
			statusVal = pipeline.PipelineTerminated
		} else {
			statusVal = pipeline.PipelineFailed
		}
	}

	if !message.Success {
		statusVal = pipeline.PipelineFailed
	}
	err := c.pipelineHandler.SetPipelineState(message.Update.Pipeline, message.Update.Version, message.Update.Uid, statusVal, message.Reason, sourceChainerServer)
	if err != nil {
		logger.WithError(err).Errorf("Failed to update pipeline status for %s:%d (%s)", message.Update.Pipeline, message.Update.Version, message.Update.Uid)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &chainer.PipelineUpdateStatusResponse{}, nil
}

func (c *ChainerServer) SubscribePipelineUpdates(req *chainer.PipelineSubscriptionRequest, stream chainer.Chainer_SubscribePipelineUpdatesServer) error {
	logger := c.logger.WithField("func", "SubscribePipelineStatus")
	logger.Infof("Received subscribe request from %s", req.GetName())

	fin := make(chan bool)

	c.mu.Lock()
	c.streams[req.Name] = &ChainerSubscription{
		name:   req.Name,
		stream: stream,
		fin:    fin,
	}
	c.loadBalancer.AddServer(req.Name)
	c.mu.Unlock()

	// Handle addition of new server
	//TODO delay this for x seconds after server start and do 1 rebalance after clients have joined
	// otherwise first server may be overloaded with all pipelines in system
	c.rebalance()

	ctx := stream.Context()
	// Keep this scope alive because once this scope exits - the stream is closed
	for {
		select {
		case <-fin:
			logger.Infof("Closing stream for %s", req.GetName())
			return nil
		case <-ctx.Done():
			logger.Infof("Stream disconnected %s", req.GetName())
			c.mu.Lock()
			c.loadBalancer.RemoveServer(req.Name)
			delete(c.streams, req.Name)
			c.mu.Unlock()
			// Handle removal of server
			c.rebalance()
			return nil
		}
	}
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

func (c *ChainerServer) createPipelineMessage(pv *pipeline.PipelineVersion) *chainer.PipelineUpdateMessage {
	var stepUpdates []*chainer.PipelineStepUpdate
	for _, step := range pv.Steps {
		stepUpdate := chainer.PipelineStepUpdate{
			Sources:   c.createTopicSources(step.Inputs, pv.Name),
			Triggers:  c.createTriggerSources(step.Triggers, pv.Name),
			Sink:      &chainer.PipelineTopic{PipelineName: pv.Name, TopicName: c.topicNamer.GetModelTopicInputs(step.Name), Tensor: nil},
			TensorMap: c.topicNamer.GetFullyQualifiedTensorMap(pv.Name, step.TensorMap),
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
		stepUpdates = append(stepUpdates, &stepUpdate)
	}
	if pv.Input != nil {
		stepUpdate := chainer.PipelineStepUpdate{
			Sources:   c.createPipelineTopicSources(pv.Input.ExternalInputs),
			Sink:      &chainer.PipelineTopic{PipelineName: pv.Name, TopicName: c.topicNamer.GetPipelineTopicInputs(pv.Name), Tensor: nil},
			Triggers:  c.createPipelineTopicSources(pv.Input.ExternalTriggers),
			TensorMap: c.topicNamer.GetFullyQualifiedPipelineTensorMap(pv.Input.TensorMap),
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
		stepUpdates = append(stepUpdates, &stepUpdate)
	}
	if pv.Output != nil {
		stepUpdate := chainer.PipelineStepUpdate{
			Sources:   c.createTopicSources(pv.Output.Steps, pv.Name),
			Sink:      &chainer.PipelineTopic{PipelineName: pv.Name, TopicName: c.topicNamer.GetPipelineTopicOutputs(pv.Name), Tensor: nil},
			TensorMap: c.topicNamer.GetFullyQualifiedTensorMap(pv.Name, pv.Output.TensorMap),
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
		stepUpdates = append(stepUpdates, &stepUpdate)
	}
	//Append an error step to send any errors to pipeline output
	stepUpdates = append(stepUpdates, &chainer.PipelineStepUpdate{
		Sources:     []*chainer.PipelineTopic{{PipelineName: pv.Name, TopicName: c.topicNamer.GetModelErrorTopic(), Tensor: nil}},
		Sink:        &chainer.PipelineTopic{PipelineName: pv.Name, TopicName: c.topicNamer.GetPipelineTopicOutputs(pv.Name), Tensor: nil},
		InputJoinTy: chainer.PipelineStepUpdate_Inner,
	})

	op := chainer.PipelineUpdateMessage_Create
	if pv.State.Status == pipeline.PipelineTerminate {
		op = chainer.PipelineUpdateMessage_Delete
	}
	return &chainer.PipelineUpdateMessage{
		Pipeline: pv.Name,
		Version:  pv.Version,
		Uid:      pv.UID,
		Updates:  stepUpdates,
		Op:       op,
	}
}

func (c *ChainerServer) sendPipelineMsgToSelectedServers(msg *chainer.PipelineUpdateMessage, pv *pipeline.PipelineVersion) {
	logger := c.logger.WithField("func", "sendPipelineMsg")
	servers := c.loadBalancer.GetServersForKey(pv.UID)
	for _, serverId := range servers {
		if subscription, ok := c.streams[serverId]; ok {
			if err := subscription.stream.Send(msg); err != nil {
				logger.WithError(err).Errorf("Failed to send msg to pipeline %s", pv.String())
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

func (c *ChainerServer) rebalance() {
	logger := c.logger.WithField("func", "rebalance")
	// note that we are not retrying PipelineFailed pipelines, consider adding this
	evts := c.pipelineHandler.GetAllRunningPipelineVersions()
	for _, event := range evts {
		c.logger.Debugf("Rebalancing pipeline %s:%d with state $s", event.PipelineName, event.PipelineVersion)
		pv, err := c.pipelineHandler.GetPipelineVersion(event.PipelineName, event.PipelineVersion, event.UID)
		if err != nil {
			logger.WithError(err).Errorf("Failed to get pipeline from event %s", event.String())
			continue
		}
		c.mu.Lock()
		if len(c.streams) == 0 {
			if err := c.pipelineHandler.SetPipelineState(
				pv.Name,
				pv.Version,
				pv.UID,
				pipeline.PipelineCreate,
				"no dataflow engines available to handle pipeline",
				sourceChainerServer,
			); err != nil {
				logger.WithError(err).Errorf("Failed to set pipeline state to creating for %s", pv.String())
			}
		} else {
			msg := c.createPipelineMessage(pv)
			servers := c.loadBalancer.GetServersForKey(pv.UID)
			for server, subscription := range c.streams {
				if contains(servers, server) {
					msg.Op = chainer.PipelineUpdateMessage_Create
					if err := c.pipelineHandler.SetPipelineState(pv.Name, pv.Version, pv.UID, pipeline.PipelineCreating, "Rebalance", sourceChainerServer); err != nil {
						logger.WithError(err).Errorf("Failed to set pipeline state to creating for %s", pv.String())
					}
					if err := subscription.stream.Send(msg); err != nil {
						logger.WithError(err).Errorf("Failed to send create rebalance msg to pipeline %s", pv.String())
					}
				} else {
					msg.Op = chainer.PipelineUpdateMessage_Delete
					if err := subscription.stream.Send(msg); err != nil {
						logger.WithError(err).Errorf("Failed to send delete rebalance msg to pipeline %s", pv.String())
					}
				}
			}
		}
		c.mu.Unlock()
	}
}

func (c *ChainerServer) handlePipelineEvent(event coordinator.PipelineEventMsg) {
	logger := c.logger.WithField("func", "handlePipelineEvent")
	if event.ExperimentUpdate {
		return
	}
	if sourceChainerServer == event.Source {
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
			errMsg := "no dataflow engines available to handle pipeline"
			logger.WithField("pipeline", event.PipelineName).Warn(errMsg)

			err := c.pipelineHandler.SetPipelineState(pv.Name, pv.Version, pv.UID, pv.State.Status, errMsg, sourceChainerServer)
			if err != nil {
				logger.
					WithError(err).
					WithField("pipeline", pv.String()).
					WithField("status", pv.State.Status).
					Error("failed to set pipeline state")
			}

			return
		}

		switch pv.State.Status {
		case pipeline.PipelineCreate:
			err := c.pipelineHandler.SetPipelineState(pv.Name, pv.Version, pv.UID, pipeline.PipelineCreating, "", sourceChainerServer)
			if err != nil {
				logger.WithError(err).Errorf("Failed to set pipeline state to creating for %s", pv.String())
			}

			msg := c.createPipelineMessage(pv)
			c.sendPipelineMsgToSelectedServers(msg, pv)

		case pipeline.PipelineTerminate:
			err := c.pipelineHandler.SetPipelineState(pv.Name, pv.Version, pv.UID, pipeline.PipelineTerminating, "", sourceChainerServer)
			if err != nil {
				logger.WithError(err).Errorf("Failed to set pipeline state to terminating for %s", pv.String())
			}
			msg := c.createPipelineMessage(pv)
			c.sendPipelineMsgToSelectedServers(msg, pv)
		}
	}()
}
