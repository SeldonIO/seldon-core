/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package filters

import (
	"fmt"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"
)

const (
	alibiExplainerRequiredCapability = "alibi-explain"
)

type ExplainerFilter struct{}

func (s ExplainerFilter) Name() string {
	return "ExplainerFilter"
}

func (s ExplainerFilter) Filter(model *db.ModelVersion, replica *db.ServerReplica) bool {
	if model.ModelDefn.ModelSpec.GetExplainer() != nil {
		for _, capability := range replica.GetCapabilities() {
			if alibiExplainerRequiredCapability == capability {
				return true
			}
		}
		return false
	}
	return true
}

func (s ExplainerFilter) Description(model *db.ModelVersion, replica *db.ServerReplica) string {
	return fmt.Sprintf("model is explainer %v replica capabilities %v", model.ModelDefn.GetModelSpec().GetExplainer() == nil, replica.GetCapabilities())
}
