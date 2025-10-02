/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package gateway

import (
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	. "github.com/onsi/gomega"
)

func TestGetIntConfigMapValue(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		configMap    kafka.ConfigMap
		key          string
		defaultValue int
		wantValue    int
		wantError    bool
	}

	tests := []test{
		{
			name:         "success: string",
			configMap:    kafka.ConfigMap{"replicationFactor": "5"},
			key:          replicationFactorKey,
			defaultValue: 0,
			wantValue:    5,
			wantError:    false,
		},
		{
			name:         "fail: negative string",
			configMap:    kafka.ConfigMap{"replicationFactor": "-5"},
			key:          replicationFactorKey,
			defaultValue: 0,
			wantValue:    0,
			wantError:    true,
		},
		{
			name:         "fail: float string value",
			configMap:    kafka.ConfigMap{"replicationFactor": "5.0"},
			key:          replicationFactorKey,
			defaultValue: 0,
			wantValue:    0,
			wantError:    true,
		},
		{
			name:         "fail: string value",
			configMap:    kafka.ConfigMap{"replicationFactor": "---"},
			key:          replicationFactorKey,
			defaultValue: 0,
			wantValue:    0,
			wantError:    true,
		},
		{
			name:         "success: integer",
			configMap:    kafka.ConfigMap{"replicationFactor": 5},
			key:          replicationFactorKey,
			defaultValue: 0,
			wantValue:    5,
			wantError:    false,
		},
		{
			name:         "fail: negative integer",
			configMap:    kafka.ConfigMap{"replicationFactor": -5},
			key:          replicationFactorKey,
			defaultValue: 0,
			wantValue:    0,
			wantError:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotInt, err := GetIntConfigMapValue(test.configMap, test.key, test.defaultValue)
			if test.wantError {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(gotInt).To(Equal(test.wantValue))
			}
		})

	}

}
