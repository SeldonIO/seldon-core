/*
Copyright 2023 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resources

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
)

func TestInferRequestProtocol(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		request      string
		expectedType cli.InferProtocol
		error        bool
	}

	tests := []test{
		{
			name:         "rest",
			request:      `{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}`,
			expectedType: cli.InferGrpc,
		},
		{
			name:         "grpc",
			request:      `{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}`,
			expectedType: cli.InferRest,
		},
		{
			name:         "unknown",
			request:      `{"inputs": [{"foo":"bar"]}]}`,
			expectedType: cli.InferUnknown,
			error:        true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ty, err := getInferRequestProtocol([]byte(test.request))
			if test.error {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
			}
			g.Expect(ty).To(Equal(test.expectedType))
		})
	}
}
