/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
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
