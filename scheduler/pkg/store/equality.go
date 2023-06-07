/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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
