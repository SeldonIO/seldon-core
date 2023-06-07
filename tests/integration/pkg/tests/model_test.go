package tests

import (
	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/tests/integration/v2/pkg/resources"
	"testing"
	"time"
)

func TestSingleModelLoadInferUnload(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name         string
		modelPath    string
		inferRequest string
	}
	tests := []test{
		{
			name:         "sklearn - iris",
			modelPath:    "testdata/sklearn-iris.yaml",
			inferRequest: `{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}`,
		},
		{
			name:         "tensorflow - tfsimple",
			modelPath:    "testdata/tensorflow-tfsimple.yaml",
			inferRequest: `{"model_name":"tfsimple1","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}`,
		},
		{
			name:         "xgboost - income",
			modelPath:    "testdata/xgboost-income.yaml",
			inferRequest: `{ "parameters": {"content_type": "pd"}, "inputs": [{"name": "Age", "shape": [1, 1], "datatype": "INT64", "data": [47]},{"name": "Workclass", "shape": [1, 1], "datatype": "INT64", "data": [4]},{"name": "Education", "shape": [1, 1], "datatype": "INT64", "data": [1]},{"name": "Marital Status", "shape": [1, 1], "datatype": "INT64", "data": [1]},{"name": "Occupation", "shape": [1, 1], "datatype": "INT64", "data": [1]},{"name": "Relationship", "shape": [1, 1], "datatype": "INT64", "data": [3]},{"name": "Race", "shape": [1, 1], "datatype": "INT64", "data": [4]},{"name": "Sex", "shape": [1, 1], "datatype": "INT64", "data": [1]},{"name": "Capital Gain", "shape": [1, 1], "datatype": "INT64", "data": [0]},{"name": "Capital Loss", "shape": [1, 1], "datatype": "INT64", "data": [0]},{"name": "Hours per week", "shape": [1, 1], "datatype": "INT64", "data": [40]},{"name": "Country", "shape": [1, 1], "datatype": "INT64", "data": [9]}]}`,
		},
	}

	sapi, err := resources.NewSeldonBackendAPI()
	g.Expect(err).To(BeNil())
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// load
			err = sapi.Load(test.modelPath)
			g.Expect(err).To(BeNil())
			// wait ready
			await := func() bool {
				loaded, err := sapi.IsLoaded(test.modelPath)
				g.Expect(err).To(BeNil())
				t.Logf("Waiting for model %s:%v", test.modelPath, loaded)
				return loaded
			}
			g.Eventually(await).WithTimeout(time.Second * 60).WithPolling(time.Second).Should(BeTrue())
			// Infer grpc
			err = sapi.Infer(test.modelPath, test.inferRequest)
			g.Expect(err).To(BeNil())
			// Unload
			err = sapi.UnLoad(test.modelPath)
			g.Expect(err).To(BeNil())
		})
	}
}
