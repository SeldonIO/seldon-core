package server

import (
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/coordinator"
)

func (s *SchedulerServer) SubscribePipelineStatus(req *pb.PipelineSubscriptionRequest, stream pb.Scheduler_SubscribePipelineStatusServer) error {
	logger := s.logger.WithField("func", "SubscribePipelineStatus")
	logger.Infof("Received subscribe request from %s", req.GetSubscriberName())

	fin := make(chan bool)

	s.mu.Lock()
	s.pipelineEventStream.streams[stream] = &PipelineSubscription{
		name:   req.GetSubscriberName(),
		stream: stream,
		fin:    fin,
	}
	s.mu.Unlock()

	ctx := stream.Context()
	// Keep this scope alive because once this scope exits - the stream is closed
	for {
		select {
		case <-fin:
			logger.Infof("Closing stream for %s", req.GetSubscriberName())
			return nil
		case <-ctx.Done():
			logger.Infof("Stream disconnected %s", req.GetSubscriberName())
			s.mu.Lock()
			delete(s.pipelineEventStream.streams, stream)
			s.mu.Unlock()
			return nil
		}
	}
}

func (s *SchedulerServer) handlePipelineEvents(event coordinator.PipelineEventMsg) {
	logger := s.logger.WithField("func", "handlePipelineEvents")
	logger.Debugf("Received pipeline event %s", event.String())
	pv, err := s.pipelineHandler.GetPipelineVersion(event.PipelineName, event.PipelineVersion, event.UID)
	if err != nil {
		logger.WithError(err).Errorf("Failed to get pipeline from event %s", event.String())
		return
	}
	var pipelineVersions []*pb.PipelineWithState
	pipelineWithState := createPipelineVersionState(pv)
	pipelineVersions = append(pipelineVersions, pipelineWithState)
	status := &pb.PipelineStatusResponse{
		PipelineName: pv.Name,
		Versions:     pipelineVersions,
	}
	for stream, subscription := range s.pipelineEventStream.streams {
		if err := stream.Send(status); err != nil {
			logger.WithError(err).Errorf("Failed to send pipeline status event to %s for %s", subscription.name, event.String())
		}
	}
}

func (s *SchedulerServer) StopSendPipelineEvents() {
	for _, subscription := range s.pipelineEventStream.streams {
		close(subscription.fin)
	}
}
