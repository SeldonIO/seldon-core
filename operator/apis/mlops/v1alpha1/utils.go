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
	spec, err := validatedScalingSpec(replicas, minReplicas, maxReplicas)
	if err != nil {
		return nil, fmt.Errorf("failed scaling spec check: %s", err)
	}
	return spec, nil
}

func validatedScalingSpec(replicas *int32, minReplicas *int32, maxReplicas *int32) (*ValidatedScalingSpec, error) {
	var validatedSpec ValidatedScalingSpec

	if replicas != nil {
		if *replicas < 0 {
			return nil, fmt.Errorf("replicas %d cannot be negative", *replicas)
		}
		validatedSpec.Replicas = uint32(*replicas)
	} else if minReplicas != nil {
		validatedSpec.Replicas = uint32(*minReplicas)
	} else {
		// default to 1 if replicas and minimum not set
		validatedSpec.Replicas = 1
	}

	if minReplicas != nil {
		if *minReplicas < 0 {
			return nil, fmt.Errorf("min replicas %d cannot be negative", *minReplicas)
		}
		validatedSpec.MinReplicas = uint32(*minReplicas)
		if validatedSpec.Replicas < validatedSpec.MinReplicas {
			return nil, fmt.Errorf("number of replicas %d must be >= min replicas %d",
				validatedSpec.Replicas, validatedSpec.MinReplicas)
		}
	}

	if maxReplicas != nil {
		if *maxReplicas < 0 {
			return nil, fmt.Errorf("max replicas %d cannot be negative", *maxReplicas)
		}
		validatedSpec.MaxReplicas = uint32(*maxReplicas)
		if validatedSpec.Replicas > validatedSpec.MaxReplicas {
			return nil, fmt.Errorf("number of replicas %d must be <= max replicas %d",
				validatedSpec.Replicas,
				validatedSpec.MaxReplicas)
		}
	}

	if minReplicas != nil && maxReplicas != nil && *minReplicas > *maxReplicas {
		return nil, fmt.Errorf("min replicas %d cannot be greater than max replicas %d",
			*minReplicas, *maxReplicas)
	}

	return &validatedSpec, nil
}
