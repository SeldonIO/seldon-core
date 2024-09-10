/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package v1alpha1

import "fmt"

type ValidatedScalingSpec struct {
	Replicas    uint32
	MinReplicas uint32
	MaxReplicas uint32
}

func GetValidatedScalingSpec(replicas *int32, minReplicas *int32, maxReplicas *int32) (*ValidatedScalingSpec, error) {
	var validatedSpec ValidatedScalingSpec

	if replicas != nil && *replicas > 0 {
		validatedSpec.Replicas = uint32(*replicas)
	} else {
		if minReplicas != nil && *minReplicas > 0 {
			// set replicas to the min replicas when replicas is not set explicitly
			validatedSpec.Replicas = uint32(*minReplicas)
		} else {
			validatedSpec.Replicas = 1
		}
	}

	if minReplicas != nil && *minReplicas > 0 {
		validatedSpec.MinReplicas = uint32(*minReplicas)
		if validatedSpec.Replicas < validatedSpec.MinReplicas {
			return nil, fmt.Errorf("number of replicas %d must be >= min replicas  %d", validatedSpec.Replicas, validatedSpec.MinReplicas)
		}
	} else {
		validatedSpec.MinReplicas = 0
	}

	if maxReplicas != nil && *maxReplicas > 0 {
		validatedSpec.MaxReplicas = uint32(*maxReplicas)
		if validatedSpec.Replicas > validatedSpec.MaxReplicas {
			return nil, fmt.Errorf("number of replicas %d must be <= min replicas  %d", validatedSpec.Replicas, validatedSpec.MaxReplicas)
		}
	} else {
		validatedSpec.MaxReplicas = 0
	}

	return &validatedSpec, nil
}
