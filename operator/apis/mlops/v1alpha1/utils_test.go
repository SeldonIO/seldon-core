/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetValidatedScalingSpec(t *testing.T) {
	int32Ptr := func(i int32) *int32 {
		return &i
	}

	tests := []struct {
		name         string
		replicas     *int32
		minReplicas  *int32
		maxReplicas  *int32
		expectedSpec *ValidatedScalingSpec
		errMessage   string
	}{
		{
			name:        "success - all nil parameters",
			replicas:    nil,
			minReplicas: nil,
			maxReplicas: nil,
			expectedSpec: &ValidatedScalingSpec{
				Replicas:    1,
				MinReplicas: 0,
				MaxReplicas: 0,
			},
		},
		{
			name:        "success - valid positive values",
			replicas:    int32Ptr(5),
			minReplicas: int32Ptr(2),
			maxReplicas: int32Ptr(10),
			expectedSpec: &ValidatedScalingSpec{
				Replicas:    5,
				MinReplicas: 2,
				MaxReplicas: 10,
			},
		},
		{
			name:        "error - replicas less than minReplicas",
			replicas:    int32Ptr(2),
			minReplicas: int32Ptr(5),
			maxReplicas: int32Ptr(10),
			errMessage:  "number of replicas 2 must be >= min replicas  5",
		},
		{
			name:        "error - replicas greater than maxReplicas",
			replicas:    int32Ptr(15),
			minReplicas: int32Ptr(2),
			maxReplicas: int32Ptr(10),
			errMessage:  "number of replicas 15 must be <= min replicas  10", // Note: This is the actual error message with the bug
		},
		{
			name:     "success - zero replicas",
			replicas: int32Ptr(0),
			expectedSpec: &ValidatedScalingSpec{
				Replicas:    0,
				MinReplicas: 0,
				MaxReplicas: 0,
			},
		},
		{
			name:     "success - negative replicas defaults to 1",
			replicas: int32Ptr(-5),
			expectedSpec: &ValidatedScalingSpec{
				Replicas:    1,
				MinReplicas: 0,
				MaxReplicas: 0,
			},
		},
		{
			name:        "success - zero minReplicas defaults to 0",
			replicas:    int32Ptr(5),
			minReplicas: int32Ptr(0),
			maxReplicas: int32Ptr(10),
			expectedSpec: &ValidatedScalingSpec{
				Replicas:    5,
				MinReplicas: 0,
				MaxReplicas: 10,
			},
		},
		{
			name:        "success - negative minReplicas defaults to 0",
			replicas:    int32Ptr(5),
			minReplicas: int32Ptr(-2),
			maxReplicas: int32Ptr(10),
			expectedSpec: &ValidatedScalingSpec{
				Replicas:    5,
				MinReplicas: 0,
				MaxReplicas: 10,
			},
		},
		{
			name:        "success - zero maxReplicas defaults to 0",
			replicas:    int32Ptr(5),
			minReplicas: int32Ptr(2),
			maxReplicas: int32Ptr(0),
			expectedSpec: &ValidatedScalingSpec{
				Replicas:    5,
				MinReplicas: 2,
				MaxReplicas: 0,
			},
		},
		{
			name:        "success - negative maxReplicas defaults to 0",
			replicas:    int32Ptr(5),
			minReplicas: int32Ptr(2),
			maxReplicas: int32Ptr(-3),
			expectedSpec: &ValidatedScalingSpec{
				Replicas:    5,
				MinReplicas: 2,
				MaxReplicas: 0,
			},
		},
		{
			name:        "success - replicas equals minReplicas",
			replicas:    int32Ptr(5),
			minReplicas: int32Ptr(5),
			maxReplicas: int32Ptr(10),
			expectedSpec: &ValidatedScalingSpec{
				Replicas:    5,
				MinReplicas: 5,
				MaxReplicas: 10,
			},
		},
		{
			name:        "success - replicas equals maxReplicas",
			replicas:    int32Ptr(10),
			minReplicas: int32Ptr(5),
			maxReplicas: int32Ptr(10),
			expectedSpec: &ValidatedScalingSpec{
				Replicas:    10,
				MinReplicas: 5,
				MaxReplicas: 10,
			},
		},
		{
			name:        "error - default replicas with valid minReplicas",
			minReplicas: int32Ptr(3),
			errMessage:  "number of replicas 1 must be >= min replicas  3",
		},
		{
			name:     "success - only replicas specified",
			replicas: int32Ptr(7),
			expectedSpec: &ValidatedScalingSpec{
				Replicas:    7,
				MinReplicas: 0,
				MaxReplicas: 0,
			},
		},
		{
			name:        "error - only minReplicas specified",
			minReplicas: int32Ptr(2),
			errMessage:  "number of replicas 1 must be >= min replicas  2",
		},
		{
			name:        "success - only maxReplicas specified with default replicas",
			maxReplicas: int32Ptr(5),
			expectedSpec: &ValidatedScalingSpec{
				Replicas:    1,
				MinReplicas: 0,
				MaxReplicas: 5,
			},
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := GetValidatedScalingSpec(tt.replicas, tt.minReplicas, tt.maxReplicas)
			if tt.errMessage != "" {
				require.Error(t, err)
				require.Nil(t, result)
				require.ErrorContains(t, err, tt.errMessage)
				return
			}

			require.Nil(t, err)
			require.Equal(t, tt.expectedSpec, result)
		})
	}
}
