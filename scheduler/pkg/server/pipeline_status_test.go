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
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

func receiveMessageFromPipelineStream(
	t *testing.T, stream *stubPipelineStatusServer,
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

func TestSendCurrentPipelineStatuses(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name    string
		loadReq *pb.LoadPipelineRequest
		server  *SchedulerServer
		err     bool
		ctx     context.Context
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
			ctx: context.Background(),
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
			ctx: context.Background(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.loadReq != nil {
				err := test.server.pipelineHandler.AddPipeline(test.loadReq.Pipeline)
				g.Expect(err).To(BeNil())
			}

			stream := newStubPipelineStatusServer(1, 5*time.Millisecond, test.ctx)
			err := test.server.sendCurrentPipelineStatuses(stream, false)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())

				psr := receiveMessageFromPipelineStream(t, stream)
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
		ctx     context.Context
	}

	tests := []test{
		{
			name: "success - pipeline ok",
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
			ctx:     context.Background(),
		},
		{
			name: "failure - timeout",
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
			ctx:     context.Background(),
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

			stream := newStubPipelineStatusServer(2, 5*time.Millisecond, test.ctx)
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

			psr := receiveMessageFromPipelineStream(t, stream)
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

func TestAddAndRemovePipelineNoPipelineGw(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name    string
		loadReq *pb.Pipeline
		ctx     context.Context
	}

	pipelineRemovedMessage := "pipeline removed"
	noPipelineGwMessage := "No pipeline gateway available to handle pipeline"

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
			ctx: context.Background(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, _ := createTestScheduler(t)

			// add operator stream
			stream := newStubPipelineStatusServer(100, 5*time.Millisecond, test.ctx)
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
			msg := receiveMessageFromPipelineStream(t, stream)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(1))
			g.Expect(msg.Versions[0].State.PipelineGwStatus).To(Equal(pb.PipelineVersionState_PipelineCreate))
			g.Expect(msg.Versions[0].State.PipelineGwReason).To(Equal(""))

			// check operator stream receives message with status updated and reason
			msg = receiveMessageFromPipelineStream(t, stream)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(1))
			g.Expect(msg.Versions[0].State.PipelineGwStatus).To(Equal(pb.PipelineVersionState_PipelineCreate))
			g.Expect(msg.Versions[0].State.PipelineGwReason).To(Equal(noPipelineGwMessage))

			// check pipeline gw status and reason have been updated
			pv := getPipelineVersion(g, s.pipelineHandler, test.loadReq.Name, 1)
			g.Expect(pv.State.PipelineGwStatus).To(Equal(pipeline.PipelineCreate))
			g.Expect(pv.State.PipelineGwReason).To(Equal(noPipelineGwMessage))

			// remove the pipeline
			err = s.pipelineHandler.RemovePipeline(test.loadReq.Name)
			g.Expect(err).To(BeNil())

			// check operator stream receives remove message
			msg = receiveMessageFromPipelineStream(t, stream)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(1))
			g.Expect(msg.Versions[0].State.PipelineGwStatus).To(Equal(pb.PipelineVersionState_PipelineTerminate))
			g.Expect(msg.Versions[0].State.PipelineGwReason).To(Equal(pipelineRemovedMessage))

			// check operator stream receives message with status updated and reason
			msg = receiveMessageFromPipelineStream(t, stream)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(1))
			g.Expect(msg.Versions[0].State.PipelineGwStatus).To(Equal(pb.PipelineVersionState_PipelineTerminated))
			g.Expect(msg.Versions[0].State.PipelineGwReason).To(Equal(noPipelineGwMessage))

			// check pipeline gw status and reason
			pv = getPipelineVersion(g, s.pipelineHandler, test.loadReq.Name, 1)
			g.Expect(pv.State.PipelineGwStatus).To(Equal(pipeline.PipelineTerminated))
			g.Expect(pv.State.PipelineGwReason).To(Equal(noPipelineGwMessage))
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
		ctx         context.Context
	}

	noPipelineGwMessage := "No pipeline gateway available to handle pipeline"
	tests := []test{
		{
			name: "rebalance - no pipelinegw, operator connected (PipelineReady -> PipelineCreate)",
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
			ctx:         context.Background(),
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
			ctx:         context.Background(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, _ := createTestScheduler(t)

			// add operator stream
			stream := newStubPipelineStatusServer(1, 5*time.Millisecond, test.ctx)
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
			msg := receiveMessageFromPipelineStream(t, stream)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(1))
			g.Expect(msg.Versions[0].State.PipelineGwStatus).To(Equal(pb.PipelineVersionState_PipelineCreate))
			g.Expect(msg.Versions[0].State.PipelineGwReason).To(Equal(""))

			// receive message with status updated and reason
			msg = receiveMessageFromPipelineStream(t, stream)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(1))
			g.Expect(msg.Versions[0].State.PipelineGwStatus).To(Equal(pb.PipelineVersionState_PipelineCreate))
			g.Expect(msg.Versions[0].State.PipelineGwReason).To(Equal(noPipelineGwMessage))

			// check pipeline gw status and reason
			pv := getPipelineVersion(g, s.pipelineHandler, test.loadReq.Name, 1)
			g.Expect(pv.State.PipelineGwStatus).To(Equal(pipeline.PipelineCreate))
			g.Expect(pv.State.PipelineGwReason).To(Equal(noPipelineGwMessage))

			// set pipeline to ready
			err = s.pipelineHandler.SetPipelineGwPipelineState(
				test.loadReq.Name, 1, test.loadReq.Uid, test.initStatus, "", util.SourcePipelineStatusEvent,
			)
			g.Expect(err).To(BeNil())

			// receive message from setting the pipeline to ready
			msg = receiveMessageFromPipelineStream(t, stream)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(1))
			g.Expect(int32(msg.Versions[0].State.PipelineGwStatus)).To(Equal(pb.PipelineVersionState_PipelineStatus_value[test.initStatus.String()]))
			g.Expect(msg.Versions[0].State.PipelineGwReason).To(Equal(""))

			// check pipeline gw status and reason
			pv = getPipelineVersion(g, s.pipelineHandler, test.loadReq.Name, 1)
			g.Expect(pv.State.PipelineGwStatus).To(Equal(test.initStatus))

			// trigger rebalance
			s.pipelineGwRebalance()

			// receive message with status updated and reason
			msg = receiveMessageFromPipelineStream(t, stream)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(1))
			g.Expect(int32(msg.Versions[0].State.PipelineGwStatus)).To(Equal(pb.PipelineVersionState_PipelineStatus_value[test.finalStatus.String()]))
			g.Expect(msg.Versions[0].State.PipelineGwReason).To(Equal(noPipelineGwMessage))

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
		operation        pb.PipelineStatusResponse_PipelineOperation
		ctx              context.Context
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
			operation:        pb.PipelineStatusResponse_PipelineCreate,
			ctx:              context.Background(),
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
			versionLen:       1,
			operation:        pb.PipelineStatusResponse_PipelineDelete,
			ctx:              context.Background(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create a test scheduler - note it uses a load balancer with 1 partition
			s, _ := createTestScheduler(t)

			// create operator stream
			operatorStream := newStubPipelineStatusServer(1, 5*time.Millisecond, test.ctx)
			operatorSubscription := &PipelineSubscription{
				name:   "dummy-operator",
				stream: operatorStream,
				fin:    make(chan bool),
			}
			s.pipelineEventStream.streams[operatorStream] = operatorSubscription
			g.Expect(s.pipelineEventStream.streams[operatorStream]).ToNot(BeNil())

			// create pipelinegw stream
			pipelineGwStream := newStubPipelineStatusServer(10, 5*time.Millisecond, test.ctx)
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

			// receive add message from operator
			msg := receiveMessageFromPipelineStream(t, operatorStream)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(1))
			g.Expect(msg.Versions[0].State.PipelineGwStatus).To(Equal(pb.PipelineVersionState_PipelineCreate))
			g.Expect(msg.Versions[0].State.PipelineGwReason).To(Equal(""))

			// receive add message from pipelinegw
			msg = receiveMessageFromPipelineStream(t, pipelineGwStream)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(1))
			g.Expect(msg.Versions[0].State.PipelineGwStatus).To(Equal(pb.PipelineVersionState_PipelineCreate))
			g.Expect(msg.Versions[0].State.PipelineGwReason).To(Equal(""))
			g.Expect(msg.Operation).To(Equal(pb.PipelineStatusResponse_PipelineCreate))
			g.Expect(msg.Timestamp).To(Equal(uint64(1)))

			// receive transition to creating message from operator
			msg = receiveMessageFromPipelineStream(t, operatorStream)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(1))
			g.Expect(msg.Versions[0].State.PipelineGwStatus).To(Equal(pb.PipelineVersionState_PipelineCreating))
			g.Expect(msg.Versions[0].State.PipelineGwReason).To(Equal(""))

			// set pipelin gw status
			err = s.pipelineHandler.SetPipelineGwPipelineState(
				test.loadReq.Name, 1, test.loadReq.Uid, test.pipelineGwStatus, "", util.SourcePipelineStatusEvent,
			)
			g.Expect(err).To(BeNil())

			// receive status update message from operator only
			msg = receiveMessageFromPipelineStream(t, operatorStream)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(1))
			g.Expect(int32(msg.Versions[0].State.PipelineGwStatus)).To(Equal(pb.PipelineVersionState_PipelineStatus_value[test.pipelineGwStatus.String()]))
			g.Expect(msg.Versions[0].State.PipelineGwReason).To(Equal(""))

			pv := getPipelineVersion(g, s.pipelineHandler, test.loadReq.Name, 1)
			g.Expect(pv.State.PipelineGwStatus).To(Equal(test.pipelineGwStatus))

			// trigger rebalance and check rebalance message is received by pipelinegw only
			s.pipelineGwRebalance()

			// check message is received by operator
			if test.pipelineGwStatus != pipeline.PipelineTerminating {
				msg = receiveMessageFromPipelineStream(t, operatorStream)
				g.Expect(msg).ToNot(BeNil())
				g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
				g.Expect(msg.Versions).To(HaveLen(1))
				g.Expect(msg.Versions[0].State.PipelineGwStatus).To(Equal(pb.PipelineVersionState_PipelineCreating))
				g.Expect(msg.Versions[0].State.PipelineGwReason).To(Equal("Rebalance"))
			}

			// check rebalance message is received by pipelinegw (create or delete based on the verisonLen)
			msg = receiveMessageFromPipelineStream(t, pipelineGwStream)
			g.Expect(msg).ToNot(BeNil())
			g.Expect(msg.PipelineName).To(Equal(test.loadReq.Name))
			g.Expect(msg.Versions).To(HaveLen(test.versionLen))
			g.Expect(msg.Operation).To(Equal(test.operation))
			g.Expect(msg.Timestamp).To(Equal(uint64(2)))
		})
	}
}

func TestPipelineGwRebalance(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name      string
		pipelines []*pb.Pipeline
		replicas  int // number of pipelinegw instances
		ctx       context.Context
	}

	tests := []test{
		{
			name: "rebalance 3 pipelines across 4 replicas",
			ctx:  context.Background(),
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
			replicas: 4,
		},
		{
			name: "rebalance 3 pipelines across 7 replicas",
			ctx:  context.Background(),
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
			replicas: 7,
		},
		{
			name: "rebalance 2 pipelines across 9 replicas",
			ctx:  context.Background(),
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
			},
			replicas: 9,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, _ := createTestScheduler(t)

			var streams []*stubPipelineStatusServer
			for i := 0; i < test.replicas; i++ {
				// create pipelinegw stream
				stream := newStubPipelineStatusServer(20, 5*time.Millisecond, test.ctx)
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
			pipelineCreateAssignments := make(map[string]int)
			pipelineDeleteAssignments := make(map[string]int)

			for _, stream := range streams {
			collectLoop:
				for {
					select {
					case msg := <-stream.msgs:
						name := msg.PipelineName
						if msg.Operation == pb.PipelineStatusResponse_PipelineCreate {
							pipelineCreateAssignments[name] += 1
						} else if msg.Operation == pb.PipelineStatusResponse_PipelineDelete {
							pipelineDeleteAssignments[name] += 1
						}
					case <-time.After(100 * time.Millisecond):
						break collectLoop
					}
				}
			}

			// Expect each pipeline to have exactly 1 replica assigned
			g.Expect(pipelineCreateAssignments).To(HaveLen(len(test.pipelines)))
			g.Expect(pipelineDeleteAssignments).To(HaveLen(len(test.pipelines)))

			for _, req := range test.pipelines {
				g.Expect(pipelineCreateAssignments[req.Name]).To(Equal(1),
					fmt.Sprintf("pipeline %s should have exactly 1 replica assigned", req.Name),
				)
				g.Expect(pipelineDeleteAssignments[req.Name]).To(Equal(test.replicas-1),
					fmt.Sprintf("pipeline %s should have %d delete assignments", req.Name, test.replicas-1),
				)
			}
		})
	}
}
