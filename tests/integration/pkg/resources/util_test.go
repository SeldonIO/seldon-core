package resources

import (
	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
	"testing"
)

func TestSingleModelLoadInferUnload(t *testing.T) {
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
			ty, err := getInferRequestProtocol(test.request)
			if test.error {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
			}
			g.Expect(ty).To(Equal(test.expectedType))
		})
	}
}
