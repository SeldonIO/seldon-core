/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package fsm

import (
	"context"
	"fmt"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/statemachine"
)

// Concrete event types
type LoadModelEvent struct {
	Request *pb.LoadModelRequest
}

func (e *LoadModelEvent) Type() EventType {
	return EventTypeUnloadModel
}

func (e *LoadModelEvent) Marshal() ([]byte, error) {
	// TODO: implement proto marshaling
	return nil, fmt.Errorf("not implemented")
}

type LoadModelEventHandler struct {
	store KVStore
	fsm   statemachine.FSM
}

func NewLoadModelEventHandler(store KVStore) *LoadModelEventHandler {
	return &LoadModelEventHandler{store: store, fsm: statemachine.NewStateMachine()}
}

// Handle implementations (business logic goes here)
func (e *LoadModelEventHandler) Handle(ctx context.Context, event Event) ([]OutputEvent, error) {
	loadEvent, ok := event.(*LoadModelEvent)
	if !ok {
		return nil, fmt.Errorf("invalid event type, expected LoadModelEvent")
	}

	// 1. Load current cluster state from KVStore

	// 2. Call pure business logic from state machine statemachine.ApplyLoadModel

	// 3. Get resulting state in cluster

	// 4. convert domain events into infrastructure events

	// 5. save cluster state and fan out events

	/*
		Errors:
			- event validation failed
		Events
			- Model Event Message (currently used to maintain replicated store in Agent and to update downstream)
				- status with failed scheduling due to replicas and server
				- normal status created
			- Load Model (for Agent)
				- Server Scale UP
				- server Scale Down
				- Unload Model versions
			-

	*/

	return []OutputEvent{}, nil
}
