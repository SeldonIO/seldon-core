package tests

import (
	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/tests/integration/v2/pkg/resources"
	"os"
	"testing"
	"time"
)

func TestSingleModelLoadInferUnload(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name              string
		modelPath         string
		inferRequestPath  string
		inferResponsePath string
	}
	tests := []test{
		{
			name:              "sklearn - iris",
			modelPath:         "testdata/sklearn-iris.yaml",
			inferRequestPath:  `testdata/sklearn-iris-request.json`,
			inferResponsePath: `testdata/sklearn-iris-response.json`,
		},
		{
			name:              "tensorflow - tfsimple",
			modelPath:         "testdata/tensorflow-tfsimple.yaml",
			inferRequestPath:  `testdata/tensorflow-tfsimple-request.json`,
			inferResponsePath: `testdata/tensorflow-tfsimple-response.json`,
		},
		{
			name:              "xgboost - income",
			modelPath:         "testdata/xgboost-income.yaml",
			inferRequestPath:  `testdata/xgboost-income-request.json`,
			inferResponsePath: `testdata/xgboost-income-response.json`,
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
			res, err := sapi.Infer(test.modelPath, test.inferRequestPath)
			g.Expect(err).To(BeNil())
			expectedResponse, err := os.ReadFile(test.inferResponsePath)
			g.Expect(err).To(BeNil())
			g.Expect(expectedResponse).To(MatchJSON(res))
			// Unload
			err = sapi.UnLoad(test.modelPath)
			g.Expect(err).To(BeNil())
		})
	}
}
