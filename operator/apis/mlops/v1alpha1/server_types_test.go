/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package v1alpha1

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	v1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/ptr"
)

/* WARNING: Read this first if test below fails (either at compile-time or while running the
* test):
*
* The test below aims to ensure that the fields used in kubebuilder:printcolumn comments in
* server_types.go match the structure and condition types in the Server CRD.
*
* If the test fails, it means that the CRD was updated without updating the kubebuilder:
* printcolumn comments.
*
* Rather than fixing the test directly, FIRST UPDATE the kubebuilder:printcolumn comments,
* THEN update the test to also match the new values.
*
 */
func TestServerStatusPrintColumns(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                          string
		server                        *Server
		expectedJsonSerializationKeys []string
	}
	json_fail_reason := "Json serialization of Server.Status does not contain path \"%s\", used in kubebuilder.printcolumn comments. " +
		"Update kubebuilder:printcolumn comments to match the new serialization keys."

	tests := []test{
		{
			name: "server replicas",
			server: &Server{
				Status: ServerStatus{
					Replicas: 1,
				},
			},
			expectedJsonSerializationKeys: []string{"status.replicas"},
		},
		{
			name: "server loaded models",
			server: &Server{
				Status: ServerStatus{
					LoadedModelReplicas: 1,
				},
			},
			expectedJsonSerializationKeys: []string{"status.loadedModels"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(test.server)
			g.Expect(err).To(BeNil())
			for _, key := range test.expectedJsonSerializationKeys {
				g.Expect(gjson.GetBytes(jsonBytes, key).Exists()).To(BeTrueBecause(json_fail_reason, key))
			}
		})
	}
}

func TestServerStatusSetCondition(t *testing.T) {
	type args struct {
		condition *apis.Condition
	}
	tests := []struct {
		name string
		args args
		want *v1.ConditionStatus
	}{
		{
			name: "should not panic if condition is nil",
			args: args{
				condition: nil,
			},
			want: nil,
		},
		{
			name: "ConditionUnknown",
			args: args{
				condition: &apis.Condition{
					Status: "Unknown",
				},
			},
			want: (*v1.ConditionStatus)(ptr.String("Unknown")),
		},
		{
			name: "ConditionTrue",
			args: args{
				condition: &apis.Condition{
					Status: "True",
				},
			},
			want: (*v1.ConditionStatus)(ptr.String("True")),
		},
		{
			name: "ConditionFalse",
			args: args{
				condition: &apis.Condition{
					Status: "False",
				},
			},
			want: (*v1.ConditionStatus)(ptr.String("False")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &ServerStatus{}
			ss.SetCondition(tt.args.condition)
			if tt.want == nil {
				if ss.GetCondition("") != nil {
					t.Errorf("want %v : got %v", tt.want, ss.GetCondition(""))
				}
			}
			if tt.want != nil {
				got := ss.GetCondition("").Status
				if *tt.want != got {
					t.Errorf("want %v : got %v", *tt.want, got)
				}
			}
		})
	}
}
