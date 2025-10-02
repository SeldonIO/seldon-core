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

	chainer "github.com/seldonio/seldon-core/apis/go/v2/mlops/chainer"
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	cr "github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/conflict-resolution"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

const (
	addPipelineStreamEventSource = "pipeline.store.addpipelinestream"
)

func (s *SchedulerServer) PipelineStatusEvent(ctx context.Context, message *chainer.PipelineUpdateStatusMessage) (*chainer.PipelineUpdateStatusResponse, error) {
	s.pipelineEventStream.mu.Lock()
	defer s.pipelineEventStream.mu.Unlock()

	logger := s.logger.WithField("func", "PipelineStatusEvent")
	logger.Infof("Received pipeline status event %s", message.String())

	var statusVal pipeline.PipelineStatus
	switch message.Update.Op {
	case chainer.PipelineUpdateMessage_Create:
		if message.Success {
			statusVal = pipeline.PipelineReady
		} else {
			statusVal = pipeline.PipelineFailed
		}
	case chainer.PipelineUpdateMessage_Delete:
		if message.Success {
			statusVal = pipeline.PipelineTerminated
		} else {
			statusVal = pipeline.PipelineFailed
		}
	}

	pipelineName := message.Update.Pipeline
	pipelineVersion := message.Update.Version
	stream := message.Update.Stream
	logger.Debugf(
		"Received pipeline update event from %s for pipeline %s:%d with status %s",
		stream, pipelineName, pipelineVersion, statusVal.String(),
	)

	confRes := s.pipelineEventStream.conflictResolutioner
	if cr.IsPipelineMessageOutdated(confRes, message) {
		logger.Debugf("Message for pipeline %s:%d is outdated, ignoring", pipelineName, pipelineVersion)
		return &chainer.PipelineUpdateStatusResponse{}, nil
	}

	confRes.UpdateStatus(pipelineName, stream, statusVal)
	pipelineStatusVal, reason := cr.GetPipelineStatus(confRes, pipelineName, message)

	switch pipelineStatusVal {
	case pipeline.PipelineTerminated:
		logger.Infof("Pipeline %s has been terminated, removing from conflict resolution and envoy", pipelineName)
		confRes.Delete(pipelineName)
	case pipeline.PipelineReady:
		// Once the pipeline is ready, send event for envoy to update the routes
		// with the streams that have the pipeline ready (some streams may have failed,
		// but we can still use the streams that are ready)
		serverNames := confRes.GetStreamsWithStatus(pipelineName, pipeline.PipelineReady)
		logger.Debugf("Pipeline %s is ready on streams %v, sending event for envoy", pipelineName, serverNames)
		s.sendPipelineStreamsEventMsg(
			&coordinator.PipelineEventMsg{PipelineName: pipelineName}, serverNames,
		)
	}

	err := s.pipelineHandler.SetPipelineGwPipelineState(
		message.Update.Pipeline, message.Update.Version, message.Update.Uid, pipelineStatusVal, reason, util.SourcePipelineStatusEvent,
	)
	if err != nil {
		logger.WithError(err).Errorf("Failed to update pipeline status for %s:%d (%s)", message.Update.Pipeline, message.Update.Version, message.Update.Uid)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &chainer.PipelineUpdateStatusResponse{}, nil
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

func (s *SchedulerServer) createPipelineDeletionMessage(pv *pipeline.PipelineVersion) *pb.PipelineStatusResponse {
	return &pb.PipelineStatusResponse{
		Operation:    pb.PipelineStatusResponse_PipelineDelete,
		PipelineName: pv.Name,
		Versions: []*pb.PipelineWithState{
			pipeline.CreatePipelineWithState(pv),
		},
	}
}

func (s *SchedulerServer) createPipelineCreationMessage(pv *pipeline.PipelineVersion) *pb.PipelineStatusResponse {
	return &pb.PipelineStatusResponse{
		Operation:    pb.PipelineStatusResponse_PipelineCreate,
		PipelineName: pv.Name,
		Versions: []*pb.PipelineWithState{
			pipeline.CreatePipelineWithState(pv),
		},
	}
}

func (s *SchedulerServer) pipelineGwRebalance() {
	s.pipelineEventStream.mu.Lock()
	defer s.pipelineEventStream.mu.Unlock()

	// get only the pipeline gateway streams
	streams := []*PipelineSubscription{}
	for _, subscription := range s.pipelineEventStream.streams {
		if subscription.isPipelineGateway {
			streams = append(streams, subscription)
		}
	}

	evts := s.pipelineHandler.GetAllPipelineGwRunningPipelineVersions()
	for _, event := range evts {
		pv, err := s.pipelineHandler.GetPipelineVersion(event.PipelineName, event.PipelineVersion, event.UID)
		if err != nil {
			s.logger.WithError(err).Errorf("Failed to get pipeline version for %s:%d (%s)", event.PipelineName, event.PipelineVersion, event.UID)
			continue
		}

		s.logger.Debugf(
			"Rebalancing pipeline %s:%d with pipeline gateway state %s",
			event.PipelineName, event.PipelineVersion, pv.State.PipelineGwStatus.String(),
		)

		switch len(streams) {
		case 0:
			s.pipelineGWRebalanceNoStreams(pv)
		default:
			s.pipelineGwRebalanceStreams(pv, streams)
		}
	}
}

func (s *SchedulerServer) pipelineGWRebalanceNoStreams(pv *pipeline.PipelineVersion) {
	// no pipeline gateway available, publish event for envoy
	s.sendPipelineStreamsEventMsg(
		&coordinator.PipelineEventMsg{PipelineName: pv.Name}, []string{},
	)

	pipelineState := pipeline.PipelineCreate
	if pv.State.PipelineGwStatus == pipeline.PipelineTerminating {
		// since there are no streams, we can directly set the state to terminated
		pipelineState = pipeline.PipelineTerminated
	}

	s.logger.Debugf(
		"No pipeline gateway available to handle pipeline %s, setting state to %s",
		pv.String(), pipelineState.String(),
	)

	if err := s.pipelineHandler.SetPipelineGwPipelineState(
		pv.Name,
		pv.Version,
		pv.UID,
		pipelineState,
		"No pipeline gateway available to handle pipeline",
		util.SourcePipelineStatusEvent,
	); err != nil {
		s.logger.WithError(err).Errorf("Failed to set pipeline gw state for %s", pv.String())
	}
}

func (s *SchedulerServer) invalidateEnvoyRoutes(pipelineName string, servers []string) {
	cr := s.pipelineEventStream.conflictResolutioner
	oldServers := cr.GetStreamsWithStatus(pipelineName, pipeline.PipelineReady)

	// find servers that are in both oldServers and servers
	commonServers := []string{}
	for _, oldServer := range oldServers {
		if contains(servers, oldServer) {
			commonServers = append(commonServers, oldServer)
		}
	}

	if len(commonServers) < len(oldServers) {
		s.logger.Debugf("Updated envoy routes for pipeline %s before rebalance to %v", pipelineName, commonServers)
		s.sendPipelineStreamsEventMsg(
			&coordinator.PipelineEventMsg{PipelineName: pipelineName}, commonServers,
		)
	}
}

func (s *SchedulerServer) pipelineGwRebalanceStreams(
	pv *pipeline.PipelineVersion, streams []*PipelineSubscription,
) {
	consumerBucketId := s.getPipelineGatewayBucketId(pv.Name)
	servers := s.pipelineGWLoadBalancer.GetServersForKey(consumerBucketId)
	s.logger.Debugf("Servers for pipeline %s: %v", pv.Name, servers)
	s.logger.Debug("Consumer bucket ID: ", consumerBucketId)

	confRes := s.pipelineEventStream.conflictResolutioner
	cr.CreateNewPipelineIteration(confRes, pv.Name, servers)

	// invalidate envoy routes if some servers are no longer valid
	s.invalidateEnvoyRoutes(pv.Name, servers)

	// send messages to each pipeline gateway stream
	for _, pipelineSubscription := range streams {
		s.logger.Debug("Processing pipeline subscription for ", pipelineSubscription.name)
		server := pipelineSubscription.name
		stream := pipelineSubscription.stream

		if contains(servers, server) {
			s.logger.Debug("pipeline-gateway replica contains pipeline, sending status update for ", server)

			var msg *pb.PipelineStatusResponse
			var err error

			if pv.State.PipelineGwStatus == pipeline.PipelineTerminating {
				s.logger.Debugf("Pipeline %s is terminating, sending deletion message", pv.Name)
				msg = s.createPipelineDeletionMessage(pv)
			} else {
				s.logger.Debugf("Pipeline %s is available or progressing, sending creation message", pv.Name)
				msg = s.createPipelineCreationMessage(pv)

				// set pipeline gw status to creating and display rebalance reason
				if err := s.pipelineHandler.SetPipelineGwPipelineState(
					pv.Name,
					pv.Version,
					pv.UID,
					pipeline.PipelineCreating,
					"Rebalance",
					util.SourcePipelineStatusEvent,
				); err != nil {
					s.logger.WithError(err).Errorf("Failed to set pipeline gw state for %s", pv.String())
				}
			}
			if err != nil {
				s.logger.WithError(err).Errorf("Failed to create pipelines status message for %s", pv.Name)
				continue
			}

			msg.Timestamp = confRes.GetTimestamp(pv.Name)
			if err := stream.Send(msg); err != nil {
				s.logger.WithError(err).Errorf("Failed to send create rebalance msg to pipeline %s", pv.Name)
			}
		} else {
			s.logger.Debugf("Server %s does not contain pipeline %s, sending deletion message", server, pv.Name)
			msg := s.createPipelineDeletionMessage(pv)
			msg.Timestamp = confRes.GetTimestamp(pv.Name)
			if err := stream.Send(msg); err != nil {
				s.logger.WithError(err).Errorf("Failed to send delete rebalance msg to pipeline %s", pv.Name)
			}
		}
	}
}

func (s *SchedulerServer) handlePipelineEvents(event coordinator.PipelineEventMsg) {
	logger := s.logger.WithField("func", "handlePipelineEvents")
	logger.Debugf("Received pipeline event %s", event.String())
	s.sendPipelineEvents(&event)
}

func (s *SchedulerServer) sendPipelineEvents(event *coordinator.PipelineEventMsg) {
	logger := s.logger.WithField("func", "sendPipelineEvents")
	if event.ExperimentUpdate {
		return
	}

	s.pipelineEventStream.mu.Lock()
	defer s.pipelineEventStream.mu.Unlock()

	// create a pipeline status response message based on the pipeline version
	pv, err := s.pipelineHandler.GetPipelineVersion(event.PipelineName, event.PipelineVersion, event.UID)
	if err != nil {
		logger.WithError(err).Errorf("Failed to get pipeline from event %s", event.String())
		return
	}
	logger.Debugf(
		"Handling pipeline event for %s with state %v, pipelinegw %v, models %t, source %s",
		event.String(), pv.State.Status, pv.State.PipelineGwStatus, pv.State.ModelsReady, event.Source,
	)

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

	// message to send to non-pipeline gateway streams (i.e. controller)
	// we have to update the controller with all the pipeline status changes
	// that includes messages originating from the `pipelineStatusEventSource`
	status := &pb.PipelineStatusResponse{
		PipelineName: pv.Name,
		Versions: []*pb.PipelineWithState{
			pipeline.CreatePipelineWithState(pv),
		},
	}
	s.sendPipelineEventsToStreams(event, status, otherStreams)

	// don't consider events from pipeline status or chainer server
	var pipelineEventSources = map[string]struct{}{
		util.SourcePipelineStatusEvent: {},
		util.SourceChainerServer:       {},
	}
	if _, ok := pipelineEventSources[event.Source]; ok {
		return
	}

	// if deletion process was triggered, we remove the pipeline from envoy
	if pv.State.PipelineGwStatus == pipeline.PipelineTerminate {
		s.sendPipelineStreamsEventMsg(
			&coordinator.PipelineEventMsg{PipelineName: pv.Name}, []string{},
		)
	}

	if len(pipelineGwStreams) == 0 && pv.State.PipelineGwStatus != pipeline.PipelineTerminated {
		errMsg := "No pipeline gateway available to handle pipeline"
		logger.WithField("pipeline", pv.Name).Warn(errMsg)

		pipelineGwStatus := pv.State.PipelineGwStatus
		if pipelineGwStatus == pipeline.PipelineTerminating || pipelineGwStatus == pipeline.PipelineTerminate {
			pipelineGwStatus = pipeline.PipelineTerminated
		}

		if err := s.pipelineHandler.SetPipelineGwPipelineState(
			pv.Name,
			pv.Version,
			pv.UID,
			pipelineGwStatus,
			errMsg,
			util.SourcePipelineStatusEvent,
		); err != nil {
			logger.
				WithError(err).
				WithField("pipeline", pv.Name).
				WithField("status", pipelineGwStatus).
				Errorf("Failed to set pipeline gw state")
		}
		return
	}

	switch pv.State.PipelineGwStatus {
	case pipeline.PipelineCreate:
		logger.Debug("Pipeline is being created, sending creation message")
		if err := s.pipelineHandler.SetPipelineGwPipelineState(
			pv.Name,
			pv.Version,
			pv.UID,
			pipeline.PipelineCreating,
			"",
			util.SourcePipelineStatusEvent,
		); err != nil {
			logger.WithError(err).Errorf("Failed to set pipeline gw state for %s", pv.String())
		}
		status = s.createPipelineCreationMessage(pv)
	case pipeline.PipelineTerminate:
		logger.Debug("Pipeline is being terminated, sending deletion message")
		if err := s.pipelineHandler.SetPipelineGwPipelineState(
			pv.Name,
			pv.Version,
			pv.UID,
			pipeline.PipelineTerminating,
			"",
			util.SourcePipelineStatusEvent,
		); err != nil {
			logger.WithError(err).Errorf("Failed to set pipeline gw state for %s", pv.String())
		}
		status = s.createPipelineDeletionMessage(pv)
	}

	// for pipeline gateway streams, we need to assign a timestamp
	// to we can identify the latest message (for conflict resolution)
	s.sendPipelineEventsToStreamWithTimestamp(event, status, pipelineGwStreams)
}

func (s *SchedulerServer) sendPipelineStreamsEventMsg(event *coordinator.PipelineEventMsg, streamNames []string) {
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
		PipelineEventMsg: *event,
		StreamNames:      streamNames,
		StreamIps:        streamIps,
	}
	s.eventHub.PublishPipelineStreamsEvent(addPipelineStreamEventSource, eventMsg)
}

func (s *SchedulerServer) sendPipelineEventsToStreams(
	event *coordinator.PipelineEventMsg,
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

func (s *SchedulerServer) sendPipelineEventsToStreamWithTimestamp(
	event *coordinator.PipelineEventMsg,
	status *pb.PipelineStatusResponse,
	streams map[pb.Scheduler_SubscribePipelineStatusServer]*PipelineSubscription,
) {
	streamNames := make([]string, 0, len(streams))
	for _, subscription := range streams {
		streamNames = append(streamNames, subscription.name)
	}

	// assign a timestamp to the message
	confRes := s.pipelineEventStream.conflictResolutioner
	cr.CreateNewPipelineIteration(confRes, event.PipelineName, streamNames)
	status.Timestamp = confRes.GetTimestamp(event.PipelineName)

	s.sendPipelineEventsToStreams(event, status, streams)
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
