package controllers

const (
	ENV_ISTIO_ENABLED                = "ISTIO_ENABLED"
	ENV_ISTIO_GATEWAY                = "ISTIO_GATEWAY"
	ENV_ISTIO_TLS_MODE               = "ISTIO_TLS_MODE"
	ANNOTATION_ISTIO_GATEWAY         = "seldon.io/istio-gateway"
	ANNOTATION_ISTIO_RETRIES         = "seldon.io/istio-retries"
	ANNOTATION_ISTIO_RETRIES_TIMEOUT = "seldon.io/istio-retries-timeout"
	ANNOTATION_ISTIO_HOST            = "seldon.io/istio-host"
)
