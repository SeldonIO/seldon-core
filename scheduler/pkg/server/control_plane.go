/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package server

import (
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

func (s *SchedulerServer) SubscribeControlPlane(req *pb.ControlPlaneSubscriptionRequest, stream pb.Scheduler_SubscribeControlPlaneServer) error {
	logger := s.logger.WithField("func", "SubscribeControlPlane")
	logger.Infof("Received subscribe request from %s", req.GetSubscriberName())

	err := s.sendStartServerStreamMarker(stream)
	if err != nil {
		logger.WithError(err).Errorf("Failed to send start marker to %s", req.GetSubscriberName())
		return err
	}

	s.synchroniser.WaitReady()

	err = s.sendResourcesMarker(stream)
	if err != nil {
		logger.WithError(err).Errorf("Failed to send resources marker to %s", req.GetSubscriberName())
		return err
	}

	fin := make(chan bool)
	s.controlPlaneStream.mu.Lock()
	s.controlPlaneStream.streams[stream] = &ControlPlaneSubsription{
		name:   req.GetSubscriberName(),
		stream: stream,
		fin:    fin,
	}
	s.controlPlaneStream.mu.Unlock()

	ctx := stream.Context()
	// Keep this scope alive because once this scope exits - the stream is closed
	for {
		select {
		case <-fin:
			logger.Infof("Closing stream for %s", req.GetSubscriberName())
			return nil
		case <-ctx.Done():
			logger.Infof("Stream disconnected %s", req.GetSubscriberName())
			s.controlPlaneStream.mu.Lock()
			delete(s.controlPlaneStream.streams, stream)
			s.controlPlaneStream.mu.Unlock()
			return nil
		}
	}
}

func (s *SchedulerServer) StopSendControlPlaneEvents() {
	s.controlPlaneStream.mu.Lock()
	defer s.controlPlaneStream.mu.Unlock()
	for _, subscription := range s.controlPlaneStream.streams {
		close(subscription.fin)
	}
}

// this is to mark the initial start of a new stream (at application level)
// as otherwise the other side sometimes doesnt know if the scheduler has established a new stream explicitly
func (s *SchedulerServer) sendStartServerStreamMarker(stream pb.Scheduler_SubscribeControlPlaneServer) error {
	ssr := &pb.ControlPlaneResponse{Event: pb.ControlPlaneResponse_SEND_SERVERS}
	_, err := sendWithTimeout(func() error { return stream.Send(ssr) }, s.timeout)
	if err != nil {
		return err
	}
	return nil
}

// this is to mark a stage to send resources (models, pipelines, experiments) from the controller
func (s *SchedulerServer) sendResourcesMarker(stream pb.Scheduler_SubscribeControlPlaneServer) error {
	ssr := &pb.ControlPlaneResponse{Event: pb.ControlPlaneResponse_SEND_RESOURCES}
	_, err := sendWithTimeout(func() error { return stream.Send(ssr) }, s.timeout)
	if err != nil {
		return err
	}
	return nil
}
