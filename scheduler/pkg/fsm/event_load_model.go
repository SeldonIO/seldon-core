package fsm

import (
	"context"
	"fmt"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
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
}

func NewLoadModelEventHandler(store KVStore) *LoadModelEventHandler {
	return &LoadModelEventHandler{store: store}
}

// Handle implementations (business logic goes here)
func (e *LoadModelEventHandler) Handle(ctx context.Context, event Event) ([]OutputEvent, error) {
	loadEvent, ok := event.(*LoadModelEvent)
	if !ok {
		return nil, fmt.Errorf("invalid event type")
	}

	// save new model to storage if new

	// if not new update model storage

	// generate model event status for other services to ingest

	// schedule model into server

	// retrieve servers and their models

	// filter out the server

	// find a suitable server to schedule

	// generate agent schedule event

	// generate event to update pipeline status if pipeline is affected by model

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

	// TODO: Implement business logic
	// - Validate request
	// - Update model state in KVStore
	// - Generate output events (e.g., "ModelLoadStarted", "NotifyServer", etc.)

	_ = loadEvent // use the event

	return []OutputEvent{}, nil
}
