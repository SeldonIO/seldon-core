/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package util

import (
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"
)

func NewTestModelVersion(model *pb.Model, version uint32, server string, replicas map[int32]*db.ReplicaStatus, state db.ModelState) *db.ModelVersion {
	return &db.ModelVersion{
		Version:   version,
		ModelDefn: model,
		Server:    server,
		Replicas:  replicas,
		State:     &db.ModelStatus{State: state},
	}
}
