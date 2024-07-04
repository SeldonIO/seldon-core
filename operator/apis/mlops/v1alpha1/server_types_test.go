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
	"encoding/json"

	"github.com/tidwall/gjson"
	. "github.com/onsi/gomega"
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
		name    string
		server   *Server
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
