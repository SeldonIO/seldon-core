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

func TestModelsStatusEvents(t *testing.T) {
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
		loadReq *pb.LoadModelRequest
		timeout time.Duration
	}

	tests := []test{
		{
			name: "model ok",
			loadReq: &pb.LoadModelRequest{
				Model: &pb.Model{
					Meta: &pb.MetaData{Name: "foo"},
				},
			},
			timeout: 1 * time.Millisecond,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, hub := createTestScheduler()
			if test.loadReq != nil {
				err := s.modelStore.UpdateModel(test.loadReq)
				g.Expect(err).To(BeNil())
			}

			stream := newStubModelStatusServer(1, test.timeout)
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

		})
	}
}
