/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package health_probe_test

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"testing"
	"time"

	g "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"go.uber.org/mock/gomock"

	health_probe "github.com/seldonio/seldon-core/scheduler/v2/pkg/health-probe"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/health-probe/mocks"
)

func TestHTTPServer_Start(t *testing.T) {
	const (
		pathReadiness = "/ready"
		pathLiveness  = "/live"
		pathStartup   = "/startup"

		port = 8080
	)

	tests := []struct {
		name                     string
		readinessEnabled         bool
		livenessEnabled          bool
		startupEnabled           bool
		expectedSuccessEndpoints []string
		expected500Endpoints     []string
		setupMock                func(manager *mocks.MockManager)
		expect503                bool
	}{
		{
			name:                     "success - readiness only",
			readinessEnabled:         true,
			expectedSuccessEndpoints: []string{pathReadiness},
			setupMock: func(manager *mocks.MockManager) {
				manager.EXPECT().CheckReadiness().Return(nil)
			},
		},
		{
			name:                     "success - liveness only",
			livenessEnabled:          true,
			expectedSuccessEndpoints: []string{pathLiveness},
			setupMock: func(manager *mocks.MockManager) {
				manager.EXPECT().CheckLiveness().Return(nil)
			},
		},
		{
			name:                     "success - startup only",
			startupEnabled:           true,
			expectedSuccessEndpoints: []string{pathStartup},
			setupMock: func(manager *mocks.MockManager) {
				manager.EXPECT().CheckStartup().Return(nil)
			},
		},
		{
			name:                 "failure - readiness only",
			readinessEnabled:     true,
			expected500Endpoints: []string{pathReadiness},
			setupMock: func(manager *mocks.MockManager) {
				manager.EXPECT().CheckReadiness().Return(errors.New("some error"))
			},
		},
		{
			name:                 "failure - liveness only",
			livenessEnabled:      true,
			expected500Endpoints: []string{pathLiveness},
			setupMock: func(manager *mocks.MockManager) {
				manager.EXPECT().CheckLiveness().Return(errors.New("some error"))
			},
		},
		{
			name:                 "failure - startup only",
			startupEnabled:       true,
			expected500Endpoints: []string{pathStartup},
			setupMock: func(manager *mocks.MockManager) {
				manager.EXPECT().CheckStartup().Return(errors.New("some error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g.RegisterTestingT(t)
			ctrl := gomock.NewController(t)

			mockManager := mocks.NewMockManager(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockManager)
			}
			logger := logrus.New()
			server := health_probe.NewHTTPServer(port, mockManager, logger)

			mockManager.EXPECT().HasCallbacks(health_probe.ProbeReadiness).Return(tt.readinessEnabled)
			mockManager.EXPECT().HasCallbacks(health_probe.ProbeLiveness).Return(tt.livenessEnabled)
			mockManager.EXPECT().HasCallbacks(health_probe.ProbeStartUp).Return(tt.startupEnabled)

			go func() {
				if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					t.Errorf("HTTP server error: %v", err)
				}
			}()
			defer func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				err := server.Shutdown(ctx)
				g.Expect(err).To(g.BeNil())
			}()

			time.Sleep(100 * time.Millisecond)

			client := http.Client{}
			for _, endpoint := range tt.expectedSuccessEndpoints {
				resp, err := client.Get("http://localhost:" + strconv.Itoa(port) + endpoint)
				g.Expect(err).To(g.BeNil())
				g.Expect(resp.StatusCode).To(g.Equal(http.StatusOK))
			}

			expected404Endpoints := []string{}
			if !tt.readinessEnabled {
				expected404Endpoints = append(expected404Endpoints, pathReadiness)
			}
			if !tt.livenessEnabled {
				expected404Endpoints = append(expected404Endpoints, pathLiveness)
			}
			if !tt.startupEnabled {
				expected404Endpoints = append(expected404Endpoints, pathStartup)
			}

			for _, endpoint := range expected404Endpoints {
				resp, err := client.Get("http://localhost:" + strconv.Itoa(port) + endpoint)
				g.Expect(err).To(g.BeNil())
				g.Expect(resp.StatusCode).To(g.Equal(http.StatusNotFound))
			}

			for _, endpoint := range tt.expected500Endpoints {
				resp, err := client.Get("http://localhost:" + strconv.Itoa(port) + endpoint)
				g.Expect(err).To(g.BeNil())
				g.Expect(resp.StatusCode).To(g.Equal(http.StatusInternalServerError))
			}
		})
	}
}
