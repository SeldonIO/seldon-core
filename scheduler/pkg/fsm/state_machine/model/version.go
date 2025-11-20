/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package model

import pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

type VersionStatus struct {
	*pb.ModelVersionStatus
}

func NewModelVersion(proto *pb.ModelVersionStatus) *VersionStatus {
	return &VersionStatus{ModelVersionStatus: proto}
}

func (mvs *VersionStatus) GetReplica(id int32) *Replica {
	if proto, exists := mvs.ModelReplicaState[id]; exists {
		return NewModelReplica(proto)
	}
	return nil
}

// createInitialModelVersion creates a fresh model version in unknown state
func createInitialModelVersion(model *pb.Model, version uint32) *VersionStatus {
	return NewModelVersion(&pb.ModelVersionStatus{
		Version:           version,
		ServerName:        "",
		KubernetesMeta:    model.GetMeta().GetKubernetesMeta(),
		ModelReplicaState: make(map[int32]*pb.ModelReplicaStatus),
		State: &pb.ModelStatus{
			State:               pb.ModelStatus_ModelStateUnknown,
			Reason:              "",
			AvailableReplicas:   0,
			UnavailableReplicas: 0,
			LastChangeTimestamp: nil,
			ModelGwState:        pb.ModelStatus_ModelCreate,
			ModelGwReason:       "",
		},
		ModelDefn: model,
	})
}

// IsInactive checks if a model version status has no active replicas
func (mvs *VersionStatus) Active() bool {
	if mvs == nil || len(mvs.ModelReplicaState) == 0 {
		return false
	}

	for _, state := range mvs.ModelReplicaState {
		if NewModelReplicaStatus(state).Active() {
			return true
		}
	}
	return false
}

func (mvs *VersionStatus) HasServer() bool {
	return mvs.ServerName != ""
}
