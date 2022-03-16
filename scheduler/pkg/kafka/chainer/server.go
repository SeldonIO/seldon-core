package chainer

import (
	"context"
	"fmt"
	"net"
	"sync"

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
	pipelineEventHandlerName     = "kafka.chainer.server.pipelines"
	pendingEventsQueueSize   int = 10
	seldonTopicPrefix            = "seldon"
	modelTopic                   = "model"
	pipelineTopic                = "pipeline"
)

type ChainerServer struct {
	logger          log.FieldLogger
	mu              sync.Mutex
	namespace       string
	streams         map[chainer.Chainer_SubscribePipelineUpdatesServer]*ChainerSubscription
	eventHub        *coordinator.EventHub
	pipelineHandler pipeline.PipelineHandler
	chainer.UnimplementedChainerServer
}

type ChainerSubscription struct {
	name   string
	stream chainer.Chainer_SubscribePipelineUpdatesServer
	fin    chan bool
}

func NewChainerServer(logger log.FieldLogger, eventHub *coordinator.EventHub, pipelineHandler pipeline.PipelineHandler, namespace string) *ChainerServer {
	c := &ChainerServer{
		logger:          logger.WithField("source", "chainer"),
		namespace:       namespace,
		streams:         make(map[chainer.Chainer_SubscribePipelineUpdatesServer]*ChainerSubscription),
		eventHub:        eventHub,
		pipelineHandler: pipelineHandler,
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
	statusVal := pipeline.PipelineReady
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
		source := fmt.Sprintf("%s.%s.%s.%s", seldonTopicPrefix, c.namespace, modelTopic, inp)
		sources = append(sources, source)
	}
	if len(sources) == 0 {
		sources = append(sources, fmt.Sprintf("%s.%s.%s.%s.inputs", seldonTopicPrefix, c.namespace, pipelineTopic, pipelineName))
	}
	return sources
}

func (c *ChainerServer) createPipelineMessage(pv *pipeline.PipelineVersion) *chainer.PipelineUpdateMessage {
	var stepUpdates []*chainer.PipelineStepUpdate
	for _, step := range pv.Steps {
		stepUpdate := chainer.PipelineStepUpdate{
			Sources: c.createTopicSources(step.Inputs, pv.Name),
			Sink:    step.Name,
			Ty:      chainer.PipelineStepUpdate_Inner,
		}
		stepUpdates = append(stepUpdates, &stepUpdate)
	}
	if pv.Output != nil {
		stepUpdate := chainer.PipelineStepUpdate{
			Sources: c.createTopicSources(pv.Output.Inputs, pv.Name),
			Sink:    fmt.Sprintf("%s.%s.%s.%s.outputs", seldonTopicPrefix, c.namespace, pipelineTopic, pv.Name),
			Ty:      chainer.PipelineStepUpdate_Inner,
		}
		stepUpdates = append(stepUpdates, &stepUpdate)
	}
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
	case pipeline.PipelineCreate, pipeline.PipelineTerminate:
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
	}
}
