package server

import (
	"context"

	"github.com/seldonio/seldon-core/scheduler/pkg/store"

	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
)

func (s *SchedulerServer) SubscribeModelStatus(req *pb.ModelSubscriptionRequest, stream pb.Scheduler_SubscribeModelStatusServer) error {
	logger := s.logger.WithField("func", "SubscribeModelStatus")
	logger.Infof("Received subscribe request from %s", req.GetName())

	fin := make(chan bool)

	s.mutext.Lock()
	s.streams[stream] = &Subscription{
		name:   req.Name,
		stream: stream,
		fin:    fin,
	}
	s.mutext.Unlock()

	ctx := stream.Context()
	// Keep this scope alive because once this scope exits - the stream is closed
	for {
		select {
		case <-fin:
			logger.Infof("Closing stream for %s", req.GetName())
			return nil
		case <-ctx.Done():
			logger.Infof("Stream disconnected %s", req.GetName())
			s.mutext.Lock()
			delete(s.streams, stream)
			s.mutext.Unlock()
			return nil
		}
	}
}

func (s *SchedulerServer) ListenForEvents() {
	logger := s.logger.WithField("func", "ListenForEvents")
	for modelSnapshot := range s.chanEvent {
		logger.Infof("Got model state change for %s", modelSnapshot.Name)
		modelSnapshot := modelSnapshot
		go func() {
			err := s.ModelStatusEvent(modelSnapshot)
			if err != nil {
				logger.WithError(err).Errorf("Faile to update status for model %s", modelSnapshot.Name)
			}
		}()
	}
}

func (s *SchedulerServer) StopListenForEvents() {
	close(s.chanEvent)
}

func (s *SchedulerServer) ModelStatusEvent(modelSnapshot *store.ModelSnapshot) error {
	logger := s.logger.WithField("func", "ModelStatusEvent")
	ms, err := s.modelStatusImpl(context.Background(), modelSnapshot, false)
	if err != nil {
		logger.WithError(err).Errorf("Failed to create model status for model %s", modelSnapshot.Name)
		return err
	}
	for stream, subscription := range s.streams {
		err := stream.Send(ms)
		if err != nil {
			logger.WithError(err).Errorf("Failed to send model status event to %s", subscription.name)
		}
	}
	return nil
}
