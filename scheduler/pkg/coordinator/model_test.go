package coordinator

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestRegisterHandler(t *testing.T) {
	type test struct {
		name        string
		numHandlers int
	}

	tests := []test{
		{
			name:        "0 handlers do not block publishing",
			numHandlers: 0,
		},
		{
			name:        "1 handler receives events",
			numHandlers: 1,
		},
		{
			name:        "n handlers all receive all events",
			numHandlers: 20,
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
				h.RegisterModelEventHandler(
					fmt.Sprintf("handler-%d", i),
					queueSize,
					l,
					func(event ModelEventMsg) { atomic.AddInt64(&counter, 1) },
				)
			}

			for i := 0; i < numEvents; i++ {
				h.PublishModelEvent(
					eventSource,
					ModelEventMsg{},
				)
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
	}

	tests := []test{
		{
			name:            "Events published before handler registration are ignored",
			numPublishers:   1,
			numEventsBefore: 10,
			numEventsAfter:  0,
		},
		{
			name:            "Events published after handler registration are received",
			numPublishers:   1,
			numEventsBefore: 0,
			numEventsAfter:  10,
		},
		{
			name:            "Only events published after handler registration are received",
			numPublishers:   1,
			numEventsBefore: 10,
			numEventsAfter:  10,
		},
		{
			name:            "Multiple producers can publish to the same topic",
			numPublishers:   10,
			numEventsBefore: 10,
			numEventsAfter:  10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := log.New()
			h, err := NewEventHub(l)
			require.Nil(t, err)

			for b := 0; b < tt.numEventsBefore; b++ {
				for p := 0; p < tt.numPublishers; p++ {
					h.PublishModelEvent(
						fmt.Sprintf("publisher-%d", p),
						ModelEventMsg{},
					)
				}
			}

			var counter int64
			queueSize := tt.numPublishers * (tt.numEventsBefore + tt.numEventsAfter)
			h.RegisterModelEventHandler(
				"test.handler",
				queueSize,
				l,
				func(event ModelEventMsg) { atomic.AddInt64(&counter, 1) },
			)

			for a := 0; a < tt.numEventsAfter; a++ {
				for p := 0; p < tt.numPublishers; p++ {
					h.PublishModelEvent(
						fmt.Sprintf("publisher-%d", p),
						ModelEventMsg{},
					)
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
	}

	tests := []test{
		{
			name:            "Immediate close is allowed",
			numEventsBefore: 0,
			numEventsAfter:  10,
			registerBefore:  true,
			expectedCount:   0,
		},
		{
			name:            "No events should be received after calling Close",
			numEventsBefore: 10,
			numEventsAfter:  10,
			registerBefore:  true,
			expectedCount:   10,
		},
		{
			name:            "No events should be received by a handler registered after calling Close",
			numEventsBefore: 10,
			numEventsAfter:  10,
			registerBefore:  false,
			expectedCount:   0,
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
				h.RegisterModelEventHandler(
					"test.handler",
					queueSize,
					l,
					func(event ModelEventMsg) { atomic.AddInt64(&counter, 1) },
				)
			}

			if tt.registerBefore {
				register()
			}

			for b := 0; b < tt.numEventsBefore; b++ {
				h.PublishModelEvent(
					"test.publisher",
					ModelEventMsg{},
				)
			}

			h.Close()

			if !tt.registerBefore {
				register()
			}

			for a := 0; a < tt.numEventsAfter; a++ {
				h.PublishModelEvent(
					"test.publisher",
					ModelEventMsg{},
				)
			}

			// Handlers are async, so need a little while to run
			time.Sleep(5 * time.Millisecond)

			// Some events may not have been handled before close,
			// hence the non-strict inequality.
			require.LessOrEqual(t, tt.expectedCount, int(counter))
		})
	}
}
