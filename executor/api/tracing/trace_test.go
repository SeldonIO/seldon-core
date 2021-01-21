package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"os"
	"testing"
)

var (
	nonRootNonSampledHeader = opentracing.TextMapCarrier{
		"x-b3-traceid":      "1",
		"x-b3-spanid":       "2",
		"x-b3-parentspanid": "1",
		"x-b3-sampled":      "0",
	}
)

func newSpanContext(traceID, spanID, parentID uint64, sampled bool, baggage map[string]string) jaeger.SpanContext {
	return jaeger.NewSpanContext(
		jaeger.TraceID{Low: traceID},
		jaeger.SpanID(spanID),
		jaeger.SpanID(parentID),
		sampled,
		baggage,
	)
}

func TestInitTracing(t *testing.T) {
	t.Run("test jaeger tracing", func(t *testing.T) {
		os.Setenv("JAEGER_TRACE_PROPAGATION_TYPE", "b3")
		closer, err := InitTracing()
		if err != nil {
			t.Error("init trace err", err.Error())
		}
		defer closer.Close()
		if !opentracing.IsGlobalTracerRegistered() {
			t.Error("trace is not registered")
		}
		tracer := opentracing.GlobalTracer()
		_, err = tracer.Extract(opentracing.HTTPHeaders, nonRootNonSampledHeader)
		if err != nil {
			t.Error("trace extract error", err.Error())
		}
	})

	t.Run("test dd tracer", func(t *testing.T) {
		os.Setenv(datadogEnabled, "TRUE")
		os.Setenv("DD_PROPAGATION_STYLE_EXTRACT", "B3")
		closer, err := InitTracing()
		if err != nil {
			t.Error("init trace err", err.Error())
		}
		defer closer.Close()
		if !opentracing.IsGlobalTracerRegistered() {
			t.Error("trace is not registered")
		}
		tracer := opentracing.GlobalTracer()
		_, err = tracer.Extract(opentracing.HTTPHeaders, nonRootNonSampledHeader)
		if err != nil {
			t.Error("trace extract error", err.Error())
		}
	})
}
