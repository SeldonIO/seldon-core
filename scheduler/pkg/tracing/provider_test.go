package tracing

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	. "github.com/onsi/gomega"
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
				Enable:               true,
				OtelExporterEndpoint: "0.0.0.0:1234",
				Ratio:                1,
			},
		},
		{
			name: "disabled",
			config: &TracingConfig{
				Enable: false,
			},
		},
		{
			name: "invalid ratio zero",
			config: &TracingConfig{
				Enable: true,
			},
			err: true,
		},
		{
			name: "invalid no otel endpoint",
			config: &TracingConfig{
				Enable: true,
				Ratio:  10,
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
			config: `{"enable":false}`,
			expectedConfig: &TracingConfig{
				Enable: false,
			},
		},
		{
			name:   "enabled",
			config: `{"enable":true, "otelExporterEndpoint":"0.0.0.0:1234","ratio":0.5}`,
			expectedConfig: &TracingConfig{
				Enable:               true,
				OtelExporterEndpoint: "0.0.0.0:1234",
				Ratio:                0.5,
			},
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
