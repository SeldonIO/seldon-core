package server

import (
	"context"

	"github.com/seldonio/seldon-core/scheduler/pkg/coordinator"

	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
)

func (s *SchedulerServer) SubscribeModelStatus(req *pb.ModelSubscriptionRequest, stream pb.Scheduler_SubscribeModelStatusServer) error {
	logger := s.logger.WithField("func", "SubscribeModelStatus")
	logger.Infof("Received subscribe request from %s", req.GetName())

	fin := make(chan bool)

	s.mutext.Lock()
	s.modelEventStream.streams[stream] = &ModelSubscription{
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
			delete(s.modelEventStream.streams, stream)
			s.mutext.Unlock()
			return nil
		}
	}
}

func (s *SchedulerServer) ListenForModelEvents() {
	logger := s.logger.WithField("func", "ListenForModelEvents")
	for modelEventMsg := range s.modelEventStream.chanEvent {
		logger.Infof("Got model event msg for %s", modelEventMsg.String())
		msg := modelEventMsg
		go func() {
			err := s.sendModelStatusEvent(msg)
			if err != nil {
				logger.WithError(err).Errorf("Failed to update model status for model %s", msg.String())
			}
		}()
	}
}

func (s *SchedulerServer) StopSendModelEvents() {
	for _, subscription := range s.modelEventStream.streams {
		close(subscription.fin)
	}
}

func (s *SchedulerServer) sendModelStatusEvent(evt coordinator.ModelEventMsg) error {
	logger := s.logger.WithField("func", "sendModelStatusEvent")
	model, err := s.store.GetModel(evt.ModelName)
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
	logger.Infof("Received server subscribe request from %s", req.GetName())

	fin := make(chan bool)

	s.mutext.Lock()
	s.serverEventStream.streams[stream] = &ServerSubscription{
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
			delete(s.serverEventStream.streams, stream)
			s.mutext.Unlock()
			return nil
		}
	}
}

func (s *SchedulerServer) ListenForServerEvents() {
	logger := s.logger.WithField("func", "ListenForServerEvents")
	for modelEventMsg := range s.serverEventStream.chanEvent {
		logger.Infof("Got server state change for %s", modelEventMsg.String())

		evt := modelEventMsg
		go func() {
			err := s.sendServerStatusEvent(evt)
			if err != nil {
				logger.WithError(err).Errorf("Failed to update server status for model event %s", evt.String())
			}
		}()
	}
}

func (s *SchedulerServer) StopSendServerEvents() {
	for _, subscription := range s.serverEventStream.streams {
		close(subscription.fin)
	}
}

func (s *SchedulerServer) sendServerStatusEvent(evt coordinator.ModelEventMsg) error {
	logger := s.logger.WithField("func", "sendServerStatusEvent")
	model, err := s.store.GetModel(evt.ModelName)
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
