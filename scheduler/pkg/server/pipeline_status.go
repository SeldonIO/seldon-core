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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

const (
	addPipelineStreamEventSource = "pipeline.store.addpipelinestream"
)

func (s *SchedulerServer) PipelineStatusEvent(ctx context.Context, message *pb.PipelineUpdateStatusMessage) (*pb.PipelineUpdateStatusResponse, error) {
	logger := s.logger.WithField("func", "SubscribePipelineStatus")
	logger.Infof("Received pipeline status event %s", message.String())
	return &pb.PipelineUpdateStatusResponse{}, nil
}

func (s *SchedulerServer) SubscribePipelineStatus(req *pb.PipelineSubscriptionRequest, stream pb.Scheduler_SubscribePipelineStatusServer) error {
	logger := s.logger.WithField("func", "SubscribePipelineStatus")
	logger.Infof("Received subscribe request from %s", req.GetSubscriberName())

	s.synchroniser.WaitReady()

	fin := make(chan bool)

	s.pipelineEventStream.mu.Lock()
	s.pipelineEventStream.namesToIps[req.GetSubscriberName()] = req.GetSubscriberIp()
	s.pipelineEventStream.streams[stream] = &PipelineSubscription{
		name:              req.GetSubscriberName(),
		ip:                req.GetSubscriberIp(),
		isPipelineGateway: req.GetIsPipelineGateway(),
		stream:            stream,
		fin:               fin,
	}
	if req.IsPipelineGateway {
		s.pipelineGWLoadBalancer.AddServer(req.GetSubscriberName())
	}
	s.pipelineEventStream.mu.Unlock()

	if req.IsPipelineGateway {
		// rebalance the streams when a new subscriber is added
		s.pipelineGwRebalance()
	} else {
		// update controller with current model statuses
		err := s.sendCurrentPipelineStatuses(stream, false)
		if err != nil {
			logger.WithError(err).Errorf("Failed to send current pipeline statuses to %s", req.GetSubscriberName())
			return err
		}
	}

	ctx := stream.Context()
	// Keep this scope alive because once this scope exits - the stream is closed
	for {
		select {
		case <-fin:
			logger.Infof("Closing stream for %s", req.GetSubscriberName())
			return nil
		case <-ctx.Done():
			logger.Infof("Stream disconnected %s", req.GetSubscriberName())
			s.pipelineEventStream.mu.Lock()
			delete(s.pipelineEventStream.streams, stream)
			delete(s.pipelineEventStream.namesToIps, req.GetSubscriberName())
			if req.IsPipelineGateway {
				s.pipelineGWLoadBalancer.RemoveServer(req.GetSubscriberName())
			}
			s.pipelineEventStream.mu.Unlock()

			// rebalance the streams when a subscriber is removed
			if req.IsPipelineGateway {
				s.pipelineGwRebalance()
			}
			return nil
		}
	}
}

func (s *SchedulerServer) sendCurrentPipelineStatuses(
	stream pb.Scheduler_SubscribePipelineStatusServer,
	allVersions bool,
) error {
	pipelines, err := s.pipelineHandler.GetPipelines()
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "%s", err.Error())
	}
	for _, p := range pipelines {
		resp := createPipelineStatus(p, allVersions)
		s.logger.Debugf("Sending pipeline status %s", resp.String())

		_, err := sendWithTimeout(func() error { return stream.Send(resp) }, s.timeout)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SchedulerServer) GetAllRunningPipelines() []string {
	pipelineEventMessages := s.pipelineHandler.GetAllRunningPipelineVersions()
	runningPipelines := make([]string, 0, len(pipelineEventMessages))
	for _, pipelineEventMessage := range pipelineEventMessages {
		runningPipelines = append(runningPipelines, pipelineEventMessage.PipelineName)
	}
	return runningPipelines
}

func (s *SchedulerServer) createPipelineDeletionMessage(pip *pipeline.Pipeline, keepTopics bool) (*pb.PipelineStatusResponse, error) {
	return &pb.PipelineStatusResponse{
		PipelineName: pip.Name,
		Versions:     []*pb.PipelineWithState{},
	}, nil
}

func (s *SchedulerServer) createPipelineCreationMessage(pip *pipeline.Pipeline) (*pb.PipelineStatusResponse, error) {
	pipelineVersion := pip.GetLatestPipelineVersion()
	pipelineWithState := pipeline.CreatePipelineWithState(pipelineVersion)
	return &pb.PipelineStatusResponse{
		PipelineName: pip.Name,
		Versions:     []*pb.PipelineWithState{pipelineWithState},
	}, nil
}

func (s *SchedulerServer) pipelineGwRebalance() {
	s.pipelineEventStream.mu.Lock()
	defer s.pipelineEventStream.mu.Unlock()

	runningPipelines := s.GetAllRunningPipelines()
	s.logger.Debugf("Rebalancing pipeline gateways for running pipelines: %v", runningPipelines)

	for _, pipelineName := range runningPipelines {
		pip, _ := s.pipelineHandler.GetPipeline(pipelineName)
		consumerBucketId := s.getPipelineGatewayBucketId(pipelineName)
		servers := s.pipelineGWLoadBalancer.GetServersForKey(consumerBucketId)
		s.logger.Debugf("Servers for pipeline %s: %v", pipelineName, servers)
		s.logger.Debug("Consumer bucket ID: ", consumerBucketId)

		// need to update envoy clusters
		s.sendPipelineStreamsEventMsg(
			coordinator.PipelineEventMsg{PipelineName: pipelineName},
			servers,
		)

		// send messages to each pipeline gateway stream
		for _, pipelineSubscription := range s.pipelineEventStream.streams {
			if !pipelineSubscription.isPipelineGateway {
				s.logger.Debugf("Skipping non-pipeline gateway stream for %s", pipelineSubscription.name)
				continue
			}

			s.logger.Debug("Processing pipeline subscription for ", pipelineSubscription.name)
			server := pipelineSubscription.name
			stream := pipelineSubscription.stream

			if contains(servers, server) {
				s.logger.Debug("Server contains model, sending status update for ", server)

				state := pip.GetLatestPipelineVersion().State.Status
				var msg *pb.PipelineStatusResponse
				var err error

				if state == pipeline.PipelineTerminating {
					s.logger.Debugf("Pipeline %s is terminating, sending deletion message", pipelineName)
					msg, err = s.createPipelineDeletionMessage(pip, false)
				} else {
					s.logger.Debugf("Pipeline %s is available or progressing, sending creation message", pipelineName)
					msg, err = s.createPipelineCreationMessage(pip)
				}
				if err != nil {
					s.logger.WithError(err).Errorf("Failed to create pipelines status message for %s", pipelineName)
					continue
				}
				if err := stream.Send(msg); err != nil {
					s.logger.WithError(err).Errorf("Failed to send create rebalance msg to pipeline %s", pipelineName)
				}
			} else {
				s.logger.Debugf("Server %s does not contain pipeline %s, sending deletion message", server, pipelineName)
				msg, err := s.createPipelineDeletionMessage(pip, true)
				if err != nil {
					s.logger.WithError(err).Errorf("Failed to create pipeline deletion message for %s", pipelineName)
					continue
				}
				s.logger.Debugf("Sending deletion message for pipeline %s to server %s", pipelineName, server)
				if err := stream.Send(msg); err != nil {
					s.logger.WithError(err).Errorf("Failed to send delete rebalance msg to pipeline %s", pipelineName)
				}
			}
		}
	}
}

func (s *SchedulerServer) handlePipelineEvents(event coordinator.PipelineEventMsg) {
	logger := s.logger.WithField("func", "handlePipelineEvents")
	logger.Debugf("Received pipeline event %s", event.String())
	s.sendPipelineEvents(event)
}

func (s *SchedulerServer) sendPipelineEvents(event coordinator.PipelineEventMsg) {
	logger := s.logger.WithField("func", "sendPipelineEvents")
	if event.ExperimentUpdate {
		return
	}

	pv, err := s.pipelineHandler.GetPipelineVersion(event.PipelineName, event.PipelineVersion, event.UID)
	if err != nil {
		logger.WithError(err).Errorf("Failed to get pipeline from event %s", event.String())
		return
	}
	logger.Debugf("Handling pipeline event for %s with state %v", event.String(), pv.State.Status)

	var pipelineVersions []*pb.PipelineWithState
	pipelineWithState := pipeline.CreatePipelineWithState(pv)
	pipelineVersions = append(pipelineVersions, pipelineWithState)
	status := &pb.PipelineStatusResponse{
		PipelineName: pv.Name,
		Versions:     pipelineVersions,
	}

	s.pipelineEventStream.mu.Lock()
	defer s.pipelineEventStream.mu.Unlock()

	// find pipelinegw serverNames that should receive this event
	consumerBucketId := s.getPipelineGatewayBucketId(event.PipelineName)
	serverNames := s.pipelineGWLoadBalancer.GetServersForKey(consumerBucketId)

	// split the streams into pipeline gateways and non-gateways
	pipelineGwStreams := make(map[pb.Scheduler_SubscribePipelineStatusServer]*PipelineSubscription)
	otherStreams := make(map[pb.Scheduler_SubscribePipelineStatusServer]*PipelineSubscription)

	for stream, subscription := range s.pipelineEventStream.streams {
		if !subscription.isPipelineGateway {
			otherStreams[stream] = subscription
		} else if contains(serverNames, subscription.name) {
			pipelineGwStreams[stream] = subscription
		}
	}

	// send to pipeline gateway streams and other streams
	s.sendPipelineEventsToStreams(event, status, pipelineGwStreams)
	s.sendPipelineEventsToStreams(event, status, otherStreams)

	// publish event for envoy
	s.sendPipelineStreamsEventMsg(event, serverNames)
}

func (s *SchedulerServer) sendPipelineStreamsEventMsg(event coordinator.PipelineEventMsg, streamNames []string) {
	streamIps := make([]string, 0, len(streamNames))
	for _, streamName := range streamNames {
		ip, exists := s.pipelineEventStream.namesToIps[streamName]
		if !exists {
			s.logger.Errorf("No IP found for stream name %s", streamName)
			return
		}
		streamIps = append(streamIps, ip)
	}

	eventMsg := coordinator.PipelineStreamsEventMsg{
		PipelineEventMsg: event,
		StreamNames:      streamNames,
		StreamIps:        streamIps,
	}
	s.eventHub.PublishPipelineStreamsEvent(addPipelineStreamEventSource, eventMsg)
}

func (s *SchedulerServer) sendPipelineEventsToStreams(
	event coordinator.PipelineEventMsg,
	status *pb.PipelineStatusResponse,
	streams map[pb.Scheduler_SubscribePipelineStatusServer]*PipelineSubscription,
) {

	logger := s.logger.WithField("func", "sendPipelineEventsToStreams")
	for stream, subscription := range streams {
		hasExpired, err := sendWithTimeout(func() error { return stream.Send(status) }, s.timeout)
		if hasExpired {
			// this should trigger a reconnect from the client
			close(subscription.fin)
			delete(s.pipelineEventStream.streams, stream)
			delete(s.pipelineEventStream.namesToIps, subscription.name)
		}
		if err != nil {
			logger.WithError(err).Errorf("Failed to send pipeline status event to %s for %s", subscription.name, event.String())
		}
	}
}

func (s *SchedulerServer) StopSendPipelineEvents() {
	s.pipelineEventStream.mu.Lock()
	defer s.pipelineEventStream.mu.Unlock()
	for _, subscription := range s.pipelineEventStream.streams {
		close(subscription.fin)
	}
}

func (s *SchedulerServer) getPipelineGatewayBucketId(pipelineName string) string {
	return util.GetKafkaConsumerName(
		s.consumerGroupConfig.namespace,
		s.consumerGroupConfig.consumerGroupIdPrefix,
		pipelineName,
		pipelineGatewayConsumerNamePrefix,
		s.consumerGroupConfig.pipelineGatewayMaxNumConsumers,
	)
}
