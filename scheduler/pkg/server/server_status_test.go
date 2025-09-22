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

	pba "github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	scheduler2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/scheduler"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/experiment"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/synchroniser"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

func TestModelsStatusStream(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name    string
		loadReq *pb.LoadModelRequest
		server  *SchedulerServer
		err     bool
	}

	tests := []test{
		{
			name: "model ok",
			loadReq: &pb.LoadModelRequest{
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo"},
				},
			},
			server: &SchedulerServer{
				modelStore: store.NewMemoryStore(log.New(), store.NewLocalSchedulerStore(), nil),
				logger:     log.New(),
				timeout:    10 * time.Millisecond,
			},
		},
		{
			name: "timeout",
			loadReq: &pb.LoadModelRequest{
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo"},
				},
			},
			server: &SchedulerServer{
				modelStore: store.NewMemoryStore(log.New(), store.NewLocalSchedulerStore(), nil),
				logger:     log.New(),
				timeout:    1 * time.Millisecond,
			},
			err: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.loadReq != nil {
				err := test.server.modelStore.UpdateModel(test.loadReq)
				g.Expect(err).To(BeNil())
			}

			stream := newStubModelStatusServer(1, 5*time.Millisecond)
			err := test.server.sendCurrentModelStatuses(stream)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())

				var msr *pb.ModelStatusResponse
				select {
				case next := <-stream.msgs:
					msr = next
				default:
					t.Fail()
				}

				g.Expect(msr).ToNot(BeNil())
				g.Expect(msr.Versions).To(HaveLen(1))
				g.Expect(msr.Versions[0].State.State).To(Equal(pb.ModelStatus_ModelStateUnknown))
			}
		})
	}
}

func TestModelsStatusEvents(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name    string
		loadReq *pb.LoadModelRequest
		timeout time.Duration
		err     bool
	}

	tests := []test{
		{
			name: "model ok",
			loadReq: &pb.LoadModelRequest{
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo"},
				},
			},
			timeout: 10 * time.Millisecond,
			err:     false,
		},
		{
			name: "timeout",
			loadReq: &pb.LoadModelRequest{
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo"},
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
				err := s.modelStore.UpdateModel(test.loadReq)
				g.Expect(err).To(BeNil())
			}

			stream := newStubModelStatusServer(2, 5*time.Millisecond)
			s.modelEventStream.streams[stream] = &ModelSubscription{
				name:   "dummy",
				stream: stream,
				fin:    make(chan bool),
			}
			g.Expect(s.modelEventStream.streams[stream]).ToNot(BeNil())
			hub.PublishModelEvent(modelEventHandlerName, coordinator.ModelEventMsg{
				ModelName: "foo", ModelVersion: 1,
			})

			// to allow events to propagate
			time.Sleep(500 * time.Millisecond)

			if test.err {
				s.modelEventStream.mu.Lock()
				g.Expect(s.modelEventStream.streams).To(HaveLen(0))
				s.modelEventStream.mu.Unlock()
			} else {

				var msr *pb.ModelStatusResponse
				select {
				case next := <-stream.msgs:
					msr = next
				default:
					t.Fail()
				}

				g.Expect(msr).ToNot(BeNil())
				g.Expect(msr.Versions).To(HaveLen(1))
				g.Expect(msr.Versions[0].State.State).To(Equal(pb.ModelStatus_ModelStateUnknown))
				g.Expect(s.modelEventStream.streams).To(HaveLen(1))
			}
		})
	}
}

func TestServersStatusStream(t *testing.T) {
	type serverReplicaRequest struct {
		request  *pba.AgentSubscribeRequest
		draining bool
	}

	g := NewGomegaWithT(t)
	type test struct {
		name    string
		loadReq []serverReplicaRequest
		server  *SchedulerServer
		err     bool
	}

	tests := []test{
		{
			name: "server ok - 1 empty replica",
			loadReq: []serverReplicaRequest{
				{
					request: &pba.AgentSubscribeRequest{
						ServerName: "foo",
					},
				},
			},
			server: &SchedulerServer{
				modelStore: store.NewMemoryStore(log.New(), store.NewLocalSchedulerStore(), nil),
				logger:     log.New(),
				timeout:    10 * time.Millisecond,
			},
		},
		{
			name: "server ok - multiple replicas",
			loadReq: []serverReplicaRequest{
				{
					request: &pba.AgentSubscribeRequest{
						ServerName: "foo",
						ReplicaIdx: 0,
						LoadedModels: []*pba.ModelVersion{
							{
								Model: &pb.Model{
									Meta: &pb.MetaData{Name: "foo-model"},
								},
							},
						},
					},
				},
				{
					request: &pba.AgentSubscribeRequest{
						ServerName: "foo",
						ReplicaIdx: 1,
						LoadedModels: []*pba.ModelVersion{
							{
								Model: &pb.Model{
									Meta: &pb.MetaData{Name: "foo-model"},
								},
							},
						},
					},
				},
			},
			server: &SchedulerServer{
				modelStore: store.NewMemoryStore(log.New(), store.NewLocalSchedulerStore(), nil),
				logger:     log.New(),
				timeout:    10 * time.Millisecond,
			},
		},
		{
			name: "server ok - multiple replicas with draining",
			loadReq: []serverReplicaRequest{
				{
					request: &pba.AgentSubscribeRequest{
						ServerName: "foo",
						ReplicaIdx: 0,
						LoadedModels: []*pba.ModelVersion{
							{
								Model: &pb.Model{
									Meta: &pb.MetaData{Name: "foo-model"},
								},
							},
						},
					},
				},
				{
					request: &pba.AgentSubscribeRequest{
						ServerName: "foo",
						ReplicaIdx: 1,
						LoadedModels: []*pba.ModelVersion{
							{
								Model: &pb.Model{
									Meta: &pb.MetaData{Name: "foo-model"},
								},
							},
						},
					},
					draining: true,
				},
			},
			server: &SchedulerServer{
				modelStore: store.NewMemoryStore(log.New(), store.NewLocalSchedulerStore(), nil),
				logger:     log.New(),
				timeout:    10 * time.Millisecond,
			},
		},
		{
			name: "timeout",
			loadReq: []serverReplicaRequest{
				{
					request: &pba.AgentSubscribeRequest{
						ServerName: "foo",
					},
				},
			},
			server: &SchedulerServer{
				modelStore: store.NewMemoryStore(log.New(), store.NewLocalSchedulerStore(), nil),
				logger:     log.New(),
				timeout:    1 * time.Millisecond,
			},
			err: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expectedReplicas := int32(0)
			expectedNumLoadedModelReplicas := int32(0)
			if test.loadReq != nil {
				for _, r := range test.loadReq {
					err := test.server.modelStore.AddServerReplica(r.request)
					g.Expect(err).To(BeNil())
					if !r.draining {
						expectedReplicas++
						expectedNumLoadedModelReplicas += int32(len(r.request.LoadedModels))
					} else {
						server, _ := test.server.modelStore.GetServer("foo", true, false)
						server.Replicas[int(r.request.ReplicaIdx)].SetIsDraining()
					}
				}
			}

			stream := newStubServerStatusServer(1, 5*time.Millisecond)
			err := test.server.sendCurrentServerStatuses(stream)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())

				var ssr *pb.ServerStatusResponse
				select {
				case next := <-stream.msgs:
					ssr = next
				default:
					t.Fail()
				}

				g.Expect(ssr).ToNot(BeNil())
				g.Expect(ssr.ServerName).To(Equal("foo"))
				g.Expect(ssr.GetAvailableReplicas()).To(Equal(expectedReplicas))
				g.Expect(ssr.NumLoadedModelReplicas).To(Equal(expectedNumLoadedModelReplicas))
				g.Expect(ssr.Type).To(Equal(pb.ServerStatusResponse_StatusUpdate))
			}
		})
	}
}

func TestModelEventsForServerStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name    string
		loadReq *pba.AgentSubscribeRequest
		timeout time.Duration
		err     bool
	}

	tests := []test{
		{
			name: "server ok",
			loadReq: &pba.AgentSubscribeRequest{
				ServerName: "foo",
			},
			timeout: 10 * time.Millisecond,
			err:     false,
		},
		{
			name: "timeout",
			loadReq: &pba.AgentSubscribeRequest{
				ServerName: "foo",
			},
			timeout: 1 * time.Millisecond,
			err:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, _ := createTestScheduler(t)
			s.timeout = test.timeout
			if test.loadReq != nil {
				err := s.modelStore.AddServerReplica(test.loadReq)
				g.Expect(err).To(BeNil())
				err = s.modelStore.UpdateModel(&pb.LoadModelRequest{
					Model: &pb.Model{
						Meta: &pb.MetaData{Name: "foo"},
					},
				})
				g.Expect(err).To(BeNil())
				err = s.modelStore.UpdateLoadedModels(
					"foo", 1, "foo", []*store.ServerReplica{
						store.NewServerReplica("", 8080, 5001, 0, store.NewServer("foo", true), []string{}, 100, 100, 0, map[store.ModelVersionID]bool{}, 100),
					},
				)
				g.Expect(err).To(BeNil())
			}

			stream := newStubServerStatusServer(1, 5*time.Millisecond)
			s.serverEventStream.mu.Lock()
			s.serverEventStream.streams[stream] = &ServerSubscription{
				name:   "dummy",
				stream: stream,
				fin:    make(chan bool),
			}
			g.Expect(s.serverEventStream.streams[stream]).ToNot(BeNil())
			s.serverEventStream.mu.Unlock()

			// to allow events to propagate
			time.Sleep(500 * time.Millisecond)

			if test.err {
				s.serverEventStream.mu.Lock()
				g.Expect(s.serverEventStream.streams).To(HaveLen(0))
				s.serverEventStream.mu.Unlock()
			} else {
				var ssr *pb.ServerStatusResponse
				select {
				case next := <-stream.msgs:
					ssr = next
				default:
					t.Fail()
				}

				g.Expect(ssr).ToNot(BeNil())
				g.Expect(ssr.ServerName).To(Equal("foo"))
				g.Expect(s.serverEventStream.streams).To(HaveLen(1))
				g.Expect(ssr.Type).To(Equal(pb.ServerStatusResponse_StatusUpdate))
			}
		})
	}
}

func TestServerScaleUpEvents(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name             string
		loadReq          *pba.AgentSubscribeRequest
		timeout          time.Duration
		modelReplicas    int
		expectedReplicas int32
		scaleUp          bool
		notifyReq        *pb.ServerNotifyRequest
	}

	tests := []test{
		{
			name: "scale up requested to match max server replicas",
			loadReq: &pba.AgentSubscribeRequest{
				ServerName: "foo-server",
			},
			timeout:          10 * time.Second,
			modelReplicas:    10,
			scaleUp:          true,
			expectedReplicas: 3,
			notifyReq: &pb.ServerNotifyRequest{
				Servers: []*pb.ServerNotify{
					{
						Name:             "foo-server",
						ExpectedReplicas: 2,
						Shared:           true,
						MaxReplicas:      3,
					},
				},
				IsFirstSync: false,
			},
		},
		{
			name: "scale up requested to match max model replicas - 2",
			loadReq: &pba.AgentSubscribeRequest{
				ServerName: "foo-server",
			},
			timeout:          10 * time.Second,
			modelReplicas:    4,
			scaleUp:          true,
			expectedReplicas: 4,
			notifyReq: &pb.ServerNotifyRequest{
				Servers: []*pb.ServerNotify{
					{
						Name:             "foo-server",
						ExpectedReplicas: 2,
						Shared:           true,
						MaxReplicas:      5,
					},
				},
				IsFirstSync: false,
			},
		},
		{
			name: "scale up not requested",
			loadReq: &pba.AgentSubscribeRequest{
				ServerName: "foo-server",
			},
			timeout:       10 * time.Second,
			modelReplicas: 1,
			scaleUp:       false,
			notifyReq: &pb.ServerNotifyRequest{
				Servers: []*pb.ServerNotify{
					{
						Name:             "server1",
						ExpectedReplicas: 2,
						Shared:           true,
					},
				},
				IsFirstSync: false,
			},
		},
		{
			name: "scale up not requested - expected replicas not set",
			loadReq: &pba.AgentSubscribeRequest{
				ServerName: "foo-server",
			},
			timeout:       10 * time.Second,
			modelReplicas: 1,
			scaleUp:       false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, hub := createTestSchedulerWithConfig(
				t,
				SchedulerServerConfig{
					AutoScalingServerEnabled: true,
				},
			)
			s.timeout = test.timeout

			if test.notifyReq != nil {
				_, err := s.ServerNotify(context.Background(), test.notifyReq)
				g.Expect(err).To(BeNil())
			}

			if test.loadReq != nil {
				err := s.modelStore.AddServerReplica(test.loadReq)
				g.Expect(err).To(BeNil())
				err = s.modelStore.UpdateModel(&pb.LoadModelRequest{
					Model: &pb.Model{
						Meta:           &pb.MetaData{Name: "foo-model"},
						DeploymentSpec: &pb.DeploymentSpec{Replicas: uint32(test.modelReplicas)},
					},
				})
				g.Expect(err).To(BeNil())
				err = s.modelStore.UpdateLoadedModels(
					"foo-model", 1, "foo-server", []*store.ServerReplica{
						store.NewServerReplica("", 8080, 5001, 0, store.NewServer("foo-server", true), []string{}, 100, 100, 0, map[store.ModelVersionID]bool{}, 100),
					},
				)
				g.Expect(err).To(BeNil())
			}

			// to allow events to propagate
			time.Sleep(1 * time.Second)

			stream := newStubServerStatusServer(1, 5*time.Millisecond)

			s.serverEventStream.mu.Lock()
			s.serverEventStream.streams[stream] = &ServerSubscription{
				name:   "dummy",
				stream: stream,
				fin:    make(chan bool),
			}
			s.serverEventStream.mu.Unlock()

			g.Expect(s.serverEventStream.streams[stream]).ToNot(BeNil())

			hub.PublishServerEvent(serverEventHandlerName, coordinator.ServerEventMsg{
				ServerName: "foo-server", UpdateContext: coordinator.SERVER_SCALE_UP,
			})

			if !test.scaleUp {
				g.Expect(s.serverEventStream.streams).To(HaveLen(1))
			} else {
				g.Eventually(stream.msgs).WithTimeout(1 * time.Second).WithPolling(500 * time.Millisecond).Should(HaveLen(1))

				var ssr *pb.ServerStatusResponse
				select {
				case next := <-stream.msgs:
					ssr = next
				default:
					t.Fail()
				}

				g.Expect(ssr).ToNot(BeNil())
				g.Expect(ssr.ServerName).To(Equal("foo-server"))
				g.Expect(s.serverEventStream.streams).To(HaveLen(1))

				if !test.scaleUp {
					g.Expect(ssr.Type).To(Equal(pb.ServerStatusResponse_StatusUpdate))
				} else {
					g.Expect(ssr.ExpectedReplicas).To(Equal(test.expectedReplicas))
					g.Expect(ssr.Type).To(Equal(pb.ServerStatusResponse_ScalingRequest))
				}
			}
		})
	}
}

func TestServerScaleDownEvents(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name       string
		serverName string
		agents     []*pba.AgentSubscribeRequest
		enabled    bool
	}

	tests := []test{
		{
			// this test will create a 2 replicas server
			// and then trigger a scale down event
			name:       "scale down",
			serverName: "server1",
			agents: []*pba.AgentSubscribeRequest{
				{
					ServerName:           "server1",
					ReplicaIdx:           0,
					Shared:               true,
					AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{
						InferenceSvc:      "server1",
						InferenceHttpPort: 1,
						MemoryBytes:       1000,
						Capabilities:      []string{"sklearn"},
					},
				},
				{
					ServerName:           "server1",
					ReplicaIdx:           1,
					Shared:               true,
					AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{
						InferenceSvc:      "server1",
						InferenceHttpPort: 1,
						MemoryBytes:       1000,
						Capabilities:      []string{"sklearn"},
					},
				},
			},
			enabled: true,
		},
		{
			// this test will create a 2 replicas server
			// and then trigger a scale down event
			name:       "scale down - not triggered",
			serverName: "server1",
			agents: []*pba.AgentSubscribeRequest{
				{
					ServerName:           "server1",
					ReplicaIdx:           0,
					Shared:               true,
					AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{
						InferenceSvc:      "server1",
						InferenceHttpPort: 1,
						MemoryBytes:       1000,
						Capabilities:      []string{"sklearn"},
					},
				},
				{
					ServerName:           "server1",
					ReplicaIdx:           1,
					Shared:               true,
					AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{
						InferenceSvc:      "server1",
						InferenceHttpPort: 1,
						MemoryBytes:       1000,
						Capabilities:      []string{"sklearn"},
					},
				},
			},
			enabled: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, event := createTestSchedulerWithConfig(
				t,
				SchedulerServerConfig{
					AutoScalingServerEnabled: test.enabled,
				},
			)

			for _, agent := range test.agents {
				err := s.modelStore.AddServerReplica(agent)
				g.Expect(err).To(BeNil())
			}
			err := s.modelStore.ServerNotify(
				&pb.ServerNotify{
					Name:             test.serverName,
					ExpectedReplicas: 2,
				},
			)
			g.Expect(err).To(BeNil())

			stream := newStubServerStatusServer(1, 200*time.Millisecond)
			s.serverEventStream.streams[stream] = &ServerSubscription{
				name:   "dummy",
				stream: stream,
				fin:    make(chan bool),
			}
			g.Expect(s.serverEventStream.streams[stream]).ToNot(BeNil())

			// to allow events to propagate
			time.Sleep(500 * time.Millisecond)

			// publish a scale event
			event.PublishServerEvent("", coordinator.ServerEventMsg{
				ServerName:    test.serverName,
				UpdateContext: coordinator.SERVER_SCALE_DOWN,
			})

			// to allow events to propagate
			time.Sleep(500 * time.Millisecond)

			var ssr *pb.ServerStatusResponse
			select {
			case next := <-stream.msgs:
				if next.Type == pb.ServerStatusResponse_ScalingRequest {
					ssr = next
				}
			default:
				if test.enabled {
					t.Fail()
				}
			}

			if test.enabled {
				g.Expect(ssr).ToNot(BeNil())
				g.Expect(ssr.ServerName).To(Equal(test.serverName))
				g.Expect(ssr.Type).To(Equal(pb.ServerStatusResponse_ScalingRequest))
			} else {
				g.Expect(ssr).To(BeNil())
			}

		})
	}
}

func createTestScheduler(t *testing.T) (*SchedulerServer, *coordinator.EventHub) {
	return createTestSchedulerWithConfig(t, SchedulerServerConfig{})
}

func createTestSchedulerWithConfig(t *testing.T, config SchedulerServerConfig) (*SchedulerServer, *coordinator.EventHub) {
	return createTestSchedulerImpl(t, config)
}

func createTestSchedulerImpl(t *testing.T, config SchedulerServerConfig) (*SchedulerServer, *coordinator.EventHub) {
	logger := log.New()
	logger.SetLevel(log.WarnLevel)

	eventHub, _ := coordinator.NewEventHub(logger)

	schedulerStore := store.NewTestMemory(t, logger, store.NewLocalSchedulerStore(), eventHub)
	experimentServer := experiment.NewExperimentServer(logger, eventHub, nil, nil)
	pipelineServer := pipeline.NewPipelineStore(logger, eventHub, schedulerStore)

	scheduler := scheduler2.NewSimpleScheduler(
		logger,
		schedulerStore,
		scheduler2.DefaultSchedulerConfig(schedulerStore),
		synchroniser.NewSimpleSynchroniser(time.Duration(10*time.Millisecond)),
		eventHub,
	)

	modelGwLoadBalancer := util.NewRingLoadBalancer(1)
	pipelineGwLoadBalancer := util.NewRingLoadBalancer(1)
	s := NewSchedulerServer(
		logger, schedulerStore, experimentServer, pipelineServer, scheduler,
		eventHub, synchroniser.NewSimpleSynchroniser(time.Duration(10*time.Millisecond)), config,
		"", "", modelGwLoadBalancer, pipelineGwLoadBalancer, tls.TLSOptions{},
	)

	return s, eventHub
}

func createStream(name string, isModelGateway bool) (*stubModelStatusServer, *ModelSubscription) {
	stream := newStubModelStatusServer(20, 0)
	subscription := ModelSubscription{
		name:           name,
		stream:         stream,
		fin:            make(chan bool),
		isModelGateway: isModelGateway,
	}
	return stream, &subscription
}

func TestModelGwRebalanceMessage(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name              string
		loadReq           *pb.LoadModelRequest
		modelState        store.ModelState
		availableReplicas uint32
		KeepTopics        bool
	}

	tests := []test{
		{
			name: "rebalance model available",
			loadReq: &pb.LoadModelRequest{
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo"},
				},
			},
			modelState:        store.ModelAvailable,
			availableReplicas: 1,
			KeepTopics:        false,
		},
		{
			name: "rebalance model progressing",
			loadReq: &pb.LoadModelRequest{
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo"},
				},
			},
			modelState:        store.ModelProgressing,
			availableReplicas: 1,
			KeepTopics:        false,
		},
		{
			name: "rebalance model terminating",
			loadReq: &pb.LoadModelRequest{
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo"},
				},
			},
			modelState:        store.ModelTerminating,
			availableReplicas: 0,
			KeepTopics:        false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create a test scheduler - note it uses a load balancer with 1 partition
			s, _ := createTestScheduler(t)

			// create modelgw stream
			stream, subscription := createStream("dummy", true)
			s.modelEventStream.streams[stream] = subscription
			g.Expect(s.modelEventStream.streams[stream]).ToNot(BeNil())

			// add stream to the load balancer
			s.modelGwLoadBalancer.AddServer(subscription.name)

			// add a model to the store
			err := s.modelStore.UpdateModel(test.loadReq)
			g.Expect(err).To(BeNil())

			// set the model to available
			modelName := test.loadReq.Model.Meta.Name
			model, _ := s.modelStore.GetModel(modelName)

			mem, ok := s.modelStore.(*store.TestMemoryStore)
			g.Expect(ok).To(BeTrue())

			err = mem.DirectlyUpdateModelStatus(store.ModelID{
				Name:    modelName,
				Version: model.GetLatest().GetVersion(),
			}, store.ModelStatus{
				State:             test.modelState,
				AvailableReplicas: test.availableReplicas,
			})
			g.Expect(err).To(BeNil())

			// trigger rebalance
			s.modelGwRebalance()

			var msr *pb.ModelStatusResponse
			select {
			case next := <-stream.msgs:
				msr = next
			default:
				t.Fail()
			}

			g.Expect(msr).ToNot(BeNil())
			g.Expect(msr.ModelName).To(Equal(modelName))
			g.Expect(msr.Versions).To(HaveLen(1))
			g.Expect(msr.Versions[0].State.AvailableReplicas).To(Equal(test.availableReplicas))
			g.Expect(msr.KeepTopics).To(Equal(test.KeepTopics))
		})
	}
}

func TestModelGwRebalance(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		models   []*pb.LoadModelRequest
		replicas int // number of modelgw instances
	}

	tests := []test{
		{
			name: "rebalance multiple models across N replicas",
			models: []*pb.LoadModelRequest{
				{
					Model: &pb.Model{
						Meta: &pb.MetaData{Name: "foo"},
					},
				},
				{
					Model: &pb.Model{
						Meta: &pb.MetaData{Name: "bar"},
					},
				},
				{
					Model: &pb.Model{
						Meta: &pb.MetaData{Name: "baz"},
					},
				},
			},
			replicas: 4, // test with 4 modelgw replicas
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, _ := createTestScheduler(t)

			var streams []*stubModelStatusServer
			for i := 0; i < test.replicas; i++ {
				name := fmt.Sprintf("dummy%d", i)
				stream, subscription := createStream(name, true)
				s.modelEventStream.streams[stream] = subscription
				s.modelGwLoadBalancer.AddServer(subscription.name)
				streams = append(streams, stream)
				g.Expect(s.modelEventStream.streams[stream]).ToNot(BeNil())
			}

			// Load all models into the store and mark them as available
			for _, req := range test.models {
				err := s.modelStore.UpdateModel(req)
				g.Expect(err).To(BeNil())

				modelName := req.Model.Meta.Name
				model, _ := s.modelStore.GetModel(modelName)

				mem, ok := s.modelStore.(*store.TestMemoryStore)
				g.Expect(ok).To(BeTrue())

				err = mem.DirectlyUpdateModelStatus(store.ModelID{
					Name:    modelName,
					Version: model.GetLatest().GetVersion(),
				}, store.ModelStatus{
					State:             store.ModelAvailable,
					AvailableReplicas: 1,
				})
				g.Expect(err).To(BeNil())
			}

			s.modelGwRebalance()

			modelAssignments := make(map[string]int)
			for _, stream := range streams {
			NextStream:
				for {
					select {
					case msg := <-stream.msgs:
						name := msg.ModelName
						count := int(msg.Versions[0].State.AvailableReplicas)
						modelAssignments[name] += count
					default:
						break NextStream
					}
				}
			}

			// Expect each model to have exactly 1 replica assigned
			g.Expect(modelAssignments).To(HaveLen(len(test.models)))
			for _, req := range test.models {
				modelName := req.Model.Meta.Name
				g.Expect(modelAssignments[modelName]).To(Equal(1),
					fmt.Sprintf("model %q should have exactly 1 replica assigned", modelName))
			}
		})
	}
}

func TestSendModelStatusEvent(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name    string
		loadReq *pb.LoadModelRequest
	}

	tests := []test{
		{
			name: "rebalance model",
			loadReq: &pb.LoadModelRequest{
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo"},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create a test scheduler - note it uses a load balancer with 1 partition
			s, _ := createTestScheduler(t)

			// create operator stream
			operatorStream, operatorSubscription := createStream("dummy_operator", false)
			s.modelEventStream.streams[operatorStream] = operatorSubscription
			g.Expect(s.modelEventStream.streams[operatorStream]).ToNot(BeNil())

			// create first modelgw stream
			firstMgwStream, firstMgwSubscription := createStream("dummy_mgw_1", true)
			s.modelEventStream.streams[firstMgwStream] = firstMgwSubscription
			g.Expect(s.modelEventStream.streams[firstMgwStream]).ToNot(BeNil())

			// create second modelgw stream
			secondMgwStream, secondMgwSubscription := createStream("dummy_mgw_2", true)
			s.modelEventStream.streams[secondMgwStream] = secondMgwSubscription
			g.Expect(s.modelEventStream.streams[secondMgwStream]).ToNot(BeNil())

			// add modelgw streams to the load balancer
			s.modelGwLoadBalancer.AddServer(firstMgwSubscription.name)
			s.modelGwLoadBalancer.AddServer(secondMgwSubscription.name)

			// add a model to the store
			err := s.modelStore.UpdateModel(test.loadReq)
			g.Expect(err).To(BeNil())

			// set the model to available
			modelName := test.loadReq.Model.Meta.Name
			model, _ := s.modelStore.GetModel(modelName)

			mem, ok := s.modelStore.(*store.TestMemoryStore)
			g.Expect(ok).To(BeTrue())

			err = mem.DirectlyUpdateModelStatus(store.ModelID{
				Name:    modelName,
				Version: model.GetLatest().GetVersion(),
			}, store.ModelStatus{
				State:             store.ModelAvailable,
				AvailableReplicas: 1,
			})
			g.Expect(err).To(BeNil())

			err = s.sendModelStatusEvent(coordinator.ModelEventMsg{
				ModelName:    modelName,
				ModelVersion: 1,
			})
			g.Expect(err).To(BeNil())

			// read message from all streams
			messages := make([]*pb.ModelStatusResponse, 0, 3)
			for _, stream := range []*stubModelStatusServer{operatorStream, firstMgwStream, secondMgwStream} {
				select {
				case next := <-stream.msgs:
					messages = append(messages, next)
				default:
				}
			}

			// we only expect to get message from the operator and one modelgw streams
			g.Expect(messages).To(HaveLen(2))
		})
	}
}
