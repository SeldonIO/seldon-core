VERSION=0.1
IMAGE_BASE=seldonio/model-with-metrics
PYTHON_WRAPPER=0.19-SNAPSHOT

build_rest:
	s2i build -E environment_rest . seldonio/seldon-core-s2i-python36:${PYTHON_WRAPPER} ${IMAGE_BASE}_rest:${VERSION}

push_rest:
	docker push ${IMAGE_BASE}_rest:${VERSION}

build_grpc:
	s2i build -E environment_grpc . seldonio/seldon-core-s2i-python36:${PYTHON_WRAPPER} ${IMAGE_BASE}_grpc:${VERSION}

push_grpc:
	docker push ${IMAGE_BASE}_grpc:${VERSION}

run_rest_local:
	export PREDICTIVE_UNIT_SERVICE_PORT=9000 && TRACING=1 JAEGER_AGENT_HOST=localhost JAEGER_AGENT_PORT=6831 JAEGER_SAMPLER_TYPE=const JAEGER_SAMPLER_PARAM=1 seldon-core-microservice --service-type MODEL ModelWithMetrics REST

run_grpc_local:
	export PREDICTIVE_UNIT_SERVICE_PORT=9000 && TRACING=1 JAEGER_AGENT_HOST=localhost JAEGER_AGENT_PORT=6831 JAEGER_SAMPLER_TYPE=const JAEGER_SAMPLER_PARAM=1 seldon-core-microservice --service-type MODEL ModelWithMetrics GRPC


kind_load_rest: build_rest
	kind load -v 3 docker-image ${IMAGE_BASE}_rest:${VERSION}

kind_load_grpc: build_grpc
	kind load -v 3 docker-image ${IMAGE_BASE}_grpc:${VERSION}
