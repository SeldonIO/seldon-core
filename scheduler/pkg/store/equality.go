/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package store

import (
	"reflect"
	"slices"

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
		me := ModelEquality{}
		if !proto.Equal(model1.GetMeta(), model2.GetMeta()) {
			me.MetaDiffers = true
		}
		if !modelSpecEqual(model1.GetModelSpec(), model2.GetModelSpec()) {
			me.ModelSpecDiffers = true
		}
		if !proto.Equal(model1.GetDeploymentSpec(), model2.GetDeploymentSpec()) {
			me.DeploymentSpecDiffers = true
		}
		me.Equal = !me.MetaDiffers && !me.ModelSpecDiffers && !me.DeploymentSpecDiffers
		return me
	}
}

func modelSpecEqual(modelSpec1 *pb.ModelSpec, modelSpec2 *pb.ModelSpec) bool {
	if modelSpec1 == nil && modelSpec2 == nil {
		return true
	} else if modelSpec1 == nil && modelSpec2 != nil {
		return false
	} else if modelSpec1 != nil && modelSpec2 == nil {
		return false
	}

	if modelSpec1.Uri != modelSpec2.Uri {
		return false
	} else if modelSpec1.ArtifactVersion != modelSpec2.ArtifactVersion {
		return false
	} else if !proto.Equal(modelSpec1.StorageConfig, modelSpec2.StorageConfig) {
		return false
	} else if !slices.Equal(modelSpec1.Requirements, modelSpec2.Requirements) {
		return false
	} else if modelSpec1.MemoryBytes != modelSpec2.MemoryBytes {
		return false
	} else if modelSpec1.Server != modelSpec2.Server {
		return false
	} else if !reflect.DeepEqual(modelSpec1.Parameters, modelSpec2.Parameters) {
		return false
	}
	return true
}
