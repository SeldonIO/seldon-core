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
				ModelName: "foo", ModelVersion: 1})

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
			}
		})
	}
}

func TestServersStatusEvents(t *testing.T) {
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
			s, hub := createTestScheduler()
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
			hub.PublishModelEvent(serverModelEventHandlerName, coordinator.ModelEventMsg{
				ModelName: "foo", ModelVersion: 1})

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
			}
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
	)
	s := NewSchedulerServer(logger, schedulerStore, experimentServer, pipelineServer, scheduler, eventHub, synchroniser.NewSimpleSynchroniser(time.Duration(10*time.Millisecond)))

	return s, eventHub
}
