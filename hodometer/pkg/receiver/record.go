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

package receiver

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type Recorder interface {
	Record(event *Event)
	Details() []interface{}
}

type NoopRecorder struct {
}

var _ Recorder = (*NoopRecorder)(nil)

func NewNoopRecorder() *NoopRecorder {
	return &NoopRecorder{}
}

func (nr *NoopRecorder) Record(_ *Event) {}

func (nr *NoopRecorder) Details() []interface{} {
	// Non-empty list can be represented as JSON, etc.
	return []interface{}{}
}

type OrderedRecorder struct {
	mu     sync.RWMutex
	events []*Event
	logger logrus.FieldLogger
}

var _ Recorder = (*OrderedRecorder)(nil)

func NewOrderedRecorder(l logrus.FieldLogger) *OrderedRecorder {
	return &OrderedRecorder{
		mu:     sync.RWMutex{},
		events: []*Event{},
		logger: l.WithField("source", "OrderedRecorder"),
	}
}

func (or *OrderedRecorder) Record(event *Event) {
	or.mu.Lock()
	or.events = append(or.events, event)
	or.mu.Unlock()
}

func (or *OrderedRecorder) Details() []interface{} {
	or.mu.RLock()
	details := make([]interface{}, len(or.events))
	for i, e := range or.events {
		details[i] = e
	}
	or.mu.RUnlock()

	return details
}

type CountingRecorder struct {
	nested    Recorder
	mu        sync.RWMutex
	numEvents uint
}

var _ Recorder = (*CountingRecorder)(nil)

func NewCountingRecorder(r Recorder) *CountingRecorder {
	return &CountingRecorder{
		nested:    r,
		mu:        sync.RWMutex{},
		numEvents: uint(0),
	}
}

func (cr *CountingRecorder) Record(event *Event) {
	cr.mu.Lock()
	cr.numEvents++
	cr.mu.Unlock()

	cr.nested.Record(event)
}

func (cr *CountingRecorder) Details() []interface{} {
	cr.mu.RLock()
	numEvents := cr.numEvents
	cr.mu.RUnlock()

	return []interface{}{
		map[string]interface{}{
			"totalEvents": numEvents,
		},
	}
}
