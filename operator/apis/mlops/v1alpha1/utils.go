/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package v1alpha1

import (
	"fmt"
	"math"
)

type ValidatedScalingSpec struct {
	Replicas    uint32
	MinReplicas uint32
	MaxReplicas uint32
}

func GetValidatedScalingSpec(replicas *int32, minReplicas *int32, maxReplicas *int32) (*ValidatedScalingSpec, error) {
	var validatedSpec ValidatedScalingSpec

	if replicas == nil && minReplicas == nil && maxReplicas == nil {
		validatedSpec.Replicas = 1
		validatedSpec.MinReplicas = 0
		validatedSpec.MaxReplicas = math.MaxUint32
		return &validatedSpec, nil
	}

	if replicas == nil && minReplicas == nil {
		validatedSpec.Replicas = 1
	}

	if replicas != nil && *replicas > 0 {
		validatedSpec.Replicas = uint32(*replicas)
	} else {
		if minReplicas != nil && *minReplicas > 0 {

			if replicas != nil && *replicas < *minReplicas {
				return nil, fmt.Errorf("number of replicas %d cannot be less than minimum replica %d", *replicas, *minReplicas)
			}
			// set replicas to the min replicas when replicas is not set explicitly
			validatedSpec.Replicas = uint32(*minReplicas)
		}
	}

	if minReplicas != nil && *minReplicas > 0 {
		validatedSpec.MinReplicas = uint32(*minReplicas)
		if validatedSpec.Replicas < validatedSpec.MinReplicas {
			return nil, fmt.Errorf("number of replicas %d must be >= min replicas %d", validatedSpec.Replicas, validatedSpec.MinReplicas)
		}
	} else {
		validatedSpec.MinReplicas = 0
	}

	if maxReplicas != nil && *maxReplicas > 0 {
		validatedSpec.MaxReplicas = uint32(*maxReplicas)
		if validatedSpec.MinReplicas > validatedSpec.MaxReplicas {
			return nil, fmt.Errorf("min number of replicas %d must be <= max number of replicas %d", validatedSpec.MinReplicas, validatedSpec.MaxReplicas)
		}
		if validatedSpec.Replicas > validatedSpec.MaxReplicas {
			return nil, fmt.Errorf("number of replicas %d must be <= max replicas %d", validatedSpec.Replicas, validatedSpec.MaxReplicas)
		}
	} else {
		validatedSpec.MaxReplicas = math.MaxUint32
	}

	return &validatedSpec, nil
}
