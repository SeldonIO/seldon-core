/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package model

import pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

type Replica struct {
	*pb.ModelReplicaStatus
}

func NewModelReplica(proto *pb.ModelReplicaStatus) *Replica {
	return &Replica{ModelReplicaStatus: proto}
}

type ReplicaStatus struct {
	*pb.ModelReplicaStatus
}

func NewModelReplicaStatus(proto *pb.ModelReplicaStatus) *ReplicaStatus {
	return &ReplicaStatus{ModelReplicaStatus: proto}
}
