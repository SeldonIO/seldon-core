package server

import (
	"context"

	"github.com/seldonio/seldon-core/scheduler/pkg/coordinator"

	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
)

func (s *SchedulerServer) SubscribeModelStatus(req *pb.ModelSubscriptionRequest, stream pb.Scheduler_SubscribeModelStatusServer) error {
	logger := s.logger.WithField("func", "SubscribeModelStatus")
	logger.Infof("Received subscribe request from %s", req.GetSubscriberName())

	fin := make(chan bool)

	s.mu.Lock()
	s.modelEventStream.streams[stream] = &ModelSubscription{
		name:   req.GetSubscriberName(),
		stream: stream,
		fin:    fin,
	}
	s.mu.Unlock()

	err := s.sendCurrentModelStatuses(stream)
	if err != nil {
		return err
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
			s.mu.Lock()
			delete(s.modelEventStream.streams, stream)
			s.mu.Unlock()
			return nil
		}
	}
}

//TODO as this could be 1000s of models may need to look at ways to optimize?
func (s *SchedulerServer) sendCurrentModelStatuses(stream pb.Scheduler_SubscribeModelStatusServer) error {
	modelNames := s.modelStore.GetAllModels()
	for _, modelName := range modelNames {
		model, err := s.modelStore.GetModel(modelName)
		if err != nil {
			return err
		}
		ms, err := s.modelStatusImpl(context.Background(), model, false)
		if err != nil {
			return err
		}
		err = stream.Send(ms)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SchedulerServer) handleModelEvent(event coordinator.ModelEventMsg) {
	logger := s.logger.WithField("func", "handleModelEvent")
	logger.Infof("Got model event msg for %s", event.String())

	// TODO - Should this spawn a goroutine?
	// Surely we're risking reordering of events, e.g. load/unload -> unload/load?
	go func() {
		err := s.sendModelStatusEvent(event)
		if err != nil {
			logger.WithError(err).Errorf("Failed to update model status for model %s", event.String())
		}
	}()
}

func (s *SchedulerServer) StopSendModelEvents() {
	for _, subscription := range s.modelEventStream.streams {
		close(subscription.fin)
	}
}

func (s *SchedulerServer) sendModelStatusEvent(evt coordinator.ModelEventMsg) error {
	logger := s.logger.WithField("func", "sendModelStatusEvent")
	model, err := s.modelStore.GetModel(evt.ModelName)
	if err != nil {
		return err
	}
	if model.GetLatest().GetVersion() == evt.ModelVersion {
		ms, err := s.modelStatusImpl(context.Background(), model, false)
		if err != nil {
			logger.WithError(err).Errorf("Failed to create model status for model %s", evt.String())
			return err
		}
		for stream, subscription := range s.modelEventStream.streams {
			err := stream.Send(ms)
			if err != nil {
				logger.WithError(err).Errorf("Failed to send model status event to %s for %s", subscription.name, evt.String())
			}
		}
	}
	return nil
}

func (s *SchedulerServer) SubscribeServerStatus(req *pb.ServerSubscriptionRequest, stream pb.Scheduler_SubscribeServerStatusServer) error {
	logger := s.logger.WithField("func", "SubscribeModelStatus")
	logger.Infof("Received server subscribe request from %s", req.GetSubscriberName())

	fin := make(chan bool)

	s.mu.Lock()
	s.serverEventStream.streams[stream] = &ServerSubscription{
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
			delete(s.serverEventStream.streams, stream)
			s.mu.Unlock()
			return nil
		}
	}
}

// TODO - Create a ServerStatusMsg type to disambiguate?
func (s *SchedulerServer) handleServerEvent(event coordinator.ModelEventMsg) {
	logger := s.logger.WithField("func", "handleServerEvent")
	logger.Infof("Got server state change for %s", event.String())

	// TODO - Should this spawn a goroutine?
	// Surely we're risking reordering of events, e.g. load/unload -> unload/load?
	go func() {
		err := s.sendServerStatusEvent(event)
		if err != nil {
			logger.WithError(err).Errorf("Failed to update server status for model event %s", event.String())
		}
	}()
}

func (s *SchedulerServer) StopSendServerEvents() {
	for _, subscription := range s.serverEventStream.streams {
		close(subscription.fin)
	}
}

func (s *SchedulerServer) sendServerStatusEvent(evt coordinator.ModelEventMsg) error {
	logger := s.logger.WithField("func", "sendServerStatusEvent")
	model, err := s.modelStore.GetModel(evt.ModelName)
	if err != nil {
		return err
	}
	modelVersion := model.GetVersion(evt.ModelVersion)
	if modelVersion == nil {
		logger.Warnf("Failed to find model version %s so ignoring event", evt.String())
		return nil
	}
	if modelVersion.Server() == "" {
		logger.Warnf("Empty server for %s so ignoring event", evt.String())
		return nil
	}
	ss, err := s.ServerStatus(context.Background(), &pb.ServerReference{Name: modelVersion.Server()})
	if err != nil {
		return err
	}
	for stream, subscription := range s.serverEventStream.streams {
		err := stream.Send(ss)
		if err != nil {
			logger.WithError(err).Errorf("Failed to send server status event to %s", subscription.name)
		}
	}
	return nil
}
