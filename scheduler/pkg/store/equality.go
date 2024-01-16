/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package store

import (
	"google.golang.org/protobuf/proto"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
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
