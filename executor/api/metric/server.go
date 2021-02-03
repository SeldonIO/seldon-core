package metric

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var RecreateServerHistogram = false
var RecreateServerSummary = false

type ServerMetrics struct {
	ServerHandledHistogram *prometheus.HistogramVec
	ServerHandledSummary   *prometheus.SummaryVec
	Predictor              *v1.PredictorSpec
	DeploymentName         string
}

func NewServerMetrics(spec *v1.PredictorSpec, deploymentName string) *ServerMetrics {
	labelNames := []string{DeploymentNameMetric, PredictorNameMetric, PredictorVersionMetric, ServiceMetric, "method", "code"}

	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    ServerRequestsMetricName,
			Help:    "A histogram of latencies for executor server",
			Buckets: DefBuckets,
		},
		labelNames,
	)
	err := prometheus.Register(histogram)
	if err != nil {
		if e, ok := err.(prometheus.AlreadyRegisteredError); ok {
			if RecreateServerHistogram {
				prometheus.Unregister(e.ExistingCollector)
				prometheus.Register(histogram)
			} else {
				histogram = e.ExistingCollector.(*prometheus.HistogramVec)
			}
		}
	}

	summary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       ServerRequestsMetricName + "_summary",
			Help:       "A summary of latencies for executor server",
			Objectives: DefObjectives,
		},
		labelNames,
	)
	err = prometheus.Register(summary)
	if err != nil {
		if e, ok := err.(prometheus.AlreadyRegisteredError); ok {
			if RecreateServerSummary {
				prometheus.Unregister(e.ExistingCollector)
				prometheus.Register(summary)
			} else {
				summary = e.ExistingCollector.(*prometheus.SummaryVec)
			}
		}
	}

	return &ServerMetrics{
		ServerHandledHistogram: histogram,
		ServerHandledSummary:   summary,
		Predictor:              spec,
		DeploymentName:         deploymentName,
	}
}

// UnaryServerInterceptor is a gRPC server-side interceptor that provides Prometheus monitoring for Unary RPCs.
func (m *ServerMetrics) UnaryServerInterceptor() func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()
		resp, err := handler(ctx, req)
		st, _ := status.FromError(err)
		elapsedTime := time.Since(startTime).Seconds()
		m.ServerHandledHistogram.WithLabelValues(m.DeploymentName, m.Predictor.Name, m.Predictor.Annotations["version"], info.FullMethod, "unary", st.Code().String()).Observe(elapsedTime)
		m.ServerHandledSummary.WithLabelValues(m.DeploymentName, m.Predictor.Name, m.Predictor.Annotations["version"], info.FullMethod, "unary", st.Code().String()).Observe(elapsedTime)
		return resp, err
	}
}
