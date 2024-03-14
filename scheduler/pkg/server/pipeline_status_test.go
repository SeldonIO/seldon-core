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
	scheduler2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/scheduler"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/experiment"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
)

func TestPipelineStatusStream(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name    string
		loadReq *pb.LoadPipelineRequest
		server  *SchedulerServer
		err     bool
	}

	tests := []test{
		{
			name: "pipeline ok",
			loadReq: &pb.LoadPipelineRequest{
				Pipeline: &pb.Pipeline{
					Name:    "foo",
					Version: 1,
					Uid:     "x",
					Steps: []*pb.PipelineStep{
						{
							Name: "a",
						},
						{
							Name:   "b",
							Inputs: []string{"a.outputs"},
						},
					},
				},
			},
			server: &SchedulerServer{
				pipelineHandler: pipeline.NewPipelineStore(log.New(), nil, nil),
				logger:          log.New(),
				timeout:         10 * time.Millisecond,
			},
		},
		{
			name: "timeout",
			loadReq: &pb.LoadPipelineRequest{
				Pipeline: &pb.Pipeline{
					Name:    "foo",
					Version: 1,
					Uid:     "x",
					Steps: []*pb.PipelineStep{
						{
							Name: "a",
						},
						{
							Name:   "b",
							Inputs: []string{"a.outputs"},
						},
					},
				},
			},
			server: &SchedulerServer{
				pipelineHandler: pipeline.NewPipelineStore(log.New(), nil, nil),
				logger:          log.New(),
				timeout:         1 * time.Millisecond,
			},
			err: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.loadReq != nil {
				err := test.server.pipelineHandler.AddPipeline(test.loadReq.Pipeline)
				g.Expect(err).To(BeNil())
			}

			stream := newStubPipelineStatusServer(1, 5*time.Millisecond)
			err := test.server.sendCurrentPipelineStatuses(stream, false)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())

				var psr *pb.PipelineStatusResponse
				select {
				case next := <-stream.msgs:
					psr = next
				default:
					t.Fail()
				}

				g.Expect(psr).ToNot(BeNil())
				g.Expect(psr.Versions).To(HaveLen(1))
				g.Expect(psr.Versions[0].State.Status).To(Equal(pb.PipelineVersionState_PipelineCreate))
			}
		})
	}
}

func TestPipelineStatusEvents(t *testing.T) {
	g := NewGomegaWithT(t)

	createTestScheduler := func() (*SchedulerServer, *coordinator.EventHub) {
		logger := log.New()
		logger.SetLevel(log.WarnLevel)

		eventHub, err := coordinator.NewEventHub(logger)
		g.Expect(err).To(BeNil())

		schedulerStore := store.NewMemoryStore(logger, store.NewLocalSchedulerStore(), eventHub)
		experimentServer := experiment.NewExperimentServer(logger, eventHub, nil, nil)
		pipelineServer := pipeline.NewPipelineStore(logger, eventHub, schedulerStore)

		scheduler := scheduler2.NewSimpleScheduler(
			logger,
			schedulerStore,
			scheduler2.DefaultSchedulerConfig(schedulerStore),
		)
		s := NewSchedulerServer(logger, schedulerStore, experimentServer, pipelineServer, scheduler, eventHub)

		return s, eventHub
	}
	type test struct {
		name    string
		loadReq *pb.LoadPipelineRequest
		timeout time.Duration
		err     bool
	}

	tests := []test{
		{
			name: "pipeline ok",
			loadReq: &pb.LoadPipelineRequest{
				Pipeline: &pb.Pipeline{
					Name:    "foo",
					Version: 1,
					Uid:     "x",
					Steps: []*pb.PipelineStep{
						{
							Name: "a",
						},
						{
							Name:   "b",
							Inputs: []string{"a.outputs"},
						},
					},
				},
			},
			timeout: 10 * time.Millisecond,
			err:     false,
		},
		{
			name: "timeout",
			loadReq: &pb.LoadPipelineRequest{
				Pipeline: &pb.Pipeline{
					Name:    "foo",
					Version: 1,
					Uid:     "x",
					Steps: []*pb.PipelineStep{
						{
							Name: "a",
						},
						{
							Name:   "b",
							Inputs: []string{"a.outputs"},
						},
					},
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
				err := s.pipelineHandler.AddPipeline(test.loadReq.Pipeline)
				g.Expect(err).To(BeNil())
			}

			stream := newStubPipelineStatusServer(1, 5*time.Millisecond)
			s.pipelineEventStream.streams[stream] = &PipelineSubscription{
				name:   "dummy",
				stream: stream,
				fin:    make(chan bool),
			}
			g.Expect(s.pipelineEventStream.streams[stream]).ToNot(BeNil())
			hub.PublishPipelineEvent(pipelineEventHandlerName, coordinator.PipelineEventMsg{
				PipelineName: "foo", PipelineVersion: 1})

			// to allow events to propagate
			time.Sleep(500 * time.Millisecond)

			if test.err {
				g.Expect(s.pipelineEventStream.streams).To(HaveLen(0))
			} else {

				var psr *pb.PipelineStatusResponse
				select {
				case next := <-stream.msgs:
					psr = next
				default:
					t.Fail()
				}

				g.Expect(psr).ToNot(BeNil())
				g.Expect(psr.Versions).To(HaveLen(1))
				g.Expect(psr.Versions[0].State.Status).To(Equal(pb.PipelineVersionState_PipelineCreate))
				g.Expect(s.pipelineEventStream.streams).To(HaveLen(1))
			}
		})
	}
}
