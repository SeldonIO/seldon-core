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
	FeedbackHttpServiceName   = "feedback"
)
