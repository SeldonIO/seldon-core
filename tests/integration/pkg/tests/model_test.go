//go:build integration

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

package tests

import (
	"os"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/seldonio/seldon-core/tests/integration/v2/pkg/resources"
)

func TestModelInference(t *testing.T) {
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
			err = sapi.Unload(test.modelPath)
			g.Expect(err).To(BeNil())
		})
	}
}
