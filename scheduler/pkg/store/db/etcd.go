/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	log "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	defaultKeyPrefix     = "/seldon/"
	operationGet         = "get"
	operationPut         = "put"
	operationDelete      = "delete"
	operationList        = "list"
	modelsKeyPrefix      = "models/"
	pipelinesKeyPrefix   = "pipelines/"
	experimentsKeyPrefix = "experiments/"
)

type etcdDatabase struct {
	client       *clientv3.Client
	keyPrefix    string
	logger       log.FieldLogger
	tracer       trace.Tracer
	metrics      *etcdMetrics
	enableTraces bool
}

type etcdMetrics struct {
	operationDuration *prometheus.HistogramVec
	operationErrors   *prometheus.CounterVec
	operationTotal    *prometheus.CounterVec
}

type EtcdConfig struct {
	Config       clientv3.Config
	KeyPrefix    string
	Logger       log.FieldLogger
	Tracer       trace.Tracer
	EnableTraces bool
}

func NewEtcdDatabase(cfg EtcdConfig) (Database, error) {
	cli, err := clientv3.New(cfg.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	// Set defaults
	if cfg.KeyPrefix == "" {
		cfg.KeyPrefix = defaultKeyPrefix
	}
	if cfg.Logger == nil {
		cfg.Logger = log.WithField("source", "etcdDatabase")
	}

	metrics, err := initMetrics()
	if err != nil {
		_ = cli.Close()
		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	return &etcdDatabase{
		client:       cli,
		keyPrefix:    cfg.KeyPrefix,
		logger:       cfg.Logger.WithField("source", "etcdDatabase"),
		tracer:       cfg.Tracer,
		metrics:      metrics,
		enableTraces: cfg.EnableTraces,
	}, nil
}

func initMetrics() (*etcdMetrics, error) {
	operationDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "seldon_scheduler_etcd_operation_duration_seconds",
			Help:    "Duration of etcd operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	operationErrors := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "seldon_scheduler_etcd_operation_errors_total",
			Help: "Total number of etcd operation errors",
		},
		[]string{"operation"},
	)

	operationTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "seldon_scheduler_etcd_operation_total",
			Help: "Total number of etcd operations",
		},
		[]string{"operation"},
	)

	// Register metrics
	collectors := []prometheus.Collector{operationDuration, operationErrors, operationTotal}
	for _, collector := range collectors {
		if err := prometheus.Register(collector); err != nil {
			if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
				// Metric already registered, use existing one
				switch c := are.ExistingCollector.(type) {
				case *prometheus.HistogramVec:
					if collector == operationDuration {
						operationDuration = c
					}
				case *prometheus.CounterVec:
					if collector == operationErrors {
						operationErrors = c
					} else if collector == operationTotal {
						operationTotal = c
					}
				}
			} else {
				return nil, err
			}
		}
	}

	return &etcdMetrics{
		operationDuration: operationDuration,
		operationErrors:   operationErrors,
		operationTotal:    operationTotal,
	}, nil
}

func (db *etcdDatabase) buildKey(keyType, key string) string {
	return db.keyPrefix + keyType + key
}

func (db *etcdDatabase) startSpan(ctx context.Context, operation string, key string) (context.Context, trace.Span) {
	if db.tracer == nil || !db.enableTraces {
		return ctx, nil
	}

	ctx, span := db.tracer.Start(ctx, fmt.Sprintf("etcd.%s", operation),
		trace.WithAttributes(
			attribute.String("db.system", "etcd"),
			attribute.String("db.operation", operation),
			attribute.String("db.key.path", key), //todo: might
		),
	)
	return ctx, span
}

func (db *etcdDatabase) recordMetrics(operation string, start time.Time, err error) {
	duration := time.Since(start).Seconds()
	db.metrics.operationDuration.WithLabelValues(operation).Observe(duration)
	db.metrics.operationTotal.WithLabelValues(operation).Inc()

	if err != nil {
		db.metrics.operationErrors.WithLabelValues(operation).Inc()
	}
}

func (db *etcdDatabase) GetModel(ctx context.Context, modelKey string) (*store.Model, error) {
	logger := db.logger.WithField("func", "GetModel").WithField("modelKey", modelKey)
	start := time.Now()
	var err error

	key := db.buildKey(modelsKeyPrefix, modelKey)
	ctx, span := db.startSpan(ctx, operationGet, key)
	defer func(err error) {
		if err != nil {
			if span != nil {
				span.RecordError(err)
			}
		}

		if span != nil {
			defer span.End()
		}

		db.recordMetrics(operationGet, start, err)
	}(err)

	getResp, err := db.client.Get(ctx, key)
	if err != nil {
		logger.WithError(err).Error("Failed to get model from etcd")
		return nil, fmt.Errorf("failed to get model from etcd: %w", err)
	}

	if len(getResp.Kvs) == 0 {
		return nil, nil
	}

	model := &store.Model{}
	err = json.Unmarshal(getResp.Kvs[0].Value, model)
	if err != nil {
		logger.WithError(err).Error("Failed to unmarshal model from etcd")
		return nil, fmt.Errorf("failed to unmarshal model: %w", err)
	}

	// opportunity to set more attributes in the span if we wanted to
	if span != nil {
		span.SetAttributes(
			attribute.Int64("model.revision", getResp.Kvs[0].ModRevision),
		)
	}

	return model, nil
}

func (db *etcdDatabase) Close() error {
	logger := db.logger.WithField("func", "Close")
	logger.Info("Closing etcd database connection")

	if db.client == nil {
		return nil
	}

	err := db.client.Close()
	if err != nil {
		logger.WithError(err).Error("Failed to close etcd client")
		return fmt.Errorf("failed to close etcd client: %w", err)
	}

	logger.Info("Successfully closed etcd database connection")
	return nil
}
