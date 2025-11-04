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
	state_generator "github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state generator"
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
	store               KVStore
	modelStateGenerator state_generator.Model
}

func NewLoadModelEventHandler(store KVStore) *LoadModelEventHandler {
	return &LoadModelEventHandler{store: store, modelStateGenerator: state_generator.NewModelStateGenerator()}
}

// Handle implementations (business logic goes here)
func (e *LoadModelEventHandler) Handle(ctx context.Context, event Event) ([]OutputEvent, error) {
	loadEvent, ok := event.(*LoadModelEvent)
	if !ok {
		return nil, fmt.Errorf("invalid event type, expected LoadModelEvent")
	}

	// 1. Load current cluster state from KVStore

	// 2. Call pure business logic from state machine statemachine.ApplyLoadModel

	e.modelStateGenerator.ApplyLoadModel()
	// 3. Get resulting state in cluster

	// 4. convert domain events into infrastructure events

	// 5. save cluster state and fan out events

	/*
		Errors:
			- event validation failed
		Events
			- Models Event Message (currently used to maintain replicated store in Agent and to update downstream)
				- status with failed scheduling due to replicas and server
				- normal status created
			- Load Models (for Agent)
				- Server Scale UP
				- server Scale Down
				- Unload Models versions
			-

	*/

	return []OutputEvent{}, nil
}
