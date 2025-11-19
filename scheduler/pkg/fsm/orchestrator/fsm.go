/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package orchestrator

import (
	"context"
	"fmt"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/events"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/orchestrator/handlers"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/storage"
)

// Fsm is the finite state machine that processes events
type Fsm struct {
	store        storage.ClusterManager
	stateMachine *state_machine.StateMachine
	handlers     map[events.EventType]Handler
}

type Handler interface {
	Handle(ctx context.Context, ev events.Event) ([]events.OutputEvent, error)
}

func NewFSM(name string, store storage.ClusterManager, config *state_machine.Config) *Fsm {
	fsm := &Fsm{
		store:        store,
		stateMachine: state_machine.NewStateMachine(config),
		handlers:     make(map[events.EventType]Handler),
	}

	// Register default handlers
	fsm.RegisterHandler(events.EventTypeLoadModel, handlers.NewLoadModelEventHandler(store, state_machine.NewStateMachine(config).Model))

	return fsm
}

func (f *Fsm) RegisterHandler(eventType events.EventType, handler Handler) {
	f.handlers[eventType] = handler
}

// Apply processes an event through the FSM
// 1. Applies business logic and updates KVStore
// 2. Generates output events
func (f *Fsm) Apply(ctx context.Context, event events.Event) ([]events.OutputEvent, error) {

	// 2. Get handler for this event type
	handler, exists := f.handlers[event.Type()]
	if !exists {
		return nil, fmt.Errorf("no handler registered for event type: %s", event.Type())
	}

	// 3. Process event and update KVStore
	outputEvents, err := handler.Handle(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("handler failed: %w", err)
	}

	return outputEvents, nil
}
