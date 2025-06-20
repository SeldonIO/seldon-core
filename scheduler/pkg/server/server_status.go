/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package server

import (
	"time"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

func (s *SchedulerServer) SubscribeModelStatus(req *pb.ModelSubscriptionRequest, stream pb.Scheduler_SubscribeModelStatusServer) error {
	logger := s.logger.WithField("func", "SubscribeModelStatus")
	logger.Infof("Received subscribe request from %s", req.GetSubscriberName())

	s.synchroniser.WaitReady()

	err := s.sendCurrentModelStatuses(stream)
	if err != nil {
		logger.WithError(err).Errorf("Failed to send current model statuses to %s", req.GetSubscriberName())
		return err
	}

	fin := make(chan bool)

	s.modelEventStream.mu.Lock()
	s.modelEventStream.streams[stream] = &ModelSubscription{
		name:   req.GetSubscriberName(),
		stream: stream,
		fin:    fin,
	}
	s.modelEventStream.mu.Unlock()

	ctx := stream.Context()
	// Keep this scope alive because once this scope exits - the stream is closed
	for {
		select {
		case <-fin:
			logger.Infof("Closing stream for %s", req.GetSubscriberName())
			return nil
		case <-ctx.Done():
			logger.Infof("Stream disconnected %s", req.GetSubscriberName())
			s.modelEventStream.mu.Lock()
			delete(s.modelEventStream.streams, stream)
			s.modelEventStream.mu.Unlock()
			return nil
		}
	}
}

// TODO as this could be 1000s of models may need to look at ways to optimize?
func (s *SchedulerServer) sendCurrentModelStatuses(stream pb.Scheduler_SubscribeModelStatusServer) error {
	modelNames := s.modelStore.GetAllModels()
	for _, modelName := range modelNames {
		model, err := s.modelStore.GetModel(modelName)
		if err != nil {
			return err
		}
		ms, err := s.modelStatusImpl(model, false)
		if err != nil {
			return err
		}
		// no need to have a lock here as we are in the initial setup
		_, err = sendWithTimeout(func() error { return stream.Send(ms) }, s.timeout)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SchedulerServer) handleModelEvent(event coordinator.ModelEventMsg) {
	logger := s.logger.WithField("func", "handleModelEvent")
	logger.Debugf("Got model event msg for %s", event.String())

	// TODO - Should this spawn a goroutine?
	// Surely if we do we're risking reordering of events, e.g. load/unload -> unload/load?
	err := s.sendModelStatusEvent(event)
	if err != nil {
		logger.WithError(err).Errorf("Failed to update model status for model %s", event.String())
	}
}

func (s *SchedulerServer) StopSendModelEvents() {
	s.modelEventStream.mu.Lock()
	defer s.modelEventStream.mu.Unlock()
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
	if model.GetLatest() != nil && model.GetLatest().GetVersion() == evt.ModelVersion {
		ms, err := s.modelStatusImpl(model, false)
		if err != nil {
			logger.WithError(err).Errorf("Failed to create model status for model %s", evt.String())
			return err
		}
		s.modelEventStream.mu.Lock()
		defer s.modelEventStream.mu.Unlock()
		for stream, subscription := range s.modelEventStream.streams {
			hasExpired, err := sendWithTimeout(func() error { return stream.Send(ms) }, s.timeout)
			if hasExpired {
				// this should trigger a reconnect from the client
				close(subscription.fin)
				delete(s.modelEventStream.streams, stream)
			}
			if err != nil {
				logger.WithError(err).Errorf("Failed to send model status event to %s for %s", subscription.name, evt.String())
			}
		}
	}
	return nil
}

func (s *SchedulerServer) SubscribeServerStatus(req *pb.ServerSubscriptionRequest, stream pb.Scheduler_SubscribeServerStatusServer) error {
	logger := s.logger.WithField("func", "SubscribeServerStatus")
	logger.Infof("Received subscribe request from %s", req.GetSubscriberName())

	err := s.sendCurrentServerStatuses(stream)
	if err != nil {
		logger.WithError(err).Errorf("Failed to send current server statuses to %s", req.GetSubscriberName())
		return err
	}

	fin := make(chan bool)

	s.serverEventStream.mu.Lock()
	s.serverEventStream.streams[stream] = &ServerSubscription{
		name:   req.GetSubscriberName(),
		stream: stream,
		fin:    fin,
	}
	s.serverEventStream.mu.Unlock()

	ctx := stream.Context()
	// Keep this scope alive because once this scope exits - the stream is closed
	for {
		select {
		case <-fin:
			logger.Infof("Closing stream for %s", req.GetSubscriberName())
			return nil
		case <-ctx.Done():
			logger.Infof("Stream disconnected %s", req.GetSubscriberName())
			s.serverEventStream.mu.Lock()
			delete(s.serverEventStream.streams, stream)
			s.serverEventStream.mu.Unlock()
			return nil
		}
	}
}

func (s *SchedulerServer) handleModelEventForServerStatus(event coordinator.ModelEventMsg) {
	logger := s.logger.WithField("func", "handleModelEventForServerStatus")
	logger.Debugf("Got server state change for %s", event.String())

	err := s.updateServerModelsStatus(event)
	if err != nil {
		logger.WithError(err).Errorf("Failed to update server status for model event %s", event.String())
	}
}

func (s *SchedulerServer) handleServerEvents(event coordinator.ServerEventMsg) {
	logger := s.logger.WithField("func", "handleServerEvents")
	logger.Debugf("Got server state %s change for %s", event.ServerName, event.String())

	server, err := s.modelStore.GetServer(event.ServerName, true, true)
	if err != nil {
		logger.WithError(err).Errorf("Failed to get server %s", event.ServerName)
		return
	}

	if s.config.AutoScalingServerEnabled {
		if event.UpdateContext == coordinator.SERVER_SCALE_DOWN {
			if ok, replicas := shouldScaleDown(server, float32(s.config.PackThreshold)); ok {
				logger.Infof("Server %s is scaling down to %d", event.ServerName, replicas)
				s.sendServerScale(server, replicas)
			}
		} else if event.UpdateContext == coordinator.SERVER_SCALE_UP {
			if ok, replicas := shouldScaleUp(server); ok {
				logger.Infof("Server %s is scaling up to %d", event.ServerName, replicas)
				s.sendServerScale(server, replicas)
			}
		}
	}
}

func (s *SchedulerServer) StopSendServerEvents() {
	s.serverEventStream.mu.Lock()
	defer s.serverEventStream.mu.Unlock()
	for _, subscription := range s.serverEventStream.streams {
		close(subscription.fin)
	}
}

func (s *SchedulerServer) updateServerModelsStatus(evt coordinator.ModelEventMsg) error {
	logger := s.logger.WithField("func", "updateServerModelStatus")

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

	s.serverEventStream.pendingLock.Lock()
	// we are coalescing events so we only send one event (the latest status) per server
	s.serverEventStream.pendingEvents[modelVersion.Server()] = struct{}{}
	if s.serverEventStream.trigger == nil {
		s.serverEventStream.trigger = time.AfterFunc(defaultBatchWait, s.sendServerStatus)
	}
	s.serverEventStream.pendingLock.Unlock()

	return err
}

func (s *SchedulerServer) sendServerStatus() {
	logger := s.logger.WithField("func", "sendServerStatus")

	// Sending events may be slow, so allow a new batch to start building as we send.
	s.serverEventStream.pendingLock.Lock()
	s.serverEventStream.trigger = nil
	pendingServers := s.serverEventStream.pendingEvents
	s.serverEventStream.pendingEvents = map[string]struct{}{}
	s.serverEventStream.pendingLock.Unlock()

	// Inform subscriber
	s.serverEventStream.mu.Lock()
	defer s.serverEventStream.mu.Unlock()
	for serverName := range pendingServers {
		server, err := s.modelStore.GetServer(serverName, true, true)
		if err != nil {
			logger.Errorf("Failed to get server %s", serverName)
			continue
		}
		ssr := createServerStatusUpdateResponse(server)
		s.sendServerResponse(ssr)
	}
}

func (s *SchedulerServer) sendServerScale(server *store.ServerSnapshot, expectedReplicas uint32) {
	// TODO: should there be some sort of velocity check ?
	logger := s.logger.WithField("func", "sendServerScale")
	logger.Debugf("will attempt to scale servers to %d for %v", expectedReplicas, server.Name)

	ssr := createServerScaleResponse(server, expectedReplicas)
	s.sendServerResponse(ssr)
}

func (s *SchedulerServer) sendServerResponse(ssr *pb.ServerStatusResponse) {
	logger := s.logger.WithField("func", "sendServerResponse")
	for stream, subscription := range s.serverEventStream.streams {
		hasExpired, err := sendWithTimeout(func() error { return stream.Send(ssr) }, s.timeout)
		if hasExpired {
			// this should trigger a reconnect from the client
			close(subscription.fin)
			delete(s.serverEventStream.streams, stream)
		}
		if err != nil {
			logger.WithError(err).Errorf("Failed to send server status response to %s", subscription.name)
		}
	}
}

// initial send of server statuses to a new controller
func (s *SchedulerServer) sendCurrentServerStatuses(stream pb.Scheduler_ServerStatusServer) error {
	servers, err := s.modelStore.GetServers(true, true) // shallow, with model details
	if err != nil {
		return err
	}
	for _, server := range servers {
		ssr := createServerStatusUpdateResponse(server)
		_, err := sendWithTimeout(func() error { return stream.Send(ssr) }, s.timeout)
		if err != nil {
			return err
		}

	}
	return nil
}
