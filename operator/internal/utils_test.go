/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package internal

import (
	"math"
	"testing"

	"github.com/gotidy/ptr"
	. "github.com/onsi/gomega"
)

func TestGetValidatedScalingSpec(t *testing.T) {
	type test struct {
		name        string
		replicas    *int32
		minReplicas *int32
		maxReplicas *int32
		expected    *ValidatedScalingSpec
		wantErr     string
	}

	g := NewGomegaWithT(t)

	tests := []test{
		{
			name:        "success - replicas is higher than min replicas and lower than max replicas",
			replicas:    ptr.Int32(2),
			minReplicas: ptr.Int32(1),
			maxReplicas: ptr.Int32(3),
			expected: &ValidatedScalingSpec{
				Replicas:    2,
				MinReplicas: 1,
				MaxReplicas: 3,
			},
			wantErr: "",
		},
		{
			name:        "error - replicas is less than min replicas",
			replicas:    ptr.Int32(1),
			minReplicas: ptr.Int32(2),
			maxReplicas: ptr.Int32(4),
			expected:    nil,
			wantErr:     "scaling spec is invalid: number of replicas 1 must be >= min replicas 2",
		},
		{
			name:        "error - replicas is bigger than max replicas",
			replicas:    ptr.Int32(5),
			minReplicas: ptr.Int32(1),
			maxReplicas: ptr.Int32(4),
			expected:    nil,
			wantErr:     "scaling spec is invalid: number of replicas 5 must be <= max replicas 4",
		},
		{
			name:        "error - replica is less than min replicas",
			replicas:    ptr.Int32(0),
			minReplicas: ptr.Int32(1),
			maxReplicas: nil,
			expected:    nil,
			wantErr:     "scaling spec is invalid: number of replicas 0 must be >= min replicas 1",
		},
		{
			name:        "error - min replica is bigger than max replicas",
			replicas:    ptr.Int32(6),
			minReplicas: ptr.Int32(6),
			maxReplicas: ptr.Int32(4),
			expected:    nil,
			wantErr:     "scaling spec is invalid: min number of replicas 6 must be <= max number of replicas 4",
		},
		{
			name:        "success - replicas stays at 0 when min replicas is 0 and max replicas is 4",
			replicas:    ptr.Int32(0),
			minReplicas: ptr.Int32(0),
			maxReplicas: ptr.Int32(4),
			expected: &ValidatedScalingSpec{
				Replicas:    0,
				MinReplicas: 0,
				MaxReplicas: 4,
			},
			wantErr: "",
		},
		{
			name:     "success - replicas = 0",
			replicas: ptr.Int32(0),
			expected: &ValidatedScalingSpec{
				Replicas:    0,
				MinReplicas: 0,
				MaxReplicas: math.MaxUint32,
			},
			wantErr: "",
		},
		{
			name:        "success - all replica params are the same",
			replicas:    ptr.Int32(4),
			minReplicas: ptr.Int32(4),
			maxReplicas: ptr.Int32(4),
			expected: &ValidatedScalingSpec{
				Replicas:    4,
				MinReplicas: 4,
				MaxReplicas: 4,
			},
			wantErr: "",
		},
		{
			name:        "success - min and max replicas default to right params when only replicas is set",
			replicas:    ptr.Int32(2),
			minReplicas: nil,
			maxReplicas: nil,
			expected: &ValidatedScalingSpec{
				Replicas:    2,
				MinReplicas: 0,
				MaxReplicas: math.MaxUint32,
			},
			wantErr: "",
		},
		{
			name:        "success - unset replica params defaults to 1",
			replicas:    nil,
			minReplicas: nil,
			maxReplicas: nil,
			expected: &ValidatedScalingSpec{
				Replicas:    1,
				MinReplicas: 0,
				MaxReplicas: math.MaxUint32,
			},
			wantErr: "",
		},
		{
			name:        "success - unset replica params defaults to value of min replicas",
			replicas:    nil,
			minReplicas: ptr.Int32(2),
			maxReplicas: nil,
			expected: &ValidatedScalingSpec{
				Replicas:    2,
				MinReplicas: 2,
				MaxReplicas: math.MaxUint32,
			},
			wantErr: "",
		},
		{
			name:        "success - unset replica params defaults to 1 when min replicas is set to 0",
			replicas:    nil,
			minReplicas: ptr.Int32(0),
			maxReplicas: nil,
			expected: &ValidatedScalingSpec{
				Replicas:    1,
				MinReplicas: 0,
				MaxReplicas: math.MaxUint32,
			},
			wantErr: "",
		},
		{
			name:        "success - unset minReplicas defaults to 0, unset replicas defaults to 1 when max replicas is set",
			replicas:    nil,
			minReplicas: nil,
			maxReplicas: ptr.Int32(2),
			expected: &ValidatedScalingSpec{
				Replicas:    1,
				MinReplicas: 0,
				MaxReplicas: 2,
			},
			wantErr: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scalingSpec, err := GetValidatedScalingSpec(test.replicas, test.minReplicas, test.maxReplicas)

			if test.wantErr != "" {
				if err == nil {
					t.Errorf("expected error: %v, got nil", test.wantErr)
					return
				}
				if err.Error() != test.wantErr {
					t.Errorf("expected error: %q, got: %q", test.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			g.Expect(scalingSpec).To(Equal(test.expected))
		})
	}
}
