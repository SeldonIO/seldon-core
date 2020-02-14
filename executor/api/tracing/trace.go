package tracing

import (
	"github.com/opentracing/opentracing-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/zipkin"
	"io"
)

func NewTraceByPropagation(propagation string, cfg *jaegercfg.Configuration) (closer io.Closer, err error) {
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
