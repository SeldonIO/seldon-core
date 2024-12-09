/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package metrics

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// Keep next line as used in docs
// Docs Start Metrics
// Model metrics
const (
	// Histograms do no include pipeline label for efficiency
	modelHistogramName = "seldon_model_infer_api_seconds"
	// We use base infer counters to store core metrics per pipeline
	modelInferCounterName                 = "seldon_model_infer_total"
	modelInferLatencyCounterName          = "seldon_model_infer_seconds_total"
	modelAggregateInferCounterName        = "seldon_model_aggregate_infer_total"
	modelAggregateInferLatencyCounterName = "seldon_model_aggregate_infer_seconds_total"
)

// Agent metrics
const (
	cacheEvictCounterName                              = "seldon_cache_evict_count"
	cacheMissCounterName                               = "seldon_cache_miss_count"
	loadModelCounterName                               = "seldon_load_model_counter"
	unloadModelCounterName                             = "seldon_unload_model_counter"
	loadedModelGaugeName                               = "seldon_loaded_model_gauge"
	loadedModelMemoryGaugeName                         = "seldon_loaded_model_memory_bytes_gauge"
	evictedModelMemoryGaugeName                        = "seldon_evicted_model_memory_bytes_gauge"
	serverReplicaMemoryCapacityGaugeName               = "seldon_server_replica_memory_capacity_bytes_gauge"
	serverReplicaMemoryCapacityWithOverCommitGaugeName = "seldon_server_replica_memory_capacity_overcommit_bytes_gauge"
)

// Docs End Metrics
// Keep above line as used in docs

// Metric labels
const (
	SeldonModelMetric         = "model"
	SeldonInternalModelMetric = "model_internal"
	SeldonServerMetric        = "server"
	SeldonServerReplicaMetric = "server_replica"
	MethodTypeMetric          = "method_type"
	MethodTypeRest            = "rest"
	MethodTypeGrpc            = "grpc"
	MethodMetric              = "method"
	CodeMetric                = "code"
)

// TODO Revisit splitting this interface as metric handling matures
type AgentMetricsHandler interface {
	AddModelHistogramMetricsHandler(baseHandler http.HandlerFunc) http.HandlerFunc
	UnaryServerInterceptor() func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error)
	AddModelInferMetrics(externalModelName string, internalModelName string, method string, elapsedTime float64, code string)
	AddLoadedModelMetrics(internalModelName string, memory uint64, isLoad, isSoft bool)
	AddServerReplicaMetrics(memory uint64, memoryWithOvercommit float32)
}

var (
	DefaultHistogramBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10}
)

type PrometheusMetrics struct {
	serverName       string
	serverReplicaIdx string
	logger           log.FieldLogger
	// Model metrics
	modelHistogram                                 *prometheus.HistogramVec
	modelInferCounter                              *prometheus.CounterVec
	modelInferLatencyCounter                       *prometheus.CounterVec
	modelAggregateInferCounter                     *prometheus.CounterVec
	modelAggregateInferLatencyCounter              *prometheus.CounterVec
	cacheEvictCounter                              *prometheus.CounterVec
	cacheMissCounter                               *prometheus.CounterVec
	loadModelCounter                               *prometheus.CounterVec
	unloadModelCounter                             *prometheus.CounterVec
	loadedModelGauge                               *prometheus.GaugeVec
	loadedModelMemoryGauge                         *prometheus.GaugeVec
	evictedModelMemoryGauge                        *prometheus.GaugeVec
	serverReplicaMemoryCapacityGauge               *prometheus.GaugeVec
	serverReplicaMemoryCapacityWithOvercommitGauge *prometheus.GaugeVec
	server                                         *http.Server
}

func NewPrometheusModelMetrics(serverName string, serverReplicaIdx uint, logger log.FieldLogger) (*PrometheusMetrics, error) {
	histogram, err := createModelHistogram()
	if err != nil {
		return nil, err
	}

	inferCounter, err := createModelInferCounter()
	if err != nil {
		return nil, err
	}

	inferLatencyCounter, err := createModelInferLatencyCounter()
	if err != nil {
		return nil, err
	}

	aggregateInferCounter, err := createModelAggregateInferCounter()
	if err != nil {

		return nil, err
	}

	aggregateInferLatencyCounter, err := createModelAggregateInferLatencyCounter()
	if err != nil {
		return nil, err
	}

	cacheEvictCounter, err := createCacheEvictCounter()
	if err != nil {
		return nil, err
	}

	cacheMissCounter, err := createCacheMissCounter()
	if err != nil {
		return nil, err
	}

	loadModelCounter, err := createLoadModelCounter()
	if err != nil {
		return nil, err
	}

	unloadModelCounter, err := createUnloadModelCounter()
	if err != nil {
		return nil, err
	}

	loadedModelGauge, err := createLoadedModelGauge()
	if err != nil {
		return nil, err
	}

	loadedModelMemoryGauge, err := createLoadedModelMemoryGauge()
	if err != nil {
		return nil, err
	}

	evictedModelMemoryGauge, err := createEvictedModelMemoryGauge()
	if err != nil {
		return nil, err
	}

	serverReplicaMemoryCapacityGauge, err := createServerReplicaMemoryCapacityGauge()
	if err != nil {
		return nil, err
	}

	serverReplicaMemoryCapacityWithOvercommitGauge, err := createServerReplicaMemoryCapacityWithOvercommitGauge()
	if err != nil {
		return nil, err
	}

	return &PrometheusMetrics{
		serverName:                        serverName,
		serverReplicaIdx:                  fmt.Sprintf("%d", serverReplicaIdx),
		logger:                            logger.WithField("source", "PrometheusMetrics"),
		modelHistogram:                    histogram,
		modelInferCounter:                 inferCounter,
		modelInferLatencyCounter:          inferLatencyCounter,
		modelAggregateInferCounter:        aggregateInferCounter,
		modelAggregateInferLatencyCounter: aggregateInferLatencyCounter,
		cacheEvictCounter:                 cacheEvictCounter,
		cacheMissCounter:                  cacheMissCounter,
		loadModelCounter:                  loadModelCounter,
		unloadModelCounter:                unloadModelCounter,
		loadedModelGauge:                  loadedModelGauge,
		loadedModelMemoryGauge:            loadedModelMemoryGauge,
		evictedModelMemoryGauge:           evictedModelMemoryGauge,
		serverReplicaMemoryCapacityGauge:  serverReplicaMemoryCapacityGauge,
		serverReplicaMemoryCapacityWithOvercommitGauge: serverReplicaMemoryCapacityWithOvercommitGauge,
	}, nil
}

func createModelHistogram() (*prometheus.HistogramVec, error) {
	//TODO add method for rest/grpc
	labelNames := []string{SeldonServerMetric, SeldonServerReplicaMetric, MethodMetric, CodeMetric}

	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    modelHistogramName,
			Help:    "A histogram of latencies for inference server",
			Buckets: DefaultHistogramBuckets,
		},
		labelNames,
	)
	err := prometheus.Register(histogram)
	if err != nil {
		if e, ok := err.(prometheus.AlreadyRegisteredError); ok {
			histogram = e.ExistingCollector.(*prometheus.HistogramVec)
		} else {
			return nil, err
		}
	}
	return histogram, nil
}

func createModelInferCounter() (*prometheus.CounterVec, error) {
	labelNames := []string{SeldonServerMetric, SeldonServerReplicaMetric, SeldonModelMetric, SeldonInternalModelMetric, MethodTypeMetric, CodeMetric}
	return createCounterVec(
		modelInferCounterName,
		"A count of server inference calls",
		labelNames,
	)
}

func createModelInferLatencyCounter() (*prometheus.CounterVec, error) {
	labelNames := []string{SeldonServerMetric, SeldonServerReplicaMetric, SeldonModelMetric, SeldonInternalModelMetric, MethodTypeMetric, CodeMetric}
	return createCounterVec(
		modelInferLatencyCounterName,
		"A sum of server inference call latencies",
		labelNames,
	)
}

func createModelAggregateInferCounter() (*prometheus.CounterVec, error) {
	labelNames := []string{SeldonServerMetric, SeldonServerReplicaMetric, MethodTypeMetric}
	return createCounterVec(
		modelAggregateInferCounterName,
		"A count of server inference calls (aggregate)",
		labelNames,
	)
}

func createModelAggregateInferLatencyCounter() (*prometheus.CounterVec, error) {
	labelNames := []string{SeldonServerMetric, SeldonServerReplicaMetric, MethodTypeMetric}
	return createCounterVec(
		modelAggregateInferLatencyCounterName,
		"A sum of server inference call latencies (aggregate)",
		labelNames,
	)
}

func createCacheEvictCounter() (*prometheus.CounterVec, error) {
	labelNames := []string{SeldonServerMetric, SeldonServerReplicaMetric}
	return createCounterVec(
		cacheEvictCounterName,
		"A count of model cache evict",
		labelNames,
	)
}

func createCacheMissCounter() (*prometheus.CounterVec, error) {
	labelNames := []string{SeldonServerMetric, SeldonServerReplicaMetric}
	return createCounterVec(
		cacheMissCounterName,
		"A count of model cache miss",
		labelNames,
	)
}

func createLoadModelCounter() (*prometheus.CounterVec, error) {
	labelNames := []string{SeldonServerMetric, SeldonServerReplicaMetric}
	return createCounterVec(
		loadModelCounterName,
		"A count of model load",
		labelNames,
	)
}

func createUnloadModelCounter() (*prometheus.CounterVec, error) {
	labelNames := []string{SeldonServerMetric, SeldonServerReplicaMetric}
	return createCounterVec(
		unloadModelCounterName,
		"A count of model unload",
		labelNames,
	)
}

func createLoadedModelGauge() (*prometheus.GaugeVec, error) {
	labelNames := []string{SeldonServerMetric, SeldonServerReplicaMetric, SeldonInternalModelMetric}
	return createGaugeVec(
		loadedModelGaugeName,
		"A gauge of models loaded in the system",
		labelNames,
	)
}

func createLoadedModelMemoryGauge() (*prometheus.GaugeVec, error) {
	labelNames := []string{SeldonServerMetric, SeldonServerReplicaMetric, SeldonInternalModelMetric}
	return createGaugeVec(
		loadedModelMemoryGaugeName,
		"A gauge of models loaded memory in the system",
		labelNames,
	)
}

func createEvictedModelMemoryGauge() (*prometheus.GaugeVec, error) {
	labelNames := []string{SeldonServerMetric, SeldonServerReplicaMetric, SeldonInternalModelMetric}
	return createGaugeVec(
		evictedModelMemoryGaugeName,
		"A gauge of models evicted from memory in the system",
		labelNames,
	)
}

func createServerReplicaMemoryCapacityGauge() (*prometheus.GaugeVec, error) {
	labelNames := []string{SeldonServerMetric, SeldonServerReplicaMetric}
	return createGaugeVec(
		serverReplicaMemoryCapacityGaugeName,
		"A gauge of server replica memory capacity",
		labelNames,
	)
}

func createServerReplicaMemoryCapacityWithOvercommitGauge() (*prometheus.GaugeVec, error) {
	labelNames := []string{SeldonServerMetric, SeldonServerReplicaMetric}
	return createGaugeVec(
		serverReplicaMemoryCapacityWithOverCommitGaugeName,
		"A gauge of server replica memory capacity with overcommit",
		labelNames,
	)
}

func (pm *PrometheusMetrics) AddModelHistogramMetricsHandler(baseHandler http.HandlerFunc) http.HandlerFunc {
	handler := promhttp.InstrumentHandlerDuration(
		pm.modelHistogram.MustCurryWith(prometheus.Labels{
			SeldonServerMetric:        pm.serverName,
			SeldonServerReplicaMetric: pm.serverReplicaIdx,
		}),
		baseHandler,
	)
	return handler
}

func (pm *PrometheusMetrics) UnaryServerInterceptor() func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()
		resp, err := handler(ctx, req)
		st, _ := status.FromError(err)
		elapsedTime := time.Since(startTime).Seconds()
		pm.modelHistogram.WithLabelValues(pm.serverName, pm.serverReplicaIdx, "grpc", st.Code().String()).Observe(elapsedTime)
		return resp, err
	}
}

func (pm *PrometheusMetrics) AddModelInferMetrics(externalModelName string, internalModelName string, method string, latency float64, code string) {
	pm.addInferLatency(externalModelName, internalModelName, method, latency, code)
	pm.addInferCount(externalModelName, internalModelName, method, code)
}

func (pm *PrometheusMetrics) AddLoadedModelMetrics(internalModelName string, memBytes uint64, isLoad, isSoft bool) {
	if isLoad {
		if isSoft {
			pm.addCacheMissCount()
			pm.addEvictedModelMemoryMetrics(internalModelName, memBytes, false) // remove it from disk
		} else {
			pm.addLoadCount()
			pm.addLoadedModelMetrics(internalModelName, isLoad)
		}
		pm.addLoadedModelMemoryMetrics(internalModelName, memBytes, isLoad)
	} else {
		if isSoft {
			pm.addCacheEvictCount()
			pm.addEvictedModelMemoryMetrics(internalModelName, memBytes, true)
		} else {
			pm.addLoadedModelMetrics(internalModelName, isLoad)
			pm.addEvictedModelMemoryMetrics(internalModelName, memBytes, false)
			pm.addUnloadCount()
		}
		pm.addLoadedModelMemoryMetrics(internalModelName, memBytes, isLoad)
	}
}

func (pm *PrometheusMetrics) AddServerReplicaMetrics(memory uint64, memoryWithOvercommit float32) {
	pm.addServerReplicaMemoryCapacityMetrics(memory, memoryWithOvercommit)
}

func (pm *PrometheusMetrics) addServerReplicaMemoryCapacityMetrics(memBytes uint64, memBytesWithOverCommit float32) {
	pm.serverReplicaMemoryCapacityGauge.With(prometheus.Labels{
		SeldonServerMetric:        pm.serverName,
		SeldonServerReplicaMetric: pm.serverReplicaIdx,
	}).Set(float64(memBytes))
	pm.serverReplicaMemoryCapacityWithOvercommitGauge.With(prometheus.Labels{
		SeldonServerMetric:        pm.serverName,
		SeldonServerReplicaMetric: pm.serverReplicaIdx,
	}).Set(float64(memBytesWithOverCommit))
}

func (pm *PrometheusMetrics) addLoadedModelMetrics(internalModelName string, isLoad bool) {
	if isLoad {
		pm.loadedModelGauge.With(prometheus.Labels{
			SeldonInternalModelMetric: internalModelName,
			SeldonServerMetric:        pm.serverName,
			SeldonServerReplicaMetric: pm.serverReplicaIdx,
		}).Set(1)
	} else {
		pm.loadedModelGauge.With(prometheus.Labels{
			SeldonInternalModelMetric: internalModelName,
			SeldonServerMetric:        pm.serverName,
			SeldonServerReplicaMetric: pm.serverReplicaIdx,
		}).Set(0)
	}
}

func (pm *PrometheusMetrics) addLoadedModelMemoryMetrics(internalModelName string, memBytes uint64, isLoad bool) {
	if isLoad {
		pm.loadedModelMemoryGauge.With(prometheus.Labels{
			SeldonInternalModelMetric: internalModelName,
			SeldonServerMetric:        pm.serverName,
			SeldonServerReplicaMetric: pm.serverReplicaIdx,
		}).Set(float64(memBytes))
	} else {
		pm.loadedModelMemoryGauge.With(prometheus.Labels{
			SeldonInternalModelMetric: internalModelName,
			SeldonServerMetric:        pm.serverName,
			SeldonServerReplicaMetric: pm.serverReplicaIdx,
		}).Set(0)
	}
}

func (pm *PrometheusMetrics) addEvictedModelMemoryMetrics(internalModelName string, memBytes uint64, isLoad bool) {
	// isLoad means "loaded" to disk
	if isLoad {
		pm.evictedModelMemoryGauge.With(prometheus.Labels{
			SeldonInternalModelMetric: internalModelName,
			SeldonServerMetric:        pm.serverName,
			SeldonServerReplicaMetric: pm.serverReplicaIdx,
		}).Set(float64(memBytes))
	} else {
		pm.evictedModelMemoryGauge.With(prometheus.Labels{
			SeldonInternalModelMetric: internalModelName,
			SeldonServerMetric:        pm.serverName,
			SeldonServerReplicaMetric: pm.serverReplicaIdx,
		}).Set(0)
	}
}

func (pm *PrometheusMetrics) addLoadCount() {
	pm.loadModelCounter.With(prometheus.Labels{
		SeldonServerMetric:        pm.serverName,
		SeldonServerReplicaMetric: pm.serverReplicaIdx,
	}).Inc()
}

func (pm *PrometheusMetrics) addUnloadCount() {
	pm.unloadModelCounter.With(prometheus.Labels{
		SeldonServerMetric:        pm.serverName,
		SeldonServerReplicaMetric: pm.serverReplicaIdx,
	}).Inc()
}

func (pm *PrometheusMetrics) addCacheMissCount() {
	pm.cacheMissCounter.With(prometheus.Labels{
		SeldonServerMetric:        pm.serverName,
		SeldonServerReplicaMetric: pm.serverReplicaIdx,
	}).Inc()
}

func (pm *PrometheusMetrics) addCacheEvictCount() {
	pm.cacheEvictCounter.With(prometheus.Labels{
		SeldonServerMetric:        pm.serverName,
		SeldonServerReplicaMetric: pm.serverReplicaIdx,
	}).Inc()
}

func (pm *PrometheusMetrics) addInferCount(externalModelName, internalModelName, method string, code string) {
	pm.modelInferCounter.With(prometheus.Labels{
		SeldonModelMetric:         externalModelName,
		SeldonInternalModelMetric: internalModelName,
		SeldonServerMetric:        pm.serverName,
		SeldonServerReplicaMetric: pm.serverReplicaIdx,
		MethodTypeMetric:          method,
		CodeMetric:                code,
	}).Inc()
	pm.modelAggregateInferCounter.With(prometheus.Labels{
		SeldonServerMetric:        pm.serverName,
		SeldonServerReplicaMetric: pm.serverReplicaIdx,
		MethodTypeMetric:          method,
	}).Inc()
}

func (pm *PrometheusMetrics) addInferLatency(externalModelName, internalModelName, method string, latency float64, code string) {
	pm.modelInferLatencyCounter.With(prometheus.Labels{
		SeldonModelMetric:         externalModelName,
		SeldonInternalModelMetric: internalModelName,
		SeldonServerMetric:        pm.serverName,
		SeldonServerReplicaMetric: pm.serverReplicaIdx,
		MethodTypeMetric:          method,
		CodeMetric:                code,
	}).Add(latency)
	pm.modelAggregateInferLatencyCounter.With(prometheus.Labels{
		SeldonServerMetric:        pm.serverName,
		SeldonServerReplicaMetric: pm.serverReplicaIdx,
		MethodTypeMetric:          method,
	}).Add(latency)
}

func (pm *PrometheusMetrics) Start(port int) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	pm.server = &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: mux,
	}
	pm.logger.Infof("Starting metrics server on port %d", port)
	return pm.server.ListenAndServe()
}

func (pm *PrometheusMetrics) Stop() error {
	pm.logger.Info("Graceful shutdown")
	if pm.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), util.ServerControlPlaneTimeout)
		defer cancel()
		return pm.server.Shutdown(ctx)
	}
	return nil
}
