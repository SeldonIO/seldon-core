/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
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
