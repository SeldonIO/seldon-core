package metric

const (
	CodeMetric             = "code"    // 2xx, 5xx etc
	HTTPMethodMetric       = "method"  // Http Method (Post, Get etc)
	ServiceMetric          = "service" // http or grpc service: prediction, feedback etc
	DeploymentNameMetric   = "deployment_name"
	PredictorNameMetric    = "predictor_name"
	PredictorVersionMetric = "predictor_version"
	ModelNameMetric        = "model_name"
	ModelImageMetric       = "model_image"
	ModelVersionMetric     = "model_version"

	ServerRequestsMetricName = "seldon_api_executor_server_requests_seconds"
	ClientRequestsMetricName = "seldon_api_executor_client_requests_seconds"

	PredictionHttpServiceName = "predictions"
	StatusHttpServiceName     = "status"
	MetadataHttpServiceName   = "metadata"
	FeedbackHttpServiceName   = "feedback"
)

var (
	DefBuckets    = []float64{0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10}
	DefObjectives = map[float64]float64{0.5: 0.05, 0.75: 0.025, 0.9: 0.01, 0.99: 0.001, 1.0: 0}
)
