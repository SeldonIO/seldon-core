package tracing

import (
	"io"
	"os"
	"strconv"

	"github.com/opentracing/opentracing-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/zipkin"
)

func InitTracing() (io.Closer, error) {
	//Initialise tracing

	if ddEnabled, _ := strconv.ParseBool(os.Getenv(datadogEnabled)); ddEnabled {
		tracer, err := initDatadogTracer()
		if err != nil {
			return nil, err
		}
		return tracer, nil
	}

	cfg, err := jaegercfg.FromEnv()
	if err != nil {
		// parsing errors might happen here, such as when we get a string where we expect a number
		return nil, err
	}

	if cfg.ServiceName == "" {
		cfg.ServiceName = "executor"
	}

	propagation := os.Getenv("JAEGER_TRACE_PROPAGATION_TYPE")
	closer, err := newTraceByPropagation(propagation, cfg)
	if err != nil {
		return nil, err
	}
	return closer, nil
}

func newTraceByPropagation(propagation string, cfg *jaegercfg.Configuration) (closer io.Closer, err error) {
	switch propagation {
	case "b3":
		return newB3Tracer(cfg)
	}
	return newDefaultTracer(cfg)
}

func newB3Tracer(cfg *jaegercfg.Configuration) (io.Closer, error) {
	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()
	closer, err := cfg.InitGlobalTracer(cfg.ServiceName,
		jaegercfg.Injector(opentracing.HTTPHeaders, zipkinPropagator),
		jaegercfg.Extractor(opentracing.HTTPHeaders, zipkinPropagator),
		jaegercfg.ZipkinSharedRPCSpan(true))

	return closer, err
}
func newDefaultTracer(cfg *jaegercfg.Configuration) (io.Closer, error) {
	closer, err := cfg.InitGlobalTracer(cfg.ServiceName)
	return closer, err
}
