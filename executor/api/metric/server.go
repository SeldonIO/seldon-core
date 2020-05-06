package metric

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"time"
)

type ServerMetrics struct {
	ServerHandledHistogram *prometheus.HistogramVec
	Predictor              *v1.PredictorSpec
	DeploymentName         string
}

func NewServerMetrics(spec *v1.PredictorSpec, deploymentName string) *ServerMetrics {
	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    ServerRequestsMetricName,
			Help:    "A histogram of latencies for executor server",
			Buckets: prometheus.DefBuckets,
		},
		[]string{DeploymentNameMetric, PredictorNameMetric, PredictorVersionMetric, ServiceMetric, "method", "code"},
	)
	err := prometheus.Register(histogram)
	if err != nil {
		prometheus.Unregister(histogram)
		prometheus.Register(histogram)
	}
	return &ServerMetrics{
		ServerHandledHistogram: histogram,
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
		m.ServerHandledHistogram.WithLabelValues(m.DeploymentName, m.Predictor.Name, m.Predictor.Annotations["version"], info.FullMethod, "unary", st.Code().String()).Observe(time.Since(startTime).Seconds())
		return resp, err
	}
}
