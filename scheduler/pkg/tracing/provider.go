/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package tracing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	trace2 "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type TracerProvider struct {
	serviceName   string
	traceProvider trace2.TracerProvider
	mu            sync.RWMutex
	config        *TracingConfig
	logger        logrus.FieldLogger
}

type TracingConfig struct {
	Disable              bool   `json:"disable"`
	OtelExporterEndpoint string `json:"otelExporterEndpoint"`
	OtelExporterProtocol string `json:"otelExporterProtocol"`
	Ratio                string `json:"Ratio"`
}

func NewTraceProvider(serviceName string, configPath *string, logger logrus.FieldLogger) (*TracerProvider, error) {
	t := &TracerProvider{
		serviceName: serviceName,
		logger:      logger.WithField("source", "TraceProvider"),
	}
	err := t.loadConfig(configPath)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (t *TracerProvider) GetTraceProvider() trace2.TracerProvider {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.traceProvider
}

func (t *TracerProvider) Stop() {
	logger := t.logger.WithField("func", "Stop")
	t.mu.Lock()
	defer t.mu.Unlock()
	switch v := t.traceProvider.(type) {
	case *trace.TracerProvider:
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := v.Shutdown(ctx); err != nil {
			logger.WithError(err).Warn("failed to shutdown otel provider")
		}
	}
}

func (t *TracerProvider) loadConfig(path *string) error {
	logger := t.logger.WithField("func", "loadConfig")
	var config *TracingConfig
	if path == nil || *path == "" {
		config = &TracingConfig{
			Disable: true,
		}
		logger.Info("No tracing path provided so setting NOOP TraceProvider")
	} else {
		data, err := os.ReadFile(*path)
		if err != nil {
			return err
		}
		logger.WithField("config", string(data)).Infof("Loading tracing config")
		tc := TracingConfig{}
		d := json.NewDecoder(bytes.NewReader(data))
		d.DisallowUnknownFields() // So we fail if not exactly as required in schema
		err = d.Decode(&tc)
		if err != nil {
			return err
		}
		if !tc.Disable { // Only check ratio is valid if enabled
			// Check ratio
			_, err = strconv.ParseFloat(tc.Ratio, 64)
			if err != nil {
				return err
			}
		}
		config = &tc
	}
	return t.recreateTracerProvider(config)
}

func (t *TracerProvider) recreateTracerProvider(config *TracingConfig) error {
	logger := t.logger.WithField("func", "recreateTracerProvider")
	// add further check for config semantic validity
	if !config.Disable {
		if config.OtelExporterEndpoint == "" {
			return fmt.Errorf("Trace enabled but Otel endpoint empty")
		}
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.config = config
	if !t.config.Disable {

		traceClient := otlptracegrpc.NewClient(
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(t.config.OtelExporterEndpoint),
		)
		traceExp, err := otlptrace.New(context.Background(), traceClient)
		if err != nil {
			return err
		}
		bsp := trace.NewBatchSpanProcessor(traceExp)

		ratio := 1.0
		if t.config.Ratio != "" {
			ratioParsed, err := strconv.ParseFloat(t.config.Ratio, 64)
			if err != nil {
				logger.WithError(err).Errorf("Failed to parse tracing ratio %s", t.config.Ratio)
			} else {
				ratio = ratioParsed
			}
		}

		tp := trace.NewTracerProvider(
			trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(ratio))),
			trace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String(t.serviceName))),
			trace.WithSpanProcessor(bsp),
		)
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

		switch v := t.traceProvider.(type) {
		case *trace.TracerProvider:
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			err := v.ForceFlush(ctx)
			if err != nil {
				logger.WithError(err).Warn("Failed to force flush existing otel provider")
			}
			if err := v.Shutdown(ctx); err != nil {
				logger.WithError(err).Warn("Failed to shutdown existing otel provider")
			}
		}
		t.traceProvider = tp
	} else {
		t.traceProvider = noop.NewTracerProvider()
	}
	return nil
}
