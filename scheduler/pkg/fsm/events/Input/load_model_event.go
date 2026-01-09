/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package Input

import (
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/events"
)

const EventTypeLoadModel events.EventType = "LoadModel"

type LoadModel struct {
	*pb.LoadModelRequest //todo: change this once we have confidence in what is required
}

func NewLoadModel(req *pb.LoadModelRequest) *LoadModel {
	return &LoadModel{req}
}

func (ms *LoadModel) Type() events.EventType {
	return EventTypeLoadModel
}
