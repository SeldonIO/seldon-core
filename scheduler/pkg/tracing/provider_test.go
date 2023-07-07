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

package tracing

import (
	"context"
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

func TestRecreateTracerProvider(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name   string
		config *TracingConfig
		err    bool
	}

	tests := []test{
		{
			name: "enabled",
			config: &TracingConfig{
				Disable:              false,
				OtelExporterEndpoint: "0.0.0.0:1234",
				Ratio:                "1",
			},
		},
		{
			name: "disabled",
			config: &TracingConfig{
				Disable: true,
			},
		},
		{
			name: "invalid ratio zero",
			config: &TracingConfig{
				Disable: false,
			},
			err: true,
		},
		{
			name: "invalid no otel endpoint",
			config: &TracingConfig{
				Disable: false,
				Ratio:   "10",
			},
			err: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := logrus.New()
			traceProvider, err := NewTraceProvider("test", nil, logger)
			g.Expect(err).To(BeNil())
			err = traceProvider.recreateTracerProvider(test.config)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				tracer := traceProvider.GetTraceProvider().Tracer("test")
				_, span := tracer.Start(context.TODO(), "test")
				span.End()
				traceProvider.Stop()
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name           string
		config         string
		expectedConfig *TracingConfig
		err            bool
	}

	tests := []test{
		{
			name:   "disabled",
			config: `{"disable":true}`,
			expectedConfig: &TracingConfig{
				Disable: true,
			},
		},
		{
			name:   "enabled",
			config: `{"disable":false, "otelExporterEndpoint":"0.0.0.0:1234","ratio":"0.5"}`,
			expectedConfig: &TracingConfig{
				Disable:              false,
				OtelExporterEndpoint: "0.0.0.0:1234",
				Ratio:                "0.5",
			},
		},
		{
			name:   "bad ratio",
			config: `{"disable":false, "otelExporterEndpoint":"0.0.0.0:1234","ratio":"foo"}`,
			err:    true,
		},
		{
			name:   "bad config",
			config: `{"foobar":true}`,
			err:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := logrus.New()
			path := fmt.Sprintf("%s/tracing-config.json", t.TempDir())
			err := os.WriteFile(path, []byte(test.config), 0644)
			g.Expect(err).To(BeNil())
			traceProvider, err := NewTraceProvider("test", &path, logger)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(traceProvider).ToNot(BeNil())
				g.Expect(traceProvider.config).To(Equal(test.expectedConfig))
			}
		})
	}
}

func TestRecreateTracerEarlyStop(t *testing.T) {
	logger := logrus.New()
	traceProvider, _ := NewTraceProvider("test", nil, logger)
	traceProvider.Stop()
}
