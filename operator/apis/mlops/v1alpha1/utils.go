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

// TODO consider putting these rules via CEL into the CR definition for Server, this will allow customers earlier feedback
//
//	their CRs are invalid upon trying to apply them, instead of checking controller logs after applying.
func GetValidatedScalingSpec(replicas *int32, minReplicas *int32, maxReplicas *int32) (*ValidatedScalingSpec, error) {
	validatedSpec := ValidatedScalingSpec{
		Replicas:    1,
		MinReplicas: 0,
		MaxReplicas: math.MaxUint32,
	}

	if replicas == nil && minReplicas == nil && maxReplicas == nil {
		return &validatedSpec, nil
	}

	if replicas != nil && *replicas >= 0 {
		validatedSpec.Replicas = uint32(*replicas)
	} else if replicas != nil && *replicas == 0 && minReplicas != nil && *minReplicas == 0 {
		validatedSpec.Replicas = uint32(0)
	} else if minReplicas != nil && *minReplicas > 0 {
		if replicas != nil && *replicas < *minReplicas {
			return nil, fmt.Errorf("number of replicas %d cannot be less than minimum replica %d", *replicas, *minReplicas)
		}
		// set replicas to the min replicas when replicas is not set explicitly
		validatedSpec.Replicas = uint32(*minReplicas)
	}

	if minReplicas != nil && *minReplicas > 0 {
		validatedSpec.MinReplicas = uint32(*minReplicas)
		if validatedSpec.Replicas < validatedSpec.MinReplicas {
			return nil, fmt.Errorf("number of replicas %d must be >= min replicas %d", validatedSpec.Replicas, validatedSpec.MinReplicas)
		}
	}

	if maxReplicas != nil && *maxReplicas > 0 {
		validatedSpec.MaxReplicas = uint32(*maxReplicas)
		if validatedSpec.MinReplicas > validatedSpec.MaxReplicas {
			return nil, fmt.Errorf("min number of replicas %d must be <= max number of replicas %d", validatedSpec.MinReplicas, validatedSpec.MaxReplicas)
		}
		if validatedSpec.Replicas > validatedSpec.MaxReplicas {
			return nil, fmt.Errorf("number of replicas %d must be <= max replicas %d", validatedSpec.Replicas, validatedSpec.MaxReplicas)
		}
	}

	return &validatedSpec, nil
}
