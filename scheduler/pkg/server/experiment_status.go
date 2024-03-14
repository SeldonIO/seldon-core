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
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/experiment"
)

func (s *SchedulerServer) SubscribeExperimentStatus(req *pb.ExperimentSubscriptionRequest, stream pb.Scheduler_SubscribeExperimentStatusServer) error {
	logger := s.logger.WithField("func", "SubscribeExperimentStatus")
	logger.Infof("Received subscribe request from %s", req.GetSubscriberName())

	err := s.sendCurrentExperimentStatuses(stream)
	if err != nil {
		logger.WithError(err).Errorf("Failed to send current experiment statuses to %s", req.GetSubscriberName())
		return err
	}

	fin := make(chan bool)

	s.experimentEventStream.mu.Lock()
	s.experimentEventStream.streams[stream] = &ExperimentSubscription{
		name:   req.GetSubscriberName(),
		stream: stream,
		fin:    fin,
	}
	s.experimentEventStream.mu.Unlock()

	ctx := stream.Context()
	// Keep this scope alive because once this scope exits - the stream is closed
	for {
		select {
		case <-fin:
			logger.Infof("Closing stream for %s", req.GetSubscriberName())
			return nil
		case <-ctx.Done():
			logger.Infof("Stream disconnected %s", req.GetSubscriberName())
			s.experimentEventStream.mu.Lock()
			delete(s.experimentEventStream.streams, stream)
			s.experimentEventStream.mu.Unlock()
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

func asKubernetesMetaFromExperiment(meta *experiment.KubernetesMeta) *pb.KubernetesMeta {
	if meta != nil {
		return &pb.KubernetesMeta{
			Namespace:  meta.Namespace,
			Generation: meta.Generation,
		}
	}
	return nil
}

func (s *SchedulerServer) sendCurrentExperimentStatuses(stream pb.Scheduler_ExperimentStatusServer) error {
	experiments, err := s.experimentServer.GetExperiments()
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, err.Error())
	}
	for _, exp := range experiments {
		msg := &pb.ExperimentStatusResponse{
			ExperimentName:    exp.Name,
			Active:            exp.Active,
			CandidatesReady:   exp.AreCandidatesReady(),
			MirrorReady:       exp.IsMirrorReady(),
			StatusDescription: exp.StatusDescription,
			KubernetesMeta:    asKubernetesMetaFromExperiment(exp.KubernetesMeta),
		}
		_, err := sentWithTimeout(func() error { return stream.Send(msg) }, s.timeout)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SchedulerServer) handleExperimentEvents(event coordinator.ExperimentEventMsg) {
	logger := s.logger.WithField("func", "handleExperimentEvents")
	logger.Debugf("Received experiment event %s", event.String())
	if event.Status != nil {
		s.sendExperimentStatus(event)
	}
}

func (s *SchedulerServer) sendExperimentStatus(event coordinator.ExperimentEventMsg) {
	logger := s.logger.WithField("func", "sendExperimentStatus")
	s.experimentEventStream.mu.Lock()
	defer s.experimentEventStream.mu.Unlock()
	for stream, subscription := range s.experimentEventStream.streams {
		msg := &pb.ExperimentStatusResponse{
			ExperimentName:    event.ExperimentName,
			Active:            event.Status.Active,
			CandidatesReady:   event.Status.CandidatesReady,
			MirrorReady:       event.Status.MirrorReady,
			StatusDescription: event.Status.StatusDescription,
			KubernetesMeta:    asKubernetesMeta(event),
		}
		hasExpired, err := sentWithTimeout(func() error { return stream.Send(msg) }, s.timeout)
		if hasExpired {
			// this should trigger a reconnect from the client
			close(subscription.fin)
			delete(s.experimentEventStream.streams, stream)
		}
		if err != nil {
			logger.WithError(err).Errorf("Failed to send experiment status event to %s for %s", subscription.name, event.String())
		}
	}
}

func (s *SchedulerServer) StopSendExperimentEvents() {
	for _, subscription := range s.experimentEventStream.streams {
		close(subscription.fin)
	}
}
