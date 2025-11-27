package output

import (
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/events"
)

type LoadModelVersion struct {
	ModelName string
	Model     pb.Model
	Version   int
}

const EventTypeLoadModelVersion events.EventType = "LoadModelVersion"

func (lmv *LoadModelVersion) Type() events.EventType {
	return EventTypeLoadModelVersion
}
