/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package dataflow

import (
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/config"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/chainer"
)

func TestCreateTopicSources(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		server       *ChainerServer
		pipelineName string
		inputs       []string
		sources      []*chainer.PipelineTopic
	}

	getPtrStr := func(val string) *string { return &val }
	createTopicNamer := func(namespace string, topicPrefix string) *kafka.TopicNamer {
		tn, err := kafka.NewTopicNamer(namespace, topicPrefix)
		g.Expect(err).To(BeNil())
		return tn
	}
	tests := []test{
		{
			name: "misc inputs",
			server: &ChainerServer{
				logger:     log.New(),
				topicNamer: createTopicNamer("default", "seldon"),
			},
			pipelineName: "p1",
			inputs: []string{
				"a",
				"b.inputs",
				"c.inputs.t1",
			},
			sources: []*chainer.PipelineTopic{
				{PipelineName: "p1", TopicName: "seldon.default.model.a", Tensor: nil},
				{PipelineName: "p1", TopicName: "seldon.default.model.b.inputs", Tensor: nil},
				{PipelineName: "p1", TopicName: "seldon.default.model.c.inputs", Tensor: getPtrStr("t1")},
			},
		},
		{
			name: "default inputs",
			server: &ChainerServer{
				logger:     log.New(),
				topicNamer: createTopicNamer("ns1", "seldon"),
			},
			pipelineName: "p1",
			inputs:       []string{},
			sources: []*chainer.PipelineTopic{
				{PipelineName: "p1", TopicName: "seldon.ns1.pipeline.p1.inputs", Tensor: nil},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sources := test.server.createTopicSources(test.inputs, test.pipelineName)
			g.Expect(sources).To(Equal(test.sources))
		})
	}
}

func TestCreatePipelineTopicSources(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name    string
		server  *ChainerServer
		inputs  []string
		sources []*chainer.PipelineTopic
	}

	getPtrStr := func(val string) *string { return &val }
	createTopicNamer := func(namespace string, topicPrefix string) *kafka.TopicNamer {
		tn, err := kafka.NewTopicNamer(namespace, topicPrefix)
		g.Expect(err).To(BeNil())
		return tn
	}
	tests := []test{
		{
			name: "misc inputs",
			server: &ChainerServer{
				logger:     log.New(),
				topicNamer: createTopicNamer("default", "seldon"),
			},
			inputs: []string{
				"foo.inputs",
				"foo.outputs",
				"foo.step.bar.inputs",
				"foo.step.bar.outputs",
				"foo.step.bar.inputs.tensora",
			},
			sources: []*chainer.PipelineTopic{
				{PipelineName: "foo", TopicName: "seldon.default.pipeline.foo.inputs", Tensor: nil},
				{PipelineName: "foo", TopicName: "seldon.default.pipeline.foo.outputs", Tensor: nil},
				{PipelineName: "foo", TopicName: "seldon.default.model.bar.inputs", Tensor: nil},
				{PipelineName: "foo", TopicName: "seldon.default.model.bar.outputs", Tensor: nil},
				{PipelineName: "foo", TopicName: "seldon.default.model.bar.inputs", Tensor: getPtrStr("tensora")},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sources := test.server.createPipelineTopicSources(test.inputs)
			g.Expect(sources).To(Equal(test.sources))
		})
	}
}

func TestCreateTriggerSources(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		server       *ChainerServer
		pipelineName string
		inputs       []string
		sources      []*chainer.PipelineTopic
	}

	createTopicNamer := func(namespace string, topicPrefix string) *kafka.TopicNamer {
		tn, err := kafka.NewTopicNamer(namespace, topicPrefix)
		g.Expect(err).To(BeNil())
		return tn
	}
	getPtrStr := func(val string) *string { return &val }
	tests := []test{
		{
			name: "misc inputs",
			server: &ChainerServer{
				logger:     log.New(),
				topicNamer: createTopicNamer("default", "seldon"),
			},
			pipelineName: "p1",
			inputs: []string{
				"a",
				"b.inputs",
				"c.inputs.t1",
			},
			sources: []*chainer.PipelineTopic{
				{PipelineName: "p1", TopicName: "seldon.default.model.a", Tensor: nil},
				{PipelineName: "p1", TopicName: "seldon.default.model.b.inputs", Tensor: nil},
				{PipelineName: "p1", TopicName: "seldon.default.model.c.inputs", Tensor: getPtrStr("t1")},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sources := test.server.createTriggerSources(test.inputs, test.pipelineName)
			g.Expect(sources).To(Equal(test.sources))
		})
	}
}

// test to make sure we remove old versions of the pipeline when a new version is added
func TestPipelineRollingUpgradeEvents(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name       string
		loadReqV1  *scheduler.Pipeline
		loadReqV2  *scheduler.Pipeline
		err        bool // when true old version was not marked as ready
		connection bool
	}

	tests := []test{
		{
			name: "old version removed - was ready",
			loadReqV1: &scheduler.Pipeline{

				Name:    "foo",
				Version: 1,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			loadReqV2: &scheduler.Pipeline{

				Name:    "foo",
				Version: 2,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			err:        false,
			connection: true,
		},
		{
			name: "old version removed - was not ready",
			loadReqV1: &scheduler.Pipeline{

				Name:    "foo",
				Version: 1,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			loadReqV2: &scheduler.Pipeline{

				Name:    "foo",
				Version: 2,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			err:        true,
			connection: true,
		},
		{
			name: "no new version",
			loadReqV1: &scheduler.Pipeline{

				Name:    "foo",
				Version: 1,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			err:        false,
			connection: true,
		},
		{
			name: "no connection",
			loadReqV1: &scheduler.Pipeline{

				Name:    "foo",
				Version: 1,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			loadReqV2: &scheduler.Pipeline{

				Name:    "foo",
				Version: 2,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			err:        false,
			connection: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			serverName := "dummy"
			s, _ := createTestScheduler(t, serverName)

			err := s.pipelineHandler.AddPipeline(test.loadReqV1) // version 1
			g.Expect(err).To(BeNil())
			if !test.err {
				err = s.pipelineHandler.SetPipelineState(test.loadReqV1.Name, test.loadReqV1.Version, test.loadReqV1.Uid, pipeline.PipelineReady, "", sourceChainerServer)
			}
			g.Expect(err).To(BeNil())

			if test.loadReqV2 != nil {
				err = s.pipelineHandler.AddPipeline(test.loadReqV2) // version 2
				g.Expect(err).To(BeNil())
				err = s.pipelineHandler.SetPipelineState(test.loadReqV2.Name, test.loadReqV2.Version, test.loadReqV2.Uid, pipeline.PipelineReady, "", sourceChainerServer)
				g.Expect(err).To(BeNil())
			}

			stream := newStubServerStatusServer(1)
			if test.connection {
				s.streams[serverName] = &ChainerSubscription{
					name:   "dummy",
					stream: stream,
					fin:    make(chan bool),
				}
				g.Expect(s.streams[serverName]).ToNot(BeNil())
			}

			// to allow events to propagate
			time.Sleep(500 * time.Millisecond)

			if test.connection {
				if test.loadReqV2 != nil {
					var psr *chainer.PipelineUpdateMessage
					select {
					case next := <-stream.msgs:
						psr = next
					default:
						t.Fail()
					}

					g.Expect(psr).ToNot(BeNil())
					g.Expect(psr.Pipeline).To(Equal(test.loadReqV1.Name))
					g.Expect(psr.Version).To(Equal(uint32(test.loadReqV1.Version)))
					g.Expect(psr.Op).To(Equal(chainer.PipelineUpdateMessage_Delete))
				} else {
					var psr *chainer.PipelineUpdateMessage
					select {
					case next := <-stream.msgs:
						psr = next
					default:
						psr = nil
					}

					g.Expect(psr).To(BeNil())
				}
			} else {
				pipeline, err := s.pipelineHandler.GetPipeline(test.loadReqV2.Name)
				g.Expect(err).To(BeNil())
				// error message should be set
				g.Expect(pipeline.GetLatestPipelineVersion().State.Reason).ToNot(Equal(""))
			}

		})
	}
}

func TestPipelineEvents(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name       string
		loadReq    *scheduler.Pipeline
		status     pipeline.PipelineStatus
		connection bool
	}

	tests := []test{
		{
			name: "add new pipeline version",
			loadReq: &scheduler.Pipeline{

				Name:    "foo",
				Version: 1,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			status:     pipeline.PipelineCreate,
			connection: true,
		},
		{
			name: "remove pipeline version",
			loadReq: &scheduler.Pipeline{

				Name:    "foo",
				Version: 1,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			status:     pipeline.PipelineTerminate,
			connection: true,
		},
		{
			name: "no connection",
			loadReq: &scheduler.Pipeline{

				Name:    "foo",
				Version: 1,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			status:     pipeline.PipelineTerminate,
			connection: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			serverName := "dummy"
			s, _ := createTestScheduler(t, serverName)

			err := s.pipelineHandler.AddPipeline(test.loadReq) // version 1
			g.Expect(err).To(BeNil())

			err = s.pipelineHandler.SetPipelineState(test.loadReq.Name, test.loadReq.Version, test.loadReq.Uid, test.status, "", sourceChainerServer)
			g.Expect(err).To(BeNil())

			stream := newStubServerStatusServer(1)
			if test.connection {
				s.streams[serverName] = &ChainerSubscription{
					name:   "dummy",
					stream: stream,
					fin:    make(chan bool),
				}
				g.Expect(s.streams[serverName]).ToNot(BeNil())
			}

			// to allow events to propagate
			time.Sleep(500 * time.Millisecond)

			if test.connection {
				var psr *chainer.PipelineUpdateMessage
				select {
				case next := <-stream.msgs:
					psr = next
				default:
					t.Fail()
				}

				g.Expect(psr).ToNot(BeNil())
				g.Expect(psr.Pipeline).To(Equal(test.loadReq.Name))
				g.Expect(psr.Version).To(Equal(uint32(test.loadReq.Version)))
				if test.status == pipeline.PipelineCreate {
					g.Expect(psr.Op).To(Equal(chainer.PipelineUpdateMessage_Create))
				} else {
					g.Expect(psr.Op).To(Equal(chainer.PipelineUpdateMessage_Delete))
				}
			} else {
				pipeline, err := s.pipelineHandler.GetPipeline(test.loadReq.Name)
				g.Expect(err).To(BeNil())
				// error message should be set
				g.Expect(pipeline.GetLatestPipelineVersion().State.Reason).ToNot(Equal(""))
			}

		})
	}
}

type stubChainerServer struct {
	msgs chan *chainer.PipelineUpdateMessage
	grpc.ServerStream
}

var _ chainer.Chainer_SubscribePipelineUpdatesServer = (*stubChainerServer)(nil)

func newStubServerStatusServer(capacity int) *stubChainerServer {
	return &stubChainerServer{
		msgs: make(chan *chainer.PipelineUpdateMessage, capacity),
	}
}

func (s *stubChainerServer) Send(r *chainer.PipelineUpdateMessage) error {
	s.msgs <- r
	return nil
}

// TODO: this function is defined elsewhere, refactor to avoid duplication
func createTestScheduler(t *testing.T, serverName string) (*ChainerServer, *coordinator.EventHub) {
	logger := log.New()
	logger.SetLevel(log.WarnLevel)

	eventHub, _ := coordinator.NewEventHub(logger)

	schedulerStore := store.NewMemoryStore(logger, store.NewLocalSchedulerStore(), eventHub)
	pipelineServer := pipeline.NewPipelineStore(logger, eventHub, schedulerStore)

	data :=
		`
	{
	  "bootstrap.servers":"kafka:9092",
	  "consumer":{"session.timeout.ms": 6000, "someBool": true, "someString":"foo"},
	  "producer": {"linger.ms":0},
	  "streams": {"replication.factor": 1}
	}
	`
	configFilePath := fmt.Sprintf("%s/kafka.json", t.TempDir())
	_ = os.WriteFile(configFilePath, []byte(data), 0644)
	kc, _ := config.NewKafkaConfig(configFilePath)

	b := util.NewRingLoadBalancer(1)
	b.AddServer(serverName)
	s, _ := NewChainerServer(logger, eventHub, pipelineServer, "test-ns", b, kc)

	return s, eventHub
}
