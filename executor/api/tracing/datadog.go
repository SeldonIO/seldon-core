package tracing

import (
	"fmt"
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
// correct environment variables are present. See https://docs.datadoghq.com/tracing/setup/go/
// for all Environment variables.
// Additional ones sepcific to Seldon:
// DD_SAMPLE_RATE --> 0-1, rate of sampling
// DD_ENABLED --> 0,1
func initDatadogTracer() (io.Closer, error) {

	// TODO: remove me
	fmt.Println("dd tracing start")

	serviceName := os.Getenv("DD_SERVICE")
	if serviceName == "" {
		serviceName = "executor"
	}

	fmt.Println("svc name: ", serviceName)

	samplingRate, err := strconv.ParseFloat(os.Getenv(datadogSamplingRate), 64)
	if err != nil {
		samplingRate = 1.0
	}

	fmt.Println("sample rate: ", samplingRate)

	t := opentracer.New(
		tracer.WithSampler(tracer.NewRateSampler(samplingRate)),
	)

	opentracing.SetGlobalTracer(t)

	return new(datadogTracer), nil
}
