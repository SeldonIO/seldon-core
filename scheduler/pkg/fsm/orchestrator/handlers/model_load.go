/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package handlers

import (
	"context"
	"fmt"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/events"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/storage"
)

type LoadModelEventHandler struct {
	store               storage.ClusterManager
	modelStateGenerator state_machine.Model
}

func NewLoadModelEventHandler(store storage.ClusterManager) *LoadModelEventHandler {
	return &LoadModelEventHandler{store: store, modelStateGenerator: state_machine.NewModelStateGenerator()}
}

// Handle implementations (business logic goes here)
func (e *LoadModelEventHandler) Handle(ctx context.Context, event events.Event) ([]events.OutputEvent, error) {
	loadEvent, ok := event.(*events.LoadModel)
	if !ok {
		return nil, fmt.Errorf("invalid event type, expected LoadModelEvent")
	}

	// 1. Load current cluster state from KVStore
	clusterState, err := e.store.GetClusterState(ctx)
	if err != nil {
		return nil, err
	}

	// 2. Call pure business logic from state machine statemachine.ApplyLoadModel

	// todo: think about the pointer being dereferenced
	futureClusterState, err := e.modelStateGenerator.ApplyLoadModel(clusterState, *loadEvent)
	if err != nil {
		return nil, err
	}

	// 3. Get resulting state in cluster

	// 4. convert domain events into infrastructure events

	// events := GetEvents(currentClusterState, FutureClusterState)

	// 5. save cluster state

	err = e.store.SaveClusterState(ctx, futureClusterState)
	if err != nil {
		return nil, err
	}

	// fan out events

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

	return []events.OutputEvent{}, nil
}
