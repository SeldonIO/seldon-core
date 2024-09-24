/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package coordinator

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type eventHubTestOp int

const (
	modelEventHub eventHubTestOp = iota
	serverEventHub
	pipelineEventHub
	experimentEventHub
)

func TestRegisterHandler(t *testing.T) {
	type test struct {
		name        string
		numHandlers int
		op          eventHubTestOp
	}

	tests := []test{
		// models
		{
			name:        "0 handlers do not block publishing - models",
			numHandlers: 0,
			op:          modelEventHub,
		},
		{
			name:        "1 handler receives events - models",
			numHandlers: 1,
			op:          modelEventHub,
		},
		{
			name:        "n handlers all receive all events - models",
			numHandlers: 20,
			op:          modelEventHub,
		},

		// servers
		{
			name:        "0 handlers do not block publishing - servers",
			numHandlers: 0,
			op:          serverEventHub,
		},
		{
			name:        "1 handler receives events - servers",
			numHandlers: 1,
			op:          serverEventHub,
		},
		{
			name:        "n handlers all receive all events - servers",
			numHandlers: 20,
			op:          serverEventHub,
		},

		// pipelines
		{
			name:        "0 handlers do not block publishing - pipelines",
			numHandlers: 0,
			op:          pipelineEventHub,
		},
		{
			name:        "1 handler receives events - pipelines",
			numHandlers: 1,
			op:          pipelineEventHub,
		},
		{
			name:        "n handlers all receive all events - pipelines",
			numHandlers: 20,
			op:          pipelineEventHub,
		},

		// experiments
		{
			name:        "0 handlers do not block publishing - experiments",
			numHandlers: 0,
			op:          experimentEventHub,
		},
		{
			name:        "1 handler receives events - experiments",
			numHandlers: 1,
			op:          experimentEventHub,
		},
		{
			name:        "n handlers all receive all events - experiments",
			numHandlers: 20,
			op:          experimentEventHub,
		},
	}

	queueSize := 100
	numEvents := 10
	eventSource := "test.publisher"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := log.New()
			h, err := NewEventHub(l)
			require.Nil(t, err)

			var counter int64

			for i := 0; i < tt.numHandlers; i++ {
				if tt.op == modelEventHub {
					h.RegisterModelEventHandler(
						fmt.Sprintf("handler-%d", i),
						queueSize,
						l,
						func(event ModelEventMsg) { atomic.AddInt64(&counter, 1) },
					)
				} else if tt.op == serverEventHub {
					h.RegisterServerEventHandler(
						fmt.Sprintf("handler-%d", i),
						queueSize,
						l,
						func(event ServerEventMsg) { atomic.AddInt64(&counter, 1) },
					)
				} else if tt.op == pipelineEventHub {
					h.RegisterPipelineEventHandler(
						fmt.Sprintf("handler-%d", i),
						queueSize,
						l,
						func(event PipelineEventMsg) { atomic.AddInt64(&counter, 1) },
					)
				} else if tt.op == experimentEventHub {
					h.RegisterExperimentEventHandler(
						fmt.Sprintf("handler-%d", i),
						queueSize,
						l,
						func(event ExperimentEventMsg) { atomic.AddInt64(&counter, 1) },
					)
				}
			}

			for i := 0; i < numEvents; i++ {
				if tt.op == modelEventHub {
					h.PublishModelEvent(
						eventSource,
						ModelEventMsg{},
					)
				} else if tt.op == serverEventHub {
					h.PublishServerEvent(
						eventSource,
						ServerEventMsg{},
					)
				} else if tt.op == pipelineEventHub {
					h.PublishPipelineEvent(
						eventSource,
						PipelineEventMsg{},
					)
				} else if tt.op == experimentEventHub {
					h.PublishExperimentEvent(
						eventSource,
						ExperimentEventMsg{},
					)
				}
			}

			// Handlers are async, so need a little while to run
			time.Sleep(5 * time.Millisecond)

			require.Equal(t, tt.numHandlers, len(h.bus.HandlerKeys()))
			require.Equal(t, numEvents*tt.numHandlers, int(counter))
		})
	}
}

func TestPublishModelEvent(t *testing.T) {
	type test struct {
		name            string
		numPublishers   int
		numEventsBefore int
		numEventsAfter  int
		op              eventHubTestOp
	}

	tests := []test{
		// models
		{
			name:            "Events published before handler registration are ignored - models",
			numPublishers:   1,
			numEventsBefore: 10,
			numEventsAfter:  0,
			op:              modelEventHub,
		},
		{
			name:            "Events published after handler registration are received - models",
			numPublishers:   1,
			numEventsBefore: 0,
			numEventsAfter:  10,
			op:              modelEventHub,
		},
		{
			name:            "Only events published after handler registration are received - models",
			numPublishers:   1,
			numEventsBefore: 10,
			numEventsAfter:  10,
			op:              modelEventHub,
		},
		{
			name:            "Multiple producers can publish to the same topic - models",
			numPublishers:   10,
			numEventsBefore: 10,
			numEventsAfter:  10,
			op:              modelEventHub,
		},

		// servers
		{
			name:            "Events published before handler registration are ignored - servers",
			numPublishers:   1,
			numEventsBefore: 10,
			numEventsAfter:  0,
			op:              serverEventHub,
		},
		{
			name:            "Events published after handler registration are received - servers",
			numPublishers:   1,
			numEventsBefore: 0,
			numEventsAfter:  10,
			op:              serverEventHub,
		},
		{
			name:            "Only events published after handler registration are received - servers",
			numPublishers:   1,
			numEventsBefore: 10,
			numEventsAfter:  10,
			op:              serverEventHub,
		},
		{
			name:            "Multiple producers can publish to the same topic - servers",
			numPublishers:   10,
			numEventsBefore: 10,
			numEventsAfter:  10,
			op:              serverEventHub,
		},

		// pipelines
		{
			name:            "Events published before handler registration are ignored - pipelines",
			numPublishers:   1,
			numEventsBefore: 10,
			numEventsAfter:  0,
			op:              pipelineEventHub,
		},
		{
			name:            "Events published after handler registration are received - pipelines",
			numPublishers:   1,
			numEventsBefore: 0,
			numEventsAfter:  10,
			op:              pipelineEventHub,
		},
		{
			name:            "Only events published after handler registration are received - pipelines",
			numPublishers:   1,
			numEventsBefore: 10,
			numEventsAfter:  10,
			op:              pipelineEventHub,
		},
		{
			name:            "Multiple producers can publish to the same topic - pipelines",
			numPublishers:   10,
			numEventsBefore: 10,
			numEventsAfter:  10,
			op:              pipelineEventHub,
		},

		// experiments
		{
			name:            "Events published before handler registration are ignored - experiments",
			numPublishers:   1,
			numEventsBefore: 10,
			numEventsAfter:  0,
			op:              experimentEventHub,
		},
		{
			name:            "Events published after handler registration are received - experiments",
			numPublishers:   1,
			numEventsBefore: 0,
			numEventsAfter:  10,
			op:              experimentEventHub,
		},
		{
			name:            "Only events published after handler registration are received - experiments",
			numPublishers:   1,
			numEventsBefore: 10,
			numEventsAfter:  10,
			op:              experimentEventHub,
		},
		{
			name:            "Multiple producers can publish to the same topic - experiments",
			numPublishers:   10,
			numEventsBefore: 10,
			numEventsAfter:  10,
			op:              experimentEventHub,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := log.New()
			h, err := NewEventHub(l)
			require.Nil(t, err)

			for b := 0; b < tt.numEventsBefore; b++ {
				for p := 0; p < tt.numPublishers; p++ {
					if tt.op == modelEventHub {
						h.PublishModelEvent(
							fmt.Sprintf("publisher-%d", p),
							ModelEventMsg{},
						)
					} else if tt.op == serverEventHub {
						h.PublishServerEvent(
							fmt.Sprintf("publisher-%d", p),
							ServerEventMsg{},
						)
					} else if tt.op == pipelineEventHub {
						h.PublishPipelineEvent(
							fmt.Sprintf("publisher-%d", p),
							PipelineEventMsg{},
						)
					} else if tt.op == experimentEventHub {
						h.PublishExperimentEvent(
							fmt.Sprintf("publisher-%d", p),
							ExperimentEventMsg{},
						)
					}
				}
			}

			var counter int64
			queueSize := tt.numPublishers * (tt.numEventsBefore + tt.numEventsAfter)
			if tt.op == modelEventHub {
				h.RegisterModelEventHandler(
					"test.handler",
					queueSize,
					l,
					func(event ModelEventMsg) { atomic.AddInt64(&counter, 1) },
				)
			} else if tt.op == serverEventHub {
				h.RegisterServerEventHandler(
					"test.handler",
					queueSize,
					l,
					func(event ServerEventMsg) { atomic.AddInt64(&counter, 1) },
				)
			} else if tt.op == pipelineEventHub {
				h.RegisterPipelineEventHandler(
					"test.handler",
					queueSize,
					l,
					func(event PipelineEventMsg) { atomic.AddInt64(&counter, 1) },
				)
			} else if tt.op == experimentEventHub {
				h.RegisterExperimentEventHandler(
					"test.handler",
					queueSize,
					l,
					func(event ExperimentEventMsg) { atomic.AddInt64(&counter, 1) },
				)
			}

			for a := 0; a < tt.numEventsAfter; a++ {
				for p := 0; p < tt.numPublishers; p++ {
					if tt.op == modelEventHub {
						h.PublishModelEvent(
							fmt.Sprintf("publisher-%d", p),
							ModelEventMsg{},
						)
					} else if tt.op == serverEventHub {
						h.PublishServerEvent(
							fmt.Sprintf("publisher-%d", p),
							ServerEventMsg{},
						)
					} else if tt.op == pipelineEventHub {
						h.PublishPipelineEvent(
							fmt.Sprintf("publisher-%d", p),
							PipelineEventMsg{},
						)
					} else if tt.op == experimentEventHub {
						h.PublishExperimentEvent(
							fmt.Sprintf("publisher-%d", p),
							ExperimentEventMsg{},
						)
					}
				}
			}

			// Handlers are async, so need a little while to run
			time.Sleep(5 * time.Millisecond)

			require.Equal(t, tt.numPublishers*tt.numEventsAfter, int(counter))
		})
	}
}

func TestClose(t *testing.T) {
	type test struct {
		name            string
		numEventsBefore int
		numEventsAfter  int
		registerBefore  bool
		expectedCount   int
		op              eventHubTestOp
	}

	tests := []test{
		// models
		{
			name:            "Immediate close is allowed - models",
			numEventsBefore: 0,
			numEventsAfter:  10,
			registerBefore:  true,
			expectedCount:   0,
			op:              modelEventHub,
		},
		{
			name:            "No events should be received after calling Close - models",
			numEventsBefore: 10,
			numEventsAfter:  10,
			registerBefore:  true,
			expectedCount:   10,
			op:              modelEventHub,
		},
		{
			name:            "No events should be received by a handler registered after calling Close - models",
			numEventsBefore: 10,
			numEventsAfter:  10,
			registerBefore:  false,
			expectedCount:   0,
			op:              modelEventHub,
		},

		// servers
		{
			name:            "Immediate close is allowed - servers",
			numEventsBefore: 0,
			numEventsAfter:  10,
			registerBefore:  true,
			expectedCount:   0,
			op:              serverEventHub,
		},
		{
			name:            "No events should be received after calling Close - servers",
			numEventsBefore: 10,
			numEventsAfter:  10,
			registerBefore:  true,
			expectedCount:   10,
			op:              serverEventHub,
		},
		{
			name:            "No events should be received by a handler registered after calling Close - servers",
			numEventsBefore: 10,
			numEventsAfter:  10,
			registerBefore:  false,
			expectedCount:   0,
			op:              serverEventHub,
		},

		// pipelines
		{
			name:            "Immediate close is allowed - pipelines",
			numEventsBefore: 0,
			numEventsAfter:  10,
			registerBefore:  true,
			expectedCount:   0,
			op:              pipelineEventHub,
		},
		{
			name:            "No events should be received after calling Close - pipelines",
			numEventsBefore: 10,
			numEventsAfter:  10,
			registerBefore:  true,
			expectedCount:   10,
			op:              pipelineEventHub,
		},
		{
			name:            "No events should be received by a handler registered after calling Close - pipelines",
			numEventsBefore: 10,
			numEventsAfter:  10,
			registerBefore:  false,
			expectedCount:   0,
			op:              pipelineEventHub,
		},

		// experiments
		{
			name:            "Immediate close is allowed - experiments",
			numEventsBefore: 0,
			numEventsAfter:  10,
			registerBefore:  true,
			expectedCount:   0,
			op:              experimentEventHub,
		},
		{
			name:            "No events should be received after calling Close - experiments",
			numEventsBefore: 10,
			numEventsAfter:  10,
			registerBefore:  true,
			expectedCount:   10,
			op:              experimentEventHub,
		},
		{
			name:            "No events should be received by a handler registered after calling Close - experiments",
			numEventsBefore: 10,
			numEventsAfter:  10,
			registerBefore:  false,
			expectedCount:   0,
			op:              experimentEventHub,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := log.New()
			h, err := NewEventHub(l)
			require.Nil(t, err)

			var counter int64
			queueSize := tt.numEventsBefore + tt.numEventsAfter

			register := func() {
				if tt.op == modelEventHub {
					h.RegisterModelEventHandler(
						"test.handler",
						queueSize,
						l,
						func(event ModelEventMsg) { atomic.AddInt64(&counter, 1) },
					)
				} else if tt.op == serverEventHub {
					h.RegisterServerEventHandler(
						"test.handler",
						queueSize,
						l,
						func(event ServerEventMsg) { atomic.AddInt64(&counter, 1) },
					)
				} else if tt.op == pipelineEventHub {
					h.RegisterPipelineEventHandler(
						"test.handler",
						queueSize,
						l,
						func(event PipelineEventMsg) { atomic.AddInt64(&counter, 1) },
					)
				} else if tt.op == experimentEventHub {
					h.RegisterExperimentEventHandler(
						"test.handler",
						queueSize,
						l,
						func(event ExperimentEventMsg) { atomic.AddInt64(&counter, 1) },
					)
				}
			}

			if tt.registerBefore {
				register()
			}

			for b := 0; b < tt.numEventsBefore; b++ {
				if tt.op == modelEventHub {
					h.PublishModelEvent(
						"test.publisher",
						ModelEventMsg{},
					)
				} else if tt.op == serverEventHub {
					h.PublishServerEvent(
						"test.publisher",
						ServerEventMsg{},
					)
				} else if tt.op == pipelineEventHub {
					h.PublishPipelineEvent(
						"test.publisher",
						PipelineEventMsg{},
					)
				} else if tt.op == experimentEventHub {
					h.PublishExperimentEvent(
						"test.publisher",
						ExperimentEventMsg{},
					)
				}
			}

			h.Close()

			if !tt.registerBefore {
				register()
			}

			for a := 0; a < tt.numEventsAfter; a++ {
				if tt.op == modelEventHub {
					h.PublishModelEvent(
						"test.publisher",
						ModelEventMsg{},
					)
				} else if tt.op == serverEventHub {
					h.PublishServerEvent(
						"test.publisher",
						ServerEventMsg{},
					)
				} else if tt.op == pipelineEventHub {
					h.PublishPipelineEvent(
						"test.publisher",
						PipelineEventMsg{},
					)
				} else if tt.op == experimentEventHub {
					h.PublishExperimentEvent(
						"test.publisher",
						ExperimentEventMsg{},
					)
				}
			}

			// Handlers are async, so need a little while to run
			time.Sleep(5 * time.Millisecond)

			// Some events may not have been handled before close,
			// hence the non-strict inequality.
			require.LessOrEqual(t, tt.expectedCount, int(counter))
		})
	}
}
