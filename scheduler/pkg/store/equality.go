package store

import (
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"google.golang.org/protobuf/proto"
)

type ModelEquality struct {
	Equal                 bool
	MetaDiffers           bool
	ModelSpecDiffers      bool
	DeploymentSpecDiffers bool
}

func ModelEqualityCheck(model1 *pb.Model, model2 *pb.Model) ModelEquality {
	if proto.Equal(model1, model2) {
		return ModelEquality{Equal: true}
	} else {
		me := ModelEquality{Equal: false}
		if !proto.Equal(model1.GetMeta(), model2.GetMeta()) {
			me.MetaDiffers = true
		}
		if !proto.Equal(model1.GetModelSpec(), model2.GetModelSpec()) {
			me.ModelSpecDiffers = true
		}
		if !proto.Equal(model1.GetDeploymentSpec(), model2.GetDeploymentSpec()) {
			me.DeploymentSpecDiffers = true
		}
		return me
	}
}
