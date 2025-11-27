/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package state_machine

import (
	"time"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/experiment"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/model"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/pipeline"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/server"
)

// ClusterState represents the state of things in the cluster needed for an event to be applied
type ClusterState struct {
	Models      map[string]*model.Snapshot
	Servers     map[string]*server.Snapshot
	Pipelines   map[string]*pipeline.Snapshot
	Experiments map[string]*experiment.Snapshot
}

// todo: this could be added to each cr to separate internal status to k8s status conditions
// todo: would be good for ModelStatusChanged child event
type StatusCondition struct {
	Type               string    `json:"type"`   // "Ready", "Available", "Failed"
	Status             string    `json:"status"` // "True", "False", "Unknown"
	LastTransitionTime time.Time `json:"last_transition_time"`
	Reason             string    `json:"reason"`
	Message            string    `json:"message"`
}
