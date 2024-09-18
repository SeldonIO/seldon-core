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
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	. "github.com/onsi/gomega"
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
	}

	tests := []test{
		{
			name:           "Simple",
			timeout:        100 * time.Millisecond,
			signals:        5,
			initialSignals: 0,
		},
		{
			name:           "Events before signals",
			timeout:        100 * time.Millisecond,
			signals:        5,
			initialSignals: 5,
		},
		{
			name:           "All events before signals",
			timeout:        100 * time.Millisecond,
			signals:        0,
			initialSignals: 5,
		},
		{
			name:           "All events before signals",
			timeout:        100 * time.Millisecond,
			signals:        0,
			initialSignals: 5,
			isTimeout:      true,
		},
		{
			name:           "No signals",
			timeout:        100 * time.Millisecond,
			signals:        0,
			initialSignals: 0,
		},
		{
			name:             "Duplicate events",
			timeout:          100 * time.Millisecond,
			signals:          5,
			initialSignals:   0,
			isDuplicateEvent: true,
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
					eventHub.PublishServerEvent(
						"test",
						coordinator.ServerEventMsg{
							ServerName: "test",
							ServerIdx:  idx,
						},
					)
				}
			}

			s.Signals(test.signals + test.initialSignals)

			if !test.isTimeout {
				for i := 0; i < int(test.signals); i++ {
					idx := uint32(i) + uint32(test.initialSignals)
					if test.isDuplicateEvent {
						idx = 0
					}
					eventHub.PublishServerEvent(
						"test",
						coordinator.ServerEventMsg{
							ServerName: "test",
							ServerIdx:  idx,
						},
					)
				}
			}

			g.Expect(s.IsReady()).To(BeFalse())
			s.WaitReady()
			g.Expect(s.IsReady()).To(BeTrue())
		})
	}
}
