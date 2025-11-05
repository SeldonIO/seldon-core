package state_machine

import pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

type ModelReplica struct {
	*pb.ModelReplicaStatus
}

func NewModelReplica(proto *pb.ModelReplicaStatus) *ModelReplica {
	return &ModelReplica{ModelReplicaStatus: proto}
}

type ModelReplicaStatus struct {
	*pb.ModelReplicaStatus
}

func NewModelReplicaStatus(proto *pb.ModelReplicaStatus) *ModelReplicaStatus {
	return &ModelReplicaStatus{ModelReplicaStatus: proto}
}
