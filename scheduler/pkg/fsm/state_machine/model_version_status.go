package state_machine

import pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

type ModelVersionStatus struct {
	*pb.ModelVersionStatus
}

func NewModelVersion(proto *pb.ModelVersionStatus) *ModelVersionStatus {
	return &ModelVersionStatus{ModelVersionStatus: proto}
}

func (mvs *ModelVersionStatus) GetReplica(id int32) *ModelReplica {
	if proto, exists := mvs.ModelReplicaState[id]; exists {
		return NewModelReplica(proto)
	}
	return nil
}

// createInitialModelVersion creates a fresh model version in unknown state
func createInitialModelVersion(model *pb.Model, version uint32) *ModelVersionStatus {
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
