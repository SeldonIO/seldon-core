/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package server

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
)

const (
	addPipelineStreamEventSource = "pipeline.store.addpipelinestream"
)

func (s *SchedulerServer) SubscribePipelineStatus(req *pb.PipelineSubscriptionRequest, stream pb.Scheduler_SubscribePipelineStatusServer) error {
	logger := s.logger.WithField("func", "SubscribePipelineStatus")
	logger.Infof("Received subscribe request from %s", req.GetSubscriberName())

	s.synchroniser.WaitReady()

	err := s.sendCurrentPipelineStatuses(stream, false)
	if err != nil {
		return err
	}

	fin := make(chan bool)

	s.pipelineEventStream.mu.Lock()
	s.pipelineEventStream.streams[stream] = &PipelineSubscription{
		name:              req.GetSubscriberName(),
		ip:                req.GetSubscriberIp(),
		isPipelineGateway: req.GetIsPipelineGateway(),
		stream:            stream,
		fin:               fin,
	}
	s.pipelineEventStream.mu.Unlock()

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
			s.pipelineEventStream.mu.Unlock()
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
	for stream, subscription := range s.pipelineEventStream.streams {
		hasExpired, err := sendWithTimeout(func() error { return stream.Send(status) }, s.timeout)
		if hasExpired {
			// this should trigger a reconnect from the client
			close(subscription.fin)
			delete(s.pipelineEventStream.streams, stream)
		}
		if err != nil {
			logger.WithError(err).Errorf("Failed to send pipeline status event to %s for %s", subscription.name, event.String())
		}
	}

	eventMsg := coordinator.PipelineStreamsEventMsg{
		PipelineEventMsg: event,
		StreamNames:      s.getStreamNames(),
		StreamIps:        s.getStreamIps(),
	}
	s.eventHub.PublishPipelineStreamsEvent(addPipelineStreamEventSource, eventMsg)
}

func (s *SchedulerServer) getStreamNames() []string {
	streamNames := make([]string, 0, len(s.pipelineEventStream.streams))
	for _, subscription := range s.pipelineEventStream.streams {
		if !subscription.isPipelineGateway {
			continue
		}
		streamNames = append(streamNames, subscription.name)
	}
	return streamNames
}

func (s *SchedulerServer) getStreamIps() []string {
	streamIps := make([]string, 0, len(s.pipelineEventStream.streams))
	for _, subscription := range s.pipelineEventStream.streams {
		if !subscription.isPipelineGateway {
			continue
		}
		streamIps = append(streamIps, subscription.ip)
	}
	return streamIps
}

func (s *SchedulerServer) StopSendPipelineEvents() {
	s.pipelineEventStream.mu.Lock()
	defer s.pipelineEventStream.mu.Unlock()
	for _, subscription := range s.pipelineEventStream.streams {
		close(subscription.fin)
	}
}
