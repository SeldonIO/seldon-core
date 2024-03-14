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

func (s *SchedulerServer) SubscribePipelineStatus(req *pb.PipelineSubscriptionRequest, stream pb.Scheduler_SubscribePipelineStatusServer) error {
	logger := s.logger.WithField("func", "SubscribePipelineStatus")
	logger.Infof("Received subscribe request from %s", req.GetSubscriberName())

	err := s.sendCurrentPipelineStatuses(stream, false)
	if err != nil {
		return err
	}

	fin := make(chan bool)

	s.pipelineEventStream.mu.Lock()
	s.pipelineEventStream.streams[stream] = &PipelineSubscription{
		name:   req.GetSubscriberName(),
		stream: stream,
		fin:    fin,
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

func (s *SchedulerServer) sendCurrentPipelineStatuses(stream pb.Scheduler_SubscribePipelineStatusServer, allVersions bool) error {
	pipelines, err := s.pipelineHandler.GetPipelines()
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, err.Error())
	}
	for _, p := range pipelines {
		resp := createPipelineStatus(p, allVersions)
		s.logger.Debugf("Sending pipeline status %s", resp.String())

		_, err := sentWithTimeout(func() error { return stream.Send(resp) }, sendTimeout)
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
		hasExpired, err := sentWithTimeout(func() error { return stream.Send(status) }, sendTimeout)
		if hasExpired {
			// this should trigger a reconnect from the client
			close(subscription.fin)
			delete(s.pipelineEventStream.streams, stream)
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
