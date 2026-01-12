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
