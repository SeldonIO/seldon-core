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

func TestSendCurrentPipelineStatuses(t *testing.T) {
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

func TestPublishPipelineEventWithTimeout(t *testing.T) {
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
			s, hub := createTestScheduler(t)
			s.timeout = test.timeout
			if test.loadReq != nil {
				err := s.pipelineHandler.AddPipeline(test.loadReq.Pipeline)
				g.Expect(err).To(BeNil())
			}

			stream := newStubPipelineStatusServer(2, 5*time.Millisecond)
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

func receiveMessageFromStream(
	g *WithT, t *testing.T, stream *stubPipelineStatusServer, expectedName string, expectedVersion int,
) *pb.PipelineStatusResponse {
	time.Sleep(500 * time.Millisecond)

	var msr *pb.PipelineStatusResponse
	select {
	case next := <-stream.msgs:
		msr = next
	case <-time.After(2 * time.Second):
		msr = nil
	}
	return msr
}

func getPipelineVersion(g *WithT, pipelineHandler pipeline.PipelineHandler, name string, version uint32) *pipeline.PipelineVersion {
	pip, err := pipelineHandler.GetPipeline(name)
	g.Expect(err).To(BeNil())

	pv := pip.GetPipelineVersion(version)
	g.Expect(pv).ToNot(BeNil())
	return pv
}

func TestAddAndRemovePipelineNoPipelineGw(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name    string
		loadReq *pb.Pipeline
	}

	message := "No pipeline-gw available to handle pipeline"
	tests := []test{
		{
			name: "add and remove pipeline - no pipelinegw (PipelineCreate)",
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
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, _ := createTestScheduler(t)

			// add operator stream
			stream := newStubPipelineStatusServer(100, 5*time.Millisecond)
			subscription := &PipelineSubscription{
				name:   "dummy",
				stream: stream,
				fin:    make(chan bool),
			}
			s.pipelineEventStream.streams[stream] = subscription
			g.Expect(s.pipelineEventStream.streams[stream]).ToNot(BeNil())

			// add a pipeline to the store and check no message is received
			err := s.pipelineHandler.AddPipeline(test.loadReq)
			g.Expect(err).To(BeNil())

			// check operator stream receives add message
			msg := receiveMessageFromStream(g, t, stream, test.loadReq.Name, 1)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(1))

			// check operator stream receives message with status updated and reason
			msg = receiveMessageFromStream(g, t, stream, test.loadReq.Name, 1)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(1))

			// check pipeline gw status and reason
			pv := getPipelineVersion(g, s.pipelineHandler, test.loadReq.Name, 1)
			g.Expect(pv.State.PipelineGwStatus).To(Equal(pipeline.PipelineCreate))
			g.Expect(pv.State.PipelineGwReason).To(Equal(message))

			// remove the pipeline
			err = s.pipelineHandler.RemovePipeline(test.loadReq.Name)
			g.Expect(err).To(BeNil())

			// check operator stream receives remove message
			msg = receiveMessageFromStream(g, t, stream, test.loadReq.Name, 1)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(1))

			// check operator stream receives message with status updated and reason
			msg = receiveMessageFromStream(g, t, stream, test.loadReq.Name, 1)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(1))

			// check pipeline gw status and reason
			pv = getPipelineVersion(g, s.pipelineHandler, test.loadReq.Name, 1)
			g.Expect(pv.State.PipelineGwStatus).To(Equal(pipeline.PipelineTerminated))
			g.Expect(pv.State.PipelineGwReason).To(Equal(message))
		})
	}
}

func TestPipelineGwRebalanceNoPipelineGw(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name        string
		loadReq     *pb.Pipeline
		initStatus  pipeline.PipelineStatus
		finalStatus pipeline.PipelineStatus
	}

	tests := []test{
		{
			name: "rebalance - no pipelinegw, no operator connected (PipelineReady -> PipelineCreate)",
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
			initStatus:  pipeline.PipelineReady,
			finalStatus: pipeline.PipelineCreate,
		},
		{
			name: "rebalance - no pipelinegw, operator connected (PipelineTerminating -> PipelineTerminated)",
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
			initStatus:  pipeline.PipelineTerminating,
			finalStatus: pipeline.PipelineTerminated,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, _ := createTestScheduler(t)

			// add operator stream
			stream := newStubPipelineStatusServer(1, 5*time.Millisecond)
			subscription := &PipelineSubscription{
				name:   "dummy",
				stream: stream,
				fin:    make(chan bool),
			}
			s.pipelineEventStream.streams[stream] = subscription
			g.Expect(s.pipelineEventStream.streams[stream]).ToNot(BeNil())

			// add a pipeline to the store and check no message is received
			err := s.pipelineHandler.AddPipeline(test.loadReq)
			g.Expect(err).To(BeNil())

			// receive message from adding the pipeline
			msg := receiveMessageFromStream(g, t, stream, test.loadReq.Name, 1)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(1))

			// check pipeline gw status and reason
			pv := getPipelineVersion(g, s.pipelineHandler, test.loadReq.Name, 1)
			g.Expect(pv.State.PipelineGwStatus).To(Equal(pipeline.PipelineCreate))
			g.Expect(pv.State.PipelineGwReason).To(Equal("No pipeline-gw available to handle pipeline"))

			// set pipeline to ready
			err = s.pipelineHandler.SetPipelineGwPipelineState(test.loadReq.Name, 1, test.loadReq.Uid, test.initStatus, "", sourcePipelineStatusEvent)
			g.Expect(err).To(BeNil())

			// receive message from setting the pipeline to ready
			msg = receiveMessageFromStream(g, t, stream, test.loadReq.Name, 1)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(1))

			// check pipeline gw status and reason
			pv = getPipelineVersion(g, s.pipelineHandler, test.loadReq.Name, 1)
			g.Expect(pv.State.PipelineGwStatus).To(Equal(test.initStatus))

			// trigger rebalance
			s.pipelineGwRebalance()

			// check no message is received
			msg = receiveMessageFromStream(g, t, stream, test.loadReq.Name, 1)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(1))

			// check pipeline gw status and reason
			pv = getPipelineVersion(g, s.pipelineHandler, test.loadReq.Name, 1)
			g.Expect(pv.State.PipelineGwStatus).To(Equal(test.finalStatus))
		})
	}
}

func TestPipelineGwRebalanceCorrectMessages(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name             string
		loadReq          *pb.Pipeline
		pipelineGwStatus pipeline.PipelineStatus
		versionLen       int
	}

	tests := []test{
		{
			name: "rebalance message - create pipeline",
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
			pipelineGwStatus: pipeline.PipelineReady,
			versionLen:       1,
		},
		{
			name: "rebalance message - delete pipeline",
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
			pipelineGwStatus: pipeline.PipelineTerminating,
			versionLen:       0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create a test scheduler - note it uses a load balancer with 1 partition
			s, _ := createTestScheduler(t)

			// create operator stream
			operatorStream := newStubPipelineStatusServer(1, 5*time.Millisecond)
			operatorSubscription := &PipelineSubscription{
				name:   "dummy-operator",
				stream: operatorStream,
				fin:    make(chan bool),
			}
			s.pipelineEventStream.streams[operatorStream] = operatorSubscription
			g.Expect(s.pipelineEventStream.streams[operatorStream]).ToNot(BeNil())

			// create pipelinegw stream
			pipelineGwStream := newStubPipelineStatusServer(1, 5*time.Millisecond)
			pipelineGwSubscription := &PipelineSubscription{
				name:              "dummy-pipelinegw",
				stream:            pipelineGwStream,
				isPipelineGateway: true,
				fin:               make(chan bool),
			}
			s.pipelineEventStream.streams[pipelineGwStream] = pipelineGwSubscription
			g.Expect(s.pipelineEventStream.streams[pipelineGwStream]).ToNot(BeNil())

			// add stream to the load balancer
			s.pipelineGWLoadBalancer.AddServer(pipelineGwSubscription.name)

			// add a pipeline to the store and check message is received by both operator and pipelinegw
			err := s.pipelineHandler.AddPipeline(test.loadReq)
			g.Expect(err).To(BeNil())

			// receive add message from operator and pipelinegw
			msg := receiveMessageFromStream(g, t, operatorStream, test.loadReq.Name, 1)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(1))

			msg = receiveMessageFromStream(g, t, pipelineGwStream, test.loadReq.Name, 1)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(1))

			// set pipelin gw status
			err = s.pipelineHandler.SetPipelineGwPipelineState(test.loadReq.Name, 1, test.loadReq.Uid, test.pipelineGwStatus, "", sourcePipelineStatusEvent)
			g.Expect(err).To(BeNil())

			// receive status update message from operator only
			msg = receiveMessageFromStream(g, t, operatorStream, test.loadReq.Name, 1)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(1))

			pv := getPipelineVersion(g, s.pipelineHandler, test.loadReq.Name, 1)
			g.Expect(pv.State.PipelineGwStatus).To(Equal(test.pipelineGwStatus))

			// trigger rebalance and check rebalance message is received by pipelinegw only
			s.pipelineGwRebalance()

			// check message is received by operator
			msg = receiveMessageFromStream(g, t, operatorStream, test.loadReq.Name, 1)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(1))

			// operator should only change status if pipeline is not terminating
			if test.versionLen > 0 {
				pv = getPipelineVersion(g, s.pipelineHandler, test.loadReq.Name, 1)
				g.Expect(pv.State.PipelineGwStatus).To(Equal(pipeline.PipelineCreating))
				g.Expect(pv.State.PipelineGwReason).To(Equal("Rebalance"))
			}

			// check rebalance message is received by pipelinegw (create or delete based on the verisonLen)
			msg = receiveMessageFromStream(g, t, pipelineGwStream, test.loadReq.Name, 1)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(test.versionLen))
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
			s, _ := createTestScheduler(t)

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
