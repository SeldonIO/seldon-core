package metrics

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const (
	HistogramName           = "seldon_infer_api_seconds"
	InferCounterName        = "seldon_infer_total"
	InferLatencyCounterName = "seldon_infer_seconds_total"

	SeldonModelMetric         = "model"
	SeldonInternalModelMetric = "model_internal"
	SeldonServerMetric        = "server"
	SeldonServerReplicaMetric = "server_replica"
	MethodTypeMetric          = "method_type"
	MethodTypeRest            = "rest"
	MethodTypeGrpc            = "grpc"
)

//TODO Revisit splitting this interface as metric handling matures
type MetricsHandler interface {
	AddHistogramMetricsHandler(baseHandler http.HandlerFunc) http.HandlerFunc
	UnaryServerInterceptor() func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error)
	AddInferMetrics(externalModelName string, internalModelName string, method string, elapsedTime float64)
}

var (
	DefaultHistogramBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10}
)

type PrometheusMetrics struct {
	serverName          string
	serverReplicaIdx    string
	namespace           string
	logger              log.FieldLogger
	histogram           *prometheus.HistogramVec
	inferCounter        *prometheus.CounterVec
	inferLatencyCounter *prometheus.CounterVec
	server              *http.Server
}

func NewPrometheusMetrics(serverName string, serverReplicaIdx uint, namespace string, logger log.FieldLogger) (*PrometheusMetrics, error) {
	namespace = safeNamespaceName(namespace)
	histogram, err := createHistogram(namespace)
	if err != nil {
		return nil, err
	}
	inferCounter, err := createInferCounter(namespace)
	if err != nil {
		return nil, err
	}
	inferLatencyCounter, err := createInferLatencyCounter(namespace)
	if err != nil {
		return nil, err
	}
	return &PrometheusMetrics{
		serverName:          serverName,
		serverReplicaIdx:    fmt.Sprintf("%d", serverReplicaIdx),
		namespace:           namespace,
		logger:              logger.WithField("source", "PrometheusMetrics"),
		histogram:           histogram,
		inferCounter:        inferCounter,
		inferLatencyCounter: inferLatencyCounter,
	}, nil
}

func safeNamespaceName(namespace string) string {
	return strings.ReplaceAll(namespace, "-", "_")
}

func createHistogram(namespace string) (*prometheus.HistogramVec, error) {
	//TODO add method for rest/grpc
	labelNames := []string{SeldonServerMetric, SeldonServerReplicaMetric, "method", "code"} //prom has not const for these!

	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:      HistogramName,
			Namespace: namespace,
			Help:      "A histogram of latencies for inference server",
			Buckets:   DefaultHistogramBuckets,
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

func createInferCounter(namespace string) (*prometheus.CounterVec, error) {
	labelNames := []string{SeldonServerMetric, SeldonServerReplicaMetric, SeldonModelMetric, SeldonInternalModelMetric, MethodTypeMetric}
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:      InferCounterName,
			Namespace: namespace,
			Help:      "A count of server inference calls",
		},
		labelNames,
	)
	err := prometheus.Register(counter)
	if err != nil {
		if e, ok := err.(prometheus.AlreadyRegisteredError); ok {
			counter = e.ExistingCollector.(*prometheus.CounterVec)
		} else {
			return nil, err
		}
	}
	return counter, nil
}

func createInferLatencyCounter(namespace string) (*prometheus.CounterVec, error) {
	labelNames := []string{SeldonServerMetric, SeldonServerReplicaMetric, SeldonModelMetric, SeldonInternalModelMetric, MethodTypeMetric}
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:      InferLatencyCounterName,
			Namespace: namespace,
			Help:      "A sum of server inference call latencies",
		},
		labelNames,
	)
	err := prometheus.Register(counter)
	if err != nil {
		if e, ok := err.(prometheus.AlreadyRegisteredError); ok {
			counter = e.ExistingCollector.(*prometheus.CounterVec)
		} else {
			return nil, err
		}
	}
	return counter, nil
}

func (pm *PrometheusMetrics) AddHistogramMetricsHandler(baseHandler http.HandlerFunc) http.HandlerFunc {
	handler := promhttp.InstrumentHandlerDuration(
		pm.histogram.MustCurryWith(prometheus.Labels{
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
		pm.histogram.WithLabelValues(pm.serverName, pm.serverReplicaIdx, "grpc", st.Code().String()).Observe(elapsedTime)
		return resp, err
	}
}

func (pm *PrometheusMetrics) AddInferMetrics(externalModelName string, internalModelName string, method string, latency float64) {
	pm.addInferLatency(externalModelName, internalModelName, method, latency)
	pm.addInferCount(externalModelName, internalModelName, method)
}

func (pm *PrometheusMetrics) addInferCount(externalModelName string, modelInternalName string, method string) {
	pm.inferCounter.With(prometheus.Labels{
		SeldonModelMetric:         externalModelName,
		SeldonInternalModelMetric: modelInternalName,
		SeldonServerMetric:        pm.serverName,
		SeldonServerReplicaMetric: pm.serverReplicaIdx,
		MethodTypeMetric:          method,
	}).Inc()
}

func (pm *PrometheusMetrics) addInferLatency(externalModelName string, modelInternalName string, method string, latency float64) {
	pm.inferLatencyCounter.With(prometheus.Labels{
		SeldonModelMetric:         externalModelName,
		SeldonInternalModelMetric: modelInternalName,
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
	return pm.server.Shutdown(context.TODO())
}
