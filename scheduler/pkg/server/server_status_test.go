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
	"testing"
	"time"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	pba "github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	scheduler2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/scheduler"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/experiment"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/synchroniser"
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
			s, hub := createTestScheduler()
			s.timeout = test.timeout
			if test.loadReq != nil {
				err := s.modelStore.UpdateModel(test.loadReq)
				g.Expect(err).To(BeNil())
			}

			stream := newStubModelStatusServer(1, 5*time.Millisecond)
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
				g.Expect(s.modelEventStream.streams).To(HaveLen(0))
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
			s, _ := createTestScheduler()
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
			s.serverEventStream.streams[stream] = &ServerSubscription{
				name:   "dummy",
				stream: stream,
				fin:    make(chan bool),
			}
			g.Expect(s.serverEventStream.streams[stream]).ToNot(BeNil())

			// to allow events to propagate
			time.Sleep(500 * time.Millisecond)

			if test.err {
				g.Expect(s.serverEventStream.streams).To(HaveLen(0))
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
			s, hub := createTestScheduler()
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
			s.serverEventStream.streams[stream] = &ServerSubscription{
				name:   "dummy",
				stream: stream,
				fin:    make(chan bool),
			}

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
		model      *pb.LoadModelRequest
	}

	tests := []test{
		{
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
			model: &pb.LoadModelRequest{
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "model1"},
					ModelSpec: &pb.ModelSpec{
						Uri:          "gs://model",
						Requirements: []string{"sklearn"},
					},
					DeploymentSpec: &pb.DeploymentSpec{Replicas: 1},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, event := createTestScheduler()
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

			err = s.modelStore.UpdateModel(test.model)
			g.Expect(err).To(BeNil())

			stream := newStubServerStatusServer(1, 5*time.Millisecond)
			s.serverEventStream.streams[stream] = &ServerSubscription{
				name:   "dummy",
				stream: stream,
				fin:    make(chan bool),
			}
			g.Expect(s.serverEventStream.streams[stream]).ToNot(BeNil())

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
				ssr = next
			default:
				t.Fail()
			}

			g.Expect(ssr).ToNot(BeNil())
			g.Expect(ssr.ServerName).To(Equal(test.serverName))

		})
	}
}

func createTestScheduler() (*SchedulerServer, *coordinator.EventHub) {
	logger := log.New()
	logger.SetLevel(log.WarnLevel)

	eventHub, _ := coordinator.NewEventHub(logger)

	schedulerStore := store.NewMemoryStore(logger, store.NewLocalSchedulerStore(), eventHub)
	experimentServer := experiment.NewExperimentServer(logger, eventHub, nil, nil)
	pipelineServer := pipeline.NewPipelineStore(logger, eventHub, schedulerStore)

	scheduler := scheduler2.NewSimpleScheduler(
		logger,
		schedulerStore,
		scheduler2.DefaultSchedulerConfig(schedulerStore),
		synchroniser.NewSimpleSynchroniser(time.Duration(10*time.Millisecond)),
		eventHub,
	)
	s := NewSchedulerServer(
		logger, schedulerStore, experimentServer, pipelineServer, scheduler,
		eventHub, synchroniser.NewSimpleSynchroniser(time.Duration(10*time.Millisecond)), SchedulerServerConfig{})

	return s, eventHub
}
