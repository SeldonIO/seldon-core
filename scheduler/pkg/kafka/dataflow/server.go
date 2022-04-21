package dataflow

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/seldonio/seldon-core/scheduler/pkg/kafka"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/seldonio/seldon-core/scheduler/apis/mlops/chainer"
	"github.com/seldonio/seldon-core/scheduler/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/pkg/store/pipeline"
	log "github.com/sirupsen/logrus"
)

const (
	grpcMaxConcurrentStreams     = 1_000_000
	pipelineEventHandlerName     = "kafka.dataflow.server.pipelines"
	pendingEventsQueueSize   int = 10
)

type ChainerServer struct {
	logger          log.FieldLogger
	mu              sync.Mutex
	streams         map[chainer.Chainer_SubscribePipelineUpdatesServer]*ChainerSubscription
	eventHub        *coordinator.EventHub
	pipelineHandler pipeline.PipelineHandler
	topicNamer      *kafka.TopicNamer
	chainer.UnimplementedChainerServer
}

type ChainerSubscription struct {
	name   string
	stream chainer.Chainer_SubscribePipelineUpdatesServer
	fin    chan bool
}

func NewChainerServer(logger log.FieldLogger, eventHub *coordinator.EventHub, pipelineHandler pipeline.PipelineHandler, namespace string) *ChainerServer {
	c := &ChainerServer{
		logger:          logger.WithField("source", "dataflow"),
		streams:         make(map[chainer.Chainer_SubscribePipelineUpdatesServer]*ChainerSubscription),
		eventHub:        eventHub,
		pipelineHandler: pipelineHandler,
		topicNamer:      kafka.NewTopicNamer(namespace),
	}

	eventHub.RegisterPipelineEventHandler(
		pipelineEventHandlerName,
		pendingEventsQueueSize,
		c.logger,
		c.handlePipelineEvent,
	)
	return c
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
	err := c.pipelineHandler.SetPipelineState(message.Update.Pipeline, message.Update.Version, message.Update.Uid, statusVal, message.Reason)
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
	c.streams[stream] = &ChainerSubscription{
		name:   req.Name,
		stream: stream,
		fin:    fin,
	}
	c.mu.Unlock()

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
			delete(c.streams, stream)
			c.mu.Unlock()
			return nil
		}
	}
}

func (c *ChainerServer) StopSendPipelineEvents() {
	for _, subscription := range c.streams {
		close(subscription.fin)
	}
}

func (c *ChainerServer) createTopicSources(inputs []string, pipelineName string) []string {
	var sources []string
	for _, inp := range inputs {
		source := c.topicNamer.GetModelTopic(inp)
		sources = append(sources, source)
	}
	if len(sources) == 0 {
		sources = append(sources, c.topicNamer.GetPipelineTopicInputs(pipelineName))
	}
	return sources
}

func (c *ChainerServer) createTriggerSources(inputs []string) []string {
	var sources []string
	for _, inp := range inputs {
		source := c.topicNamer.GetModelTopic(inp)
		sources = append(sources, source)
	}
	return sources
}

func (c *ChainerServer) createPipelineMessage(pv *pipeline.PipelineVersion) *chainer.PipelineUpdateMessage {
	var stepUpdates []*chainer.PipelineStepUpdate
	for _, step := range pv.Steps {
		stepUpdate := chainer.PipelineStepUpdate{
			Sources:   c.createTopicSources(step.Inputs, pv.Name),
			Triggers:  c.createTriggerSources(step.Triggers),
			Sink:      c.topicNamer.GetModelTopicInputs(step.Name),
			TensorMap: c.topicNamer.GetFullyQualifiedTensorMap(step.TensorMap),
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
	if pv.Output != nil {
		stepUpdate := chainer.PipelineStepUpdate{
			Sources: c.createTopicSources(pv.Output.Steps, pv.Name),
			Sink:    c.topicNamer.GetPipelineTopicOutputs(pv.Name),
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
		Sources:     []string{c.topicNamer.GetModelErrorTopic()},
		Sink:        c.topicNamer.GetPipelineTopicOutputs(pv.Name),
		InputJoinTy: chainer.PipelineStepUpdate_Inner,
	})

	op := chainer.PipelineUpdateMessage_Create
	if pv.State.Status != pipeline.PipelineCreate {
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

func (c *ChainerServer) handlePipelineEvent(event coordinator.PipelineEventMsg) {
	logger := c.logger.WithField("func", "handlePipelineEvent")
	pv, err := c.pipelineHandler.GetPipelineVersion(event.PipelineName, event.PipelineVersion, event.UID)
	if err != nil {
		logger.WithError(err).Errorf("Failed to get pipeline from event %s", event.String())
		return
	}
	logger.Debugf("Received event %s with state %s", event.String(), pv.State.Status.String())
	switch pv.State.Status {
	case pipeline.PipelineCreate:
		msg := c.createPipelineMessage(pv)
		for _, subscription := range c.streams {
			if err := subscription.stream.Send(msg); err != nil {
				logger.WithError(err).Errorf("Failed to send msg for pipeline %s", pv.String())
			} else {
				if err := c.pipelineHandler.SetPipelineState(pv.Name, pv.Version, pv.UID, pipeline.PipelineCreating, ""); err != nil {
					logger.WithError(err).Errorf("Failed to set pipeline %s to creating state", pv.String())
				}
			}
		}
	case pipeline.PipelineTerminate:
		msg := c.createPipelineMessage(pv)
		for _, subscription := range c.streams {
			if err := subscription.stream.Send(msg); err != nil {
				logger.WithError(err).Errorf("Failed to send msg for pipeline %s", pv.String())
			} else {
				if err := c.pipelineHandler.SetPipelineState(pv.Name, pv.Version, pv.UID, pipeline.PipelineTerminating, ""); err != nil {
					logger.WithError(err).Errorf("Failed to set pipeline %s to terminate state", pv.String())
				}
			}
		}
	}
}
