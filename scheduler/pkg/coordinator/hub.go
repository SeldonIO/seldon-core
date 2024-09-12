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
	"sync"
	"sync/atomic"

	busV3 "github.com/mustafaturan/bus/v3"
	log "github.com/sirupsen/logrus"
)

const (
	topicModelEvents      = "model.event"
	topicServerEvents     = "server.event"
	topicExperimentEvents = "experiment.event"
	topicPipelineEvents   = "pipeline.event"
)

type SequenceGenerator struct {
	counter int64
}

func (g *SequenceGenerator) Generate() string {
	next := atomic.AddInt64(&g.counter, 1)
	return fmt.Sprintf("%d", next)
}

var _ busV3.IDGenerator = (*SequenceGenerator)(nil)

type EventHub struct {
	bus                            *busV3.Bus
	logger                         log.FieldLogger
	modelEventHandlerChannels      []chan ModelEventMsg
	serverEventHandlerChannels     []chan ServerEventMsg
	experimentEventHandlerChannels []chan ExperimentEventMsg
	pipelineEventHandlerChannels   []chan PipelineEventMsg
	lock                           sync.RWMutex
	closed                         bool
}

// NewEventHub creates a new EventHub with topics pre-registered.
// The logger l does not need fields preset.
func NewEventHub(l log.FieldLogger) (*EventHub, error) {
	generator := &SequenceGenerator{}
	bus, err := busV3.NewBus(generator)
	if err != nil {
		return nil, err
	}

	hub := EventHub{
		logger: l.WithField("source", "EventHub"),
		bus:    bus,
	}

	hub.bus.RegisterTopics(topicModelEvents, topicServerEvents, topicExperimentEvents, topicPipelineEvents)

	return &hub, nil
}

func (h *EventHub) Close() {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.closed = true

	for _, c := range h.modelEventHandlerChannels {
		close(c)
	}

	for _, c := range h.serverEventHandlerChannels {
		close(c)
	}

	for _, c := range h.experimentEventHandlerChannels {
		close(c)
	}

	for _, c := range h.pipelineEventHandlerChannels {
		close(c)
	}
}
