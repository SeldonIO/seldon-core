/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

// TODO: explain synchroniser logic here
package synchroniser

import (
	"sync"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
)

func TestServersSyncSynchroniser(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name             string
		timeout          time.Duration
		signals          uint
		initialSignals   uint
		isTimeout        bool
		isDuplicateEvent bool
		context          coordinator.ServerEventUpdateContext
	}

	tests := []test{
		{
			name:           "Simple",
			timeout:        200 * time.Millisecond,
			signals:        5,
			initialSignals: 0,
			context:        coordinator.SERVER_REPLICA_CONNECTED,
		},
		{
			name:           "Events before signals",
			timeout:        200 * time.Millisecond,
			signals:        5,
			initialSignals: 5,
			context:        coordinator.SERVER_REPLICA_CONNECTED,
		},
		{
			name:           "All events before signals",
			timeout:        200 * time.Millisecond,
			signals:        0,
			initialSignals: 5,
			context:        coordinator.SERVER_REPLICA_CONNECTED,
		},
		{
			name:           "All events before signals",
			timeout:        200 * time.Millisecond,
			signals:        0,
			initialSignals: 5,
			isTimeout:      true,
			context:        coordinator.SERVER_REPLICA_CONNECTED,
		},
		{
			name:           "No signals",
			timeout:        200 * time.Millisecond,
			signals:        0,
			initialSignals: 0,
			context:        coordinator.SERVER_REPLICA_CONNECTED,
		},
		{
			name:             "Duplicate events",
			timeout:          200 * time.Millisecond,
			signals:          5,
			initialSignals:   0,
			isDuplicateEvent: true,
			context:          coordinator.SERVER_REPLICA_CONNECTED,
		},
		{
			name:           "wrong context",
			timeout:        200 * time.Millisecond,
			signals:        5,
			initialSignals: 0,
			context:        coordinator.SERVER_STATUS_UPDATE,
		},
		{
			name:           "long timeout", // timeout will not be reached
			timeout:        200 * time.Second,
			signals:        5,
			initialSignals: 0,
			context:        coordinator.SERVER_REPLICA_CONNECTED,
		},
		{
			name:           "long timeout - no signal", // timeout will not be reached
			timeout:        200 * time.Second,
			signals:        0,
			initialSignals: 0,
			context:        coordinator.SERVER_REPLICA_CONNECTED,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())
			s := NewServerBasedSynchroniser(eventHub, logger, test.timeout)

			if !test.isTimeout {
				for i := 0; i < int(test.initialSignals); i++ {
					idx := uint32(i)
					if test.isDuplicateEvent {
						idx = 0
					}
					go eventHub.PublishServerEvent(
						"test",
						coordinator.ServerEventMsg{
							ServerName:    "test",
							ServerIdx:     idx,
							UpdateContext: test.context,
						},
					)
				}
			}

			s.Signals(test.signals + test.initialSignals)
			// this should have no effect
			s.Signals(10000)

			if !test.isTimeout {
				for i := 0; i < int(test.signals); i++ {
					idx := uint32(i) + uint32(test.initialSignals)
					if test.isDuplicateEvent {
						idx = 0
					}
					go eventHub.PublishServerEvent(
						"test",
						coordinator.ServerEventMsg{
							ServerName:    "test",
							ServerIdx:     idx,
							UpdateContext: test.context,
						},
					)
				}
			}

			if test.signals+test.initialSignals > 0 {
				g.Expect(s.IsReady()).To(BeFalseBecause("still need to process events"))
			} else {
				g.Expect(s.IsReady()).To(BeTrueBecause("no events to process"))
			}

			time.Sleep(100 * time.Millisecond)

			// simulate multiple events
			numWaitReady := 10
			waitReadyWg := sync.WaitGroup{}
			waitReadyWg.Add(numWaitReady)
			waitReadyFn := func() {
				s.WaitReady()
				waitReadyWg.Done()
			}
			for i := 0; i < numWaitReady; i++ {
				go waitReadyFn()
			}
			waitReadyWg.Wait()
			g.Expect(s.IsReady()).To(BeTrue())

			if test.isTimeout || test.isDuplicateEvent || test.context != coordinator.SERVER_REPLICA_CONNECTED {
				g.Expect(len(s.connectedServers)).Should(BeNumerically("<", test.signals+test.initialSignals))
			} else {
				g.Expect(len(s.connectedServers)).To(Equal(int(test.signals + test.initialSignals)))
			}
			// make sure we are graceful after this point

			s.Signals(10)
			g.Expect(s.IsReady()).To(BeTrue())
			s.WaitReady()
			g.Expect(s.IsReady()).To(BeTrue())
		})
	}
}
