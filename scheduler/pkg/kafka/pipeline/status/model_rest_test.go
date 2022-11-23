/*
Copyright 2022 Seldon Technologies Ltd.

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

package status

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

func TestCheckModelReady(t *testing.T) {
	g := NewGomegaWithT(t)

	envoyHost := "envoy"
	envoyPort := 1234
	type test struct {
		name          string
		modelName     string
		status        int
		expectedReady bool
	}

	tests := []test{
		{
			name:          "model ready",
			modelName:     "test",
			status:        http.StatusOK,
			expectedReady: true,
		},
		{
			name:          "model not ready",
			modelName:     "test",
			status:        http.StatusBadRequest,
			expectedReady: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder(http.MethodGet, fmt.Sprintf("http://%s:%d/v2/models/%s/ready", envoyHost, envoyPort, test.modelName),
				httpmock.NewStringResponder(test.status, `{}`))
			mrc, err := NewModelRestStatusCaller(logrus.New(), envoyHost, envoyPort)
			g.Expect(err).To(BeNil())
			ready, err := mrc.CheckModelReady(context.TODO(), test.modelName, "1")
			g.Expect(err).To(BeNil())
			g.Expect(ready).To(Equal(test.expectedReady))
		})
	}
}
