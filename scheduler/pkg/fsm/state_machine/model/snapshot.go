/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package model

import (
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

// Here we will have the Model snapshot wrapper and basic methods

type Snapshot struct {
	*pb.ModelSnapshot
}

func NewSnapshot(proto *pb.ModelSnapshot) *Snapshot {
	return &Snapshot{ModelSnapshot: proto}
}

// CreateSnapshotFromModel creates a new model snapshot with initial version
// Uses K8s generation as the starting version to ensure monotonic version numbers
func CreateSnapshotFromModel(model *pb.Model) *Snapshot {
	version := calculateInitialVersion(model)

	return NewSnapshot(&pb.ModelSnapshot{
		Versions: []*pb.ModelVersionStatus{
			createInitialModelVersion(model, version).ModelVersionStatus,
		},
		Deleted: false,
	})
}

func (ms *Snapshot) NewVersion(model *pb.Model) *Snapshot {
	lastModelVersion := ms.GetLatestModelVersionStatus()
	if lastModelVersion == nil {
		return ms
	}

	ms.Versions = append(ms.Versions, createInitialModelVersion(model, lastModelVersion.Version+1).ModelVersionStatus)

	return ms
}

// calculateInitialVersion extracts the K8s generation or defaults to 1
// This ensures version numbers never reset when recreating models
func calculateInitialVersion(model *pb.Model) uint32 {
	generation := model.GetMeta().GetKubernetesMeta().GetGeneration()
	return max(uint32(1), uint32(generation))
}
