package tracing

import (
	"io"
	"os"
	"strconv"

	"github.com/opentracing/opentracing-go"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentracer"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const (
	datadogEnabled      = "DD_ENABLED"
	datadogSamplingRate = "DD_SAMPLING_RATE"
)

// datadogTracer satisfies the io.Closer interface
type datadogTracer struct {
}

func (d *datadogTracer) Close() error {
	tracer.Stop()
	return nil
}

// TODO: docs/comments
// initDatadogTracer attempts to create a tracer for using DataDog and statsd if the
// correct environment variables are present. See https://pkg.go.dev/gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer?tab=doc#StartOption
// for all Environment variables. Here are the relevant ones:
// DD_AGENT_HOST --> default localhost
// DD_DOGSTATSD_PORT --> default 8126
// DD_ENV
// DD_SERVICE
// DD_VERSION
// DD_TAGS --> "k:v,k2:v2"
// DD_TRACE_ANALYTICS_ENABLED
// DD_TRACE_REPORT_HOSTNAME
// DD_TRACE_STARTUP_LOGS
// DD_RUNTIME_METRICS_ENABLED
// DD_TRACE_DEBUG
// TODO:
// DD_SAMPLING_RATE
// DD_ENABLED
//
// DD_PROPAGATION_STYLE_INJECT=Datadog,B3
// DD_PROPAGATION_STYLE_EXTRACT=Datadog,B3
func initDatadogTracer() (io.Closer, error) {

	serviceName := os.Getenv("DD_SERVICE")
	if serviceName == "" {
		serviceName = "executor"
	}

	samplingRate, err := strconv.ParseFloat(os.Getenv(datadogSamplingRate), 64)
	if err != nil {
		samplingRate = 1.0
	}

	opentracing.SetGlobalTracer(
		opentracer.New(
			tracer.WithSampler(tracer.NewRateSampler(samplingRate)),
		))

	return new(datadogTracer), nil
}
