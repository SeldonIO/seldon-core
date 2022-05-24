package tracing

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type TracerProvider struct {
	TraceProvider *trace.TracerProvider
}

func NewTracer(serviceName string) (*TracerProvider, error) {
	tp, err := initTracer(serviceName)
	if err != nil {
		return nil, err
	}
	return &TracerProvider{
		TraceProvider: tp,
	}, nil
}

func (t *TracerProvider) Stop() {
	cxt, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := t.TraceProvider.Shutdown(cxt); err != nil {
		otel.Handle(err)
	}
}

func initTracer(serviceName string) (*trace.TracerProvider, error) {

	otelAgentAddr, ok := os.LookupEnv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if !ok {
		otelAgentAddr = "0.0.0.0:4317"
	}

	traceClient := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(otelAgentAddr),
		// For testing we can't block and not sure is needed in production in case not tracing available
		//otlptracegrpc.WithDialOption(grpc.WithBlock()),
	)
	traceExp, err := otlptrace.New(context.Background(), traceClient)
	if err != nil {
		return nil, err
	}
	bsp := trace.NewBatchSpanProcessor(traceExp)

	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String(serviceName))),
		trace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp, nil
}
