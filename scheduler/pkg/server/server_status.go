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
	"time"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	cr "github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/conflict-resolution"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

const (
	modelStatusEventSource = "model-status-server"
)

func (s *SchedulerServer) ModelStatusEvent(ctx context.Context, message *pb.ModelUpdateStatusMessage) (*pb.ModelUpdateStatusResponse, error) {
	s.modelEventStream.mu.Lock()
	defer s.modelEventStream.mu.Unlock()

	logger := s.logger.WithField("func", "ModelStatusEvent")

	var statusVal store.ModelState
	switch message.Update.Op {
	case pb.ModelUpdateMessage_Create:
		if message.Success {
			statusVal = store.ModelAvailable
		} else {
			statusVal = store.ModelFailed
		}
	case pb.ModelUpdateMessage_Delete:
		if message.Success {
			statusVal = store.ModelTerminated
		} else {
			statusVal = store.ModelTerminateFailed
		}
	}

	modelName := message.Update.Model
	modelVersion := message.Update.Version
	stream := message.Update.Stream
	logger.Debugf(
		"Received model update event from %s for model %s:%d with status %s",
		stream, modelName, modelVersion, statusVal.String(),
	)

	confRes := s.modelEventStream.conflictResolutioner
	if cr.IsModelMessageOutdated(confRes, message) {
		logger.Debugf("Message for model %s:%d is outdated, ignoring", modelName, modelVersion)
		return &pb.ModelUpdateStatusResponse{}, nil
	}

	confRes.UpdateStatus(modelName, stream, statusVal)
	modelStatusVal, reason := cr.GetModelStatus(confRes, modelName, message)
	if modelStatusVal == store.ModelTerminated {
		confRes.Delete(modelName)
	}

	err := s.modelStore.SetModelGwModelState(
		message.Update.Model,
		message.Update.Version,
		modelStatusVal,
		reason,
		modelStatusEventSource,
	)
	if err != nil {
		logger.WithError(err).Errorf("Failed to set model state for %s version %d to %s", message.Update.Model, message.Update.Version, statusVal.String())
		return nil, err
	}

	return &pb.ModelUpdateStatusResponse{}, nil
}

func (s *SchedulerServer) SubscribeModelStatus(req *pb.ModelSubscriptionRequest, stream pb.Scheduler_SubscribeModelStatusServer) error {
	logger := s.logger.WithField("func", "SubscribeModelStatus")
	logger.Infof("Received subscribe request from %s", req.GetSubscriberName())

	s.synchroniser.WaitReady()

	fin := make(chan bool)

	s.modelEventStream.mu.Lock()
	s.modelEventStream.streams[stream] = &ModelSubscription{
		name:           req.GetSubscriberName(),
		stream:         stream,
		fin:            fin,
		isModelGateway: req.IsModelGateway,
	}
	if req.IsModelGateway {
		s.modelGwLoadBalancer.AddServer(req.GetSubscriberName())
	}
	s.modelEventStream.mu.Unlock()

	if req.IsModelGateway {
		// rebalance the streams when a new subscription is added
		s.modelGwRebalance()
	} else {
		// update controller with current model statuses
		err := s.sendCurrentModelStatuses(stream)
		if err != nil {
			logger.WithError(err).Errorf("Failed to send current model statuses to %s", req.GetSubscriberName())
			return err
		}
	}

	ctx := stream.Context()
	// Keep this scope alive because once this scope exits - the stream is closed
	select {
	case <-fin:
		logger.Infof("Closing model status stream for %s", req.GetSubscriberName())
	case <-ctx.Done():
		logger.Infof("Model status stream disconnected: %s", req.GetSubscriberName())
		s.modelEventStream.mu.Lock()
		delete(s.modelEventStream.streams, stream)
		if req.IsModelGateway {
			s.modelGwLoadBalancer.RemoveServer(req.GetSubscriberName())
		}
		s.modelEventStream.mu.Unlock()
		// rebalance the streams when a subscription is removed
		if req.IsModelGateway {
			s.modelGwRebalance()
		}
		logger.Infof("Model status stream %s removed", req.GetSubscriberName())
	}

	return nil
}

func (s *SchedulerServer) sendCurrentModelStatuses(stream pb.Scheduler_SubscribeModelStatusServer) error {
	s.modelEventStream.mu.Lock()
	defer s.modelEventStream.mu.Unlock()

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
		_, err = sendWithTimeout(func() error {
			select {
			case <-stream.Context().Done():
				return stream.Context().Err()
			default:
				return stream.Send(ms)
			}
		}, s.timeout)
		if err != nil {
			return err
		}
	}
	return nil
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func (s *SchedulerServer) GetAllRunningModels() []*store.ModelSnapshot {
	var runningModels []*store.ModelSnapshot
	modelNames := s.modelStore.GetAllModels()

	for _, modelName := range modelNames {
		model, err := s.modelStore.GetModel(modelName)
		if err != nil {
			s.logger.WithError(err).Errorf("Failed to get model %s for running models", modelName)
			continue
		}
		if model.GetLatest() == nil {
			s.logger.Warnf("Model %s has no versions, skipping running models", modelName)
			continue
		}

		modelState := model.GetLatest().ModelState()
		runningStates := map[store.ModelState]struct{}{
			store.ModelCreate:      {},
			store.ModelProgressing: {},
			store.ModelAvailable:   {},
			store.ModelTerminating: {},
		}

		if _, ok := runningStates[modelState.ModelGwState]; ok {
			runningModels = append(runningModels, model)
		}
	}
	return runningModels
}

func (s *SchedulerServer) createModelDeletionMessage(model *store.ModelSnapshot, keepTopics bool) (*pb.ModelStatusResponse, error) {
	ms, err := s.modelStatusImpl(model, false)
	if err != nil {
		return nil, err
	}
	ms.Operation = pb.ModelStatusResponse_ModelDelete
	ms.KeepTopics = keepTopics
	return ms, nil
}

func (s *SchedulerServer) createModelCreationMessage(model *store.ModelSnapshot) (*pb.ModelStatusResponse, error) {
	ms, err := s.modelStatusImpl(model, false)
	if err != nil {
		return nil, err
	}
	ms.Operation = pb.ModelStatusResponse_ModelCreate
	return ms, nil
}

func (s *SchedulerServer) modelGwRebalance() {
	s.modelEventStream.mu.Lock()
	defer s.modelEventStream.mu.Unlock()

	runningModels := s.GetAllRunningModels()
	s.logger.Debugf("Rebalancing model gateways for running models: %v", runningModels)

	// get only the model gateway streams
	streams := []*ModelSubscription{}
	for _, modelSubscription := range s.modelEventStream.streams {
		if modelSubscription.isModelGateway {
			streams = append(streams, modelSubscription)
		}
	}

	for _, model := range runningModels {
		switch len(streams) {
		case 0:
			s.modelGwRebalanceNoStream(model)
		default:
			s.modelGwReblanceStreams(model)
		}
	}
}

func (s *SchedulerServer) modelGwRebalanceNoStream(model *store.ModelSnapshot) {
	modelState := store.ModelCreate
	if model.GetLatest().ModelState().ModelGwState == store.ModelTerminating {
		modelState = store.ModelTerminated
	}

	s.logger.Debugf(
		"No model gateway available to handle model %s, setting state to %s",
		model.Name, modelState.String(),
	)

	if err := s.modelStore.SetModelGwModelState(
		model.Name,
		model.GetLatest().GetVersion(),
		modelState,
		"No model gateway available to handle model",
		modelStatusEventSource,
	); err != nil {
		s.logger.WithError(err).Errorf("Failed to set model-gw state for %s", model.Name)
	}
}

func (s *SchedulerServer) modelGwReblanceStreams(model *store.ModelSnapshot) {
	consumerBucketId := s.getModelGatewayBucketId(model.Name)
	s.logger.Debugf("Rebalancing model %s with consumber bucket id %s", model.Name, consumerBucketId)

	servers := s.modelGwLoadBalancer.GetServersForKey(consumerBucketId)
	s.logger.Debugf("Servers for model %s: %v", model.Name, servers)

	confRes := s.modelEventStream.conflictResolutioner
	cr.CreateNewModelIteration(confRes, model.Name, servers)

	for _, modelSubscription := range s.modelEventStream.streams {
		if !modelSubscription.isModelGateway {
			s.logger.Debugf("Skipping non-model gateway stream for %s", modelSubscription.name)
			continue
		}

		s.logger.Debug("Processing model subscription for ", modelSubscription.name)
		server := modelSubscription.name
		stream := modelSubscription.stream

		if contains(servers, server) {
			s.logger.Debug("Server contains model, sending status update for: ", server)

			state := model.GetLatest().ModelState().ModelGwState
			var msg *pb.ModelStatusResponse
			var err error

			if state == store.ModelTerminating {
				s.logger.Debugf("Model %s is terminating, sending deletion message", model.Name)
				msg, err = s.createModelDeletionMessage(model, false)
			} else {
				s.logger.Debugf("Model %s is available or progressing, sending creation message", model.Name)
				msg, err = s.createModelCreationMessage(model)

				// set modelgw state to progressing and display rebalance reason
				if err := s.modelStore.SetModelGwModelState(
					model.Name,
					model.GetLatest().GetVersion(),
					store.ModelProgressing,
					"Rebalance",
					modelStatusEventSource,
				); err != nil {
					s.logger.WithError(err).Errorf("Failed to set pipeline gw state for %s", model.Name)
				}
			}
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to create model status message for %s", model.Name)
				continue
			}
			msg.Timestamp = confRes.GetTimestamp(model.Name)

			select {
			case <-stream.Context().Done():
				s.logger.WithError(err).Errorf("Failed to send create rebalance msg to model %s stream ctx cancelled", model.Name)
			default:
				if err := stream.Send(msg); err != nil {
					s.logger.WithError(err).Errorf("Failed to send create rebalance msg to model %s", model.Name)
				}
			}

		} else {
			s.logger.Debugf("Server %s does not contain model %s, sending deletion message", server, model.Name)
			msg, err := s.createModelDeletionMessage(model, true)
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to create model deletion message for %s", model.Name)
				continue
			}
			s.logger.Debugf("Sending deletion message for model %s to server %s", model.Name, server)
			msg.Timestamp = confRes.GetTimestamp(model.Name)

			select {
			case <-stream.Context().Done():
				s.logger.WithError(err).Errorf("Failed to send deletion message for %s stream ctx cancelled", model.Name)
			default:
				if err := stream.Send(msg); err != nil {
					s.logger.WithError(err).Errorf("Failed to send delete rebalance msg to model %s", model.Name)
				}
			}
		}
	}
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

func (s *SchedulerServer) sendModelStatusEventToStreams(
	evt coordinator.ModelEventMsg,
	ms *pb.ModelStatusResponse,
	streams map[pb.Scheduler_SubscribeModelStatusServer]*ModelSubscription,
) {
	logger := s.logger.WithField("func", "sendModelStatusEventToStreams")
	for stream, subscription := range streams {
		hasExpired, err := sendWithTimeout(func() error {
			select {
			case <-stream.Context().Done():
				return stream.Context().Err()
			default:
				return stream.Send(ms)
			}
		}, s.timeout)
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

func (s *SchedulerServer) sendModelStatusEvent(evt coordinator.ModelEventMsg) error {
	s.modelEventStream.mu.Lock()
	defer s.modelEventStream.mu.Unlock()

	logger := s.logger.WithField("func", "sendModelStatusEvent")
	model, err := s.modelStore.GetModel(evt.ModelName)
	if err != nil {
		return err
	}

	if model.GetLatest() == nil {
		logger.Warnf("Failed to find latest model version for %s so ignoring event", evt.String())
		return nil
	}

	if model.GetLatest().GetVersion() != evt.ModelVersion {
		logger.Warnf("Latest model version %d does not match event version %d for %s so ignoring event", model.GetLatest().GetVersion(), evt.ModelVersion, evt.String())
		return nil
	}

	// find the modelgw servers that should receive this event
	consumerBucketId := s.getModelGatewayBucketId(evt.ModelName)
	servers := s.modelGwLoadBalancer.GetServersForKey(consumerBucketId)

	// split streams into model gateway and other streams
	modelGwStreams := make(map[pb.Scheduler_SubscribeModelStatusServer]*ModelSubscription)
	otherStreams := make(map[pb.Scheduler_SubscribeModelStatusServer]*ModelSubscription)
	for stream, subscription := range s.modelEventStream.streams {
		if !subscription.isModelGateway {
			otherStreams[stream] = subscription
		} else if contains(servers, subscription.name) {
			modelGwStreams[stream] = subscription
		}
	}

	ms, err := s.modelStatusImpl(model, false)
	if err != nil {
		logger.WithError(err).Errorf("Failed to create model status for model %s", evt.String())
		return err
	}

	// send to all other streams
	s.sendModelStatusEventToStreams(evt, ms, otherStreams)

	// send to model gateway streams only if the message
	// is not an ack from the model gateway itself
	if evt.Source == modelStatusEventSource {
		return nil
	}

	modelState := model.GetLatest().ModelState()
	if len(modelGwStreams) == 0 && modelState.ModelGwState != store.ModelTerminated {
		// handle case where we don't have any model-gateway streams
		errMsg := "No model gateway available to handle model"
		logger.WithField("model", model.Name).Warn(errMsg)

		modelGwState := modelState.ModelGwState
		if modelState.ModelGwState == store.ModelTerminate || modelState.ModelGwState == store.ModelTerminating {
			modelGwState = store.ModelTerminated
		}

		if err := s.modelStore.SetModelGwModelState(
			model.Name,
			model.GetLatest().GetVersion(),
			modelGwState,
			errMsg,
			modelStatusEventSource,
		); err != nil {
			logger.
				WithError(err).
				WithField("model", model.Name).
				WithField("modelGwState", modelGwState).
				Error("failed to set model state")
		}
		return nil
	}

	switch modelState.ModelGwState {
	case store.ModelCreate:
		logger.Debugf("Model %s is in create state, sending creation message", model.Name)
		if err := s.modelStore.SetModelGwModelState(
			model.Name,
			model.GetLatest().GetVersion(),
			store.ModelProgressing,
			"Model is being loaded onto model gateway",
			modelStatusEventSource,
		); err != nil {
			logger.
				WithError(err).
				WithField("model", model.Name).
				Error("failed to set model state to progressing")
		}

		ms, err = s.createModelCreationMessage(model)
		if err != nil {
			logger.WithError(err).Errorf("Failed to create model creation message for %s", model.Name)
			return err
		}

		// send message to model gateway streams
		s.sendModelStatusEventToStreamsWithTimestamp(evt, ms, modelGwStreams)
	case store.ModelTerminate:
		logger.Debugf("Model %s is in terminate state, sending deletion message", model.Name)
		if err := s.modelStore.SetModelGwModelState(
			model.Name,
			model.GetLatest().GetVersion(),
			store.ModelTerminating,
			"Model is being unloaded from model gateway",
			modelStatusEventSource,
		); err != nil {
			logger.
				WithError(err).
				WithField("model", model.Name).
				Error("failed to set model state to terminating")
		}

		ms, err = s.createModelDeletionMessage(model, false)
		if err != nil {
			logger.WithError(err).Errorf("Failed to create model deletion message for %s", model.Name)
			return err
		}

		// send message to model gateway streams
		s.sendModelStatusEventToStreamsWithTimestamp(evt, ms, modelGwStreams)
	}
	return nil
}

func (s *SchedulerServer) sendModelStatusEventToStreamsWithTimestamp(
	evt coordinator.ModelEventMsg,
	ms *pb.ModelStatusResponse,
	streams map[pb.Scheduler_SubscribeModelStatusServer]*ModelSubscription,
) {
	// send message to model gateway streams
	streamNames := make([]string, 0, len(streams))
	for _, subscription := range streams {
		streamNames = append(streamNames, subscription.name)
	}

	// assign a new timestamp to the message
	confRes := s.modelEventStream.conflictResolutioner
	cr.CreateNewModelIteration(confRes, evt.ModelName, streamNames)
	ms.Timestamp = confRes.GetTimestamp(evt.ModelName)

	s.sendModelStatusEventToStreams(evt, ms, streams)
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
	select {
	case <-fin:
		logger.Infof("Closing server stream for %s", req.GetSubscriberName())
	case <-ctx.Done():
		logger.Infof("Server stream disconnected %s", req.GetSubscriberName())
		s.serverEventStream.mu.Lock()
		delete(s.serverEventStream.streams, stream)
		s.serverEventStream.mu.Unlock()
		logger.Infof("Removed server stream %s from map", req.GetSubscriberName())
	}

	return nil
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
		hasExpired, err := sendWithTimeout(func() error {
			select {
			case <-stream.Context().Done():
				return stream.Context().Err()
			default:
				return stream.Send(ssr)
			}
		}, s.timeout)
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
		_, err := sendWithTimeout(func() error {
			select {
			case <-stream.Context().Done():
				return stream.Context().Err()
			default:
				return stream.Send(ssr)
			}
		}, s.timeout)
		if err != nil {
			return err
		}

	}
	return nil
}

func (s *SchedulerServer) getModelGatewayBucketId(modelName string) string {
	return util.GetKafkaConsumerName(
		s.consumerGroupConfig.namespace,
		s.consumerGroupConfig.consumerGroupIdPrefix,
		modelName,
		modelGatewayConsumerNamePrefix,
		s.consumerGroupConfig.modelGatewayMaxNumConsumers,
	)
}
