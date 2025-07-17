/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package server

import (
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
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
			s.pipelineEventStream.mu.Lock()
			s.pipelineEventStream.streams[stream] = &PipelineSubscription{
				name:   "dummy",
				stream: stream,
				fin:    make(chan bool),
			}
			g.Expect(s.pipelineEventStream.streams[stream]).ToNot(BeNil())
			s.pipelineEventStream.mu.Unlock()

			hub.PublishPipelineEvent(pipelineEventHandlerName, coordinator.PipelineEventMsg{
				PipelineName: "foo", PipelineVersion: 1})

			// to allow events to propagate
			time.Sleep(500 * time.Millisecond)

			if test.err {
				s.pipelineEventStream.mu.Lock()
				g.Expect(s.pipelineEventStream.streams).To(HaveLen(0))
				s.pipelineEventStream.mu.Unlock()
				return
			}

			var psr *pb.PipelineStatusResponse
			select {
			case next := <-stream.msgs:
				psr = next
			case <-time.After(2 * time.Second):
				t.Fail()
			}

			g.Expect(psr).ToNot(BeNil())
			g.Expect(psr.Versions).To(HaveLen(1))
			g.Expect(psr.Versions[0].State.Status).To(Equal(pb.PipelineVersionState_PipelineCreate))

			s.pipelineEventStream.mu.Lock()
			g.Expect(s.pipelineEventStream.streams).To(HaveLen(1))
			s.pipelineEventStream.mu.Unlock()

			s.pipelineEventStream.mu.Lock()
			g.Expect(s.pipelineEventStream.streams).To(HaveLen(1))
			s.pipelineEventStream.mu.Unlock()
		})
	}
}

func TestPipelineGwRebalanceMessage(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name           string
		loadReq        *pb.Pipeline
		pipelineStatus pipeline.PipelineStatus
	}

	tests := []test{
		{
			name: "rebalance pipeline ready",
			loadReq: &pb.Pipeline{
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
			pipelineStatus: pipeline.PipelineReady,
		},
		{
			name: "rebalance pipeline create",
			loadReq: &pb.Pipeline{
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
			pipelineStatus: pipeline.PipelineCreate,
		},
		{
			name: "rebalance pipeline creating",
			loadReq: &pb.Pipeline{
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
			pipelineStatus: pipeline.PipelineCreating,
		},
		{
			name: "rebalance pipeline terminating",
			loadReq: &pb.Pipeline{
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
			pipelineStatus: pipeline.PipelineTerminating,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create a test scheduler - note it uses a load balancer with 1 partition
			s, _ := createTestScheduler()

			// create pipelinegw stream
			stream := newStubPipelineStatusServer(1, 5*time.Millisecond)
			subscription := &PipelineSubscription{
				name:   "dummy",
				stream: stream,
				fin:    make(chan bool),
			}
			s.pipelineEventStream.streams[stream] = subscription
			g.Expect(s.pipelineEventStream.streams[stream]).ToNot(BeNil())

			// add stream to the load balancer
			s.pipelineGWLoadBalancer.AddServer(subscription.name)

			// add a pipeline to the store
			err := s.pipelineHandler.AddPipeline(test.loadReq)
			g.Expect(err).To(BeNil())

			err = s.pipelineHandler.SetPipelineState(
				test.loadReq.Name,
				test.loadReq.Version,
				test.loadReq.Uid,
				test.pipelineStatus,
				"dummy_reason",
				"dummy_source",
			)
			g.Expect(err).To(BeNil())

			// trigger rebalance and allow events to propagate
			s.pipelineGwRebalance()
			time.Sleep(500 * time.Millisecond)

			var msr *pb.PipelineStatusResponse
			select {
			case next := <-stream.msgs:
				msr = next
			case <-time.After(2 * time.Second):
				t.Fail()
			}

			g.Expect(msr).ToNot(BeNil())
			g.Expect(msr.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msr.Versions).To(HaveLen(1))
		})
	}
}

func TestPipelineGwRebalance(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name      string
		pipelines []*pb.Pipeline
		replicas  int // number of pipelinegw instances
	}

	tests := []test{
		{
			name: "rebalance multiple models across N replicas",
			pipelines: []*pb.Pipeline{
				{
					Name:    "pipeline1",
					Version: 1,
					Uid:     "uid1",
					Steps: []*pb.PipelineStep{
						{
							Name: "step1",
						},
						{
							Name:   "step2",
							Inputs: []string{"step1.outputs"},
						},
					},
				},
				{
					Name:    "pipeline2",
					Version: 1,
					Uid:     "uid2",
					Steps: []*pb.PipelineStep{
						{
							Name: "step1",
						},
						{
							Name:   "step2",
							Inputs: []string{"step1.outputs"},
						},
					},
				},
				{
					Name:    "pipeline3",
					Version: 1,
					Uid:     "uid3",
					Steps: []*pb.PipelineStep{
						{
							Name: "step1",
						},
						{
							Name:   "step2",
							Inputs: []string{"step1.outputs"},
						},
					},
				},
			},
			replicas: 4, // test with 4 modelgw replicas
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, _ := createTestScheduler()

			var streams []*stubPipelineStatusServer
			for i := 0; i < test.replicas; i++ {
				// create pipelinegw stream
				stream := newStubPipelineStatusServer(20, 5*time.Millisecond)
				subscription := &PipelineSubscription{
					name:              fmt.Sprintf("dummy%d", i),
					ip:                fmt.Sprintf("127.0.0.%d", i+1),
					isPipelineGateway: true,
					stream:            stream,
					fin:               make(chan bool),
				}
				s.pipelineEventStream.streams[stream] = subscription
				s.pipelineEventStream.namesToIps[subscription.name] = fmt.Sprintf("127.0.0.%d", i+1)
				g.Expect(s.pipelineEventStream.streams[stream]).ToNot(BeNil())

				// add stream to the load balancer
				s.pipelineGWLoadBalancer.AddServer(subscription.name)
				streams = append(streams, stream)
			}
			g.Expect(len(s.pipelineEventStream.streams)).To(Equal(test.replicas))

			// Load all pipelines into the store
			for _, req := range test.pipelines {
				err := s.pipelineHandler.AddPipeline(req)
				g.Expect(err).To(BeNil())
			}

			// Allow events to propagate
			time.Sleep(500 * time.Millisecond)

			// Read all messages from the streams before rebalance
			for i, stream := range streams {
			drainLoop:
				for {
					select {
					case msg := <-stream.msgs:
						log.Infof("Drained message from stream %d: %v", i, msg)
					case <-time.After(100 * time.Millisecond):
						break drainLoop
					}
				}
			}

			log.Info("Drained all messages from streams before rebalance")

			// Trigger rebalance and allow events to propagate
			s.pipelineGwRebalance()
			time.Sleep(500 * time.Millisecond)

			// Collect pipeline assignments from all streams
			pipelineAssignments := make(map[string]int)
			for _, stream := range streams {
			collectLoop:
				for {
					select {
					case msg := <-stream.msgs:
						name := msg.PipelineName
						pipelineAssignments[name] += len(msg.Versions)
					case <-time.After(100 * time.Millisecond):
						break collectLoop
					}
				}
			}

			// Expect each pipeline to have exactly 1 replica assigned
			g.Expect(pipelineAssignments).To(HaveLen(len(test.pipelines)))
			for _, req := range test.pipelines {
				g.Expect(pipelineAssignments[req.Name]).To(Equal(1),
					fmt.Sprintf("pipeline %s should have exactly 1 replica assigned", req.Name))
			}
		})
	}
}
