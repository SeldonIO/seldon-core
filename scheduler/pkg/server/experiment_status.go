package server

import (
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/coordinator"
)

func (s *SchedulerServer) SubscribeExperimentStatus(req *pb.ExperimentSubscriptionRequest, stream pb.Scheduler_SubscribeExperimentStatusServer) error {
	logger := s.logger.WithField("func", "SubscribeExperimentStatus")
	logger.Infof("Received subscribe request from %s", req.GetSubscriberName())

	fin := make(chan bool)

	s.mu.Lock()
	s.experimentEventStream.streams[stream] = &ExperimentSubscription{
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
			delete(s.experimentEventStream.streams, stream)
			s.mu.Unlock()
			return nil
		}
	}
}

func asKubernetesMeta(event coordinator.ExperimentEventMsg) *pb.KubernetesMeta {
	if event.KubernetesMeta != nil {
		return &pb.KubernetesMeta{
			Namespace:  event.KubernetesMeta.Namespace,
			Generation: event.KubernetesMeta.Generation,
		}
	}
	return nil
}

func (s *SchedulerServer) handleExperimentEvents(event coordinator.ExperimentEventMsg) {
	logger := s.logger.WithField("func", "handleExperimentEvents")
	logger.Debugf("Received experiment event %s", event.String())
	if event.Status != nil {
		for stream, subscription := range s.experimentEventStream.streams {
			err := stream.Send(&pb.ExperimentStatusResponse{
				ExperimentName:    event.ExperimentName,
				Active:            event.Status.Active,
				CandidatesReady:   event.Status.CandidatesReady,
				MirrorReady:       event.Status.MirrorReady,
				StatusDescription: event.Status.StatusDescription,
				KubernetesMeta:    asKubernetesMeta(event),
			})
			if err != nil {
				logger.WithError(err).Errorf("Failed to send experiment status event to %s for %s", subscription.name, event.String())
			}
		}
	}
}

func (s *SchedulerServer) StopSendExperimentEvents() {
	for _, subscription := range s.experimentEventStream.streams {
		close(subscription.fin)
	}
}
