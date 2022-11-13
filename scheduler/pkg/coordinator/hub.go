/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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

	hub.bus.RegisterTopics(topicModelEvents, topicExperimentEvents, topicPipelineEvents)

	return &hub, nil
}

func (h *EventHub) Close() {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.closed = true

	for _, c := range h.modelEventHandlerChannels {
		close(c)
	}

	for _, c := range h.experimentEventHandlerChannels {
		close(c)
	}

	for _, c := range h.pipelineEventHandlerChannels {
		close(c)
	}
}
