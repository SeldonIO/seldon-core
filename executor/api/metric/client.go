package metric

import (
	"context"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type ClientMetrics struct {
	ClientHandledHistogram *prometheus.HistogramVec
	Predictor              *v1.PredictorSpec
	DeploymentName         string
	ModelName              string
	ImageName              string
	ImageVersion           string
}

var RecreateClientHistogram = false

func NewClientMetrics(spec *v1.PredictorSpec, deploymentName string, modelName string) *ClientMetrics {
	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    ClientRequestsMetricName,
			Help:    "A histogram of latencies for client calls from executor",
			Buckets: DefBuckets,
		},
		[]string{DeploymentNameMetric, PredictorNameMetric, PredictorVersionMetric, ServiceMetric, ModelNameMetric, ModelImageMetric, ModelVersionMetric, "method", "code"},
	)

	err := prometheus.Register(histogram)
	if err != nil {
		if e, ok := err.(prometheus.AlreadyRegisteredError); ok {
			if RecreateClientHistogram {
				prometheus.Unregister(e.ExistingCollector)
				prometheus.Register(histogram)
			} else {
				histogram = e.ExistingCollector.(*prometheus.HistogramVec)
			}

		}
	}
	container := v1.GetContainerForPredictiveUnit(spec, modelName)
	imageName := ""
	imageVersion := ""
	if container != nil {
		imageParts := strings.Split(container.Image, ":")
		imageName = imageParts[0]
		if len(imageParts) == 2 {
			imageVersion = imageParts[1]
		}
	}

	return &ClientMetrics{
		ClientHandledHistogram: histogram,
		Predictor:              spec,
		DeploymentName:         deploymentName,
		ModelName:              modelName,
		ImageName:              imageName,
		ImageVersion:           imageVersion,
	}
}

func (m *ClientMetrics) UnaryClientInterceptor() func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		startTime := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		st, _ := status.FromError(err)
		m.ClientHandledHistogram.WithLabelValues(m.DeploymentName, m.Predictor.Name, m.Predictor.Annotations["version"], method, m.ModelName, m.ImageName, m.ImageVersion, "unary", st.Code().String()).Observe(time.Since(startTime).Seconds())
		return err
	}
}
