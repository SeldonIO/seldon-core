/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package server

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/experiment"
)

func TestExperimentStatusStream(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name    string
		loadReq *experiment.Experiment
		server  *SchedulerServer
		err     bool
	}

	tests := []test{
		{
			name: "experiment ok",
			loadReq: &experiment.Experiment{
				Name: "foo",
				Mirror: &experiment.Mirror{
					Name: "bar",
				},
			},
			server: &SchedulerServer{
				experimentServer: experiment.NewExperimentServer(log.New(), nil, nil, nil),
				logger:           log.New(),
				timeout:          10 * time.Millisecond,
			},
		},
		{
			name: "timeout",
			loadReq: &experiment.Experiment{
				Name: "foo",
				Mirror: &experiment.Mirror{
					Name: "bar",
				},
			},
			server: &SchedulerServer{
				experimentServer: experiment.NewExperimentServer(log.New(), nil, nil, nil),
				logger:           log.New(),
				timeout:          1 * time.Millisecond,
			},
			err: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.loadReq != nil {
				err := test.server.experimentServer.StartExperiment(test.loadReq)
				g.Expect(err).To(BeNil())
			}

			stream := newStubExperimentStatusServer(1, 5*time.Millisecond)
			err := test.server.sendCurrentExperimentStatuses(stream)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())

				var esr *pb.ExperimentStatusResponse
				select {
				case next := <-stream.msgs:
					esr = next
				default:
					t.Fail()
				}

				g.Expect(esr).ToNot(BeNil())
				g.Expect(esr.ExperimentName).To(Equal("foo"))
			}
		})
	}
}

func TestExperimentStatusEvents(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name    string
		loadReq *experiment.Experiment
		timeout time.Duration
		err     bool
	}

	tests := []test{
		{
			name: "experiment ok",
			loadReq: &experiment.Experiment{
				Name: "foo",
				Mirror: &experiment.Mirror{
					Name: "bar",
				},
			},
			timeout: 10 * time.Millisecond,
			err:     false,
		},
		{
			name: "timeout",
			loadReq: &experiment.Experiment{
				Name: "foo",
				Mirror: &experiment.Mirror{
					Name: "bar",
				},
			},
			timeout: 1 * time.Millisecond,
			err:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, hub := createTestScheduler()
			s.timeout = test.timeout
			if test.loadReq != nil {
				err := s.experimentServer.StartExperiment(test.loadReq)
				g.Expect(err).To(BeNil())
			}

			stream := newStubExperimentStatusServer(1, 5*time.Millisecond)
			s.experimentEventStream.streams[stream] = &ExperimentSubscription{
				name:   "dummy",
				stream: stream,
				fin:    make(chan bool),
			}
			g.Expect(s.experimentEventStream.streams[stream]).ToNot(BeNil())
			hub.PublishExperimentEvent(experimentEventHandlerName, coordinator.ExperimentEventMsg{
				ExperimentName: "foo"})

			// to allow events to propagate
			time.Sleep(500 * time.Millisecond)

			if test.err {
				g.Expect(s.experimentEventStream.streams).To(HaveLen(0))
			} else {

				var esr *pb.ExperimentStatusResponse
				select {
				case next := <-stream.msgs:
					esr = next
				default:
					t.Fail()
				}

				g.Expect(esr).ToNot(BeNil())
				g.Expect(esr.ExperimentName).To(Equal("foo"))
				g.Expect(s.experimentEventStream.streams).To(HaveLen(1))
			}
		})
	}
}
