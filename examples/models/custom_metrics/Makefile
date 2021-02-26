VERSION=0.2
IMAGE_BASE=seldonio/model-with-metrics
PYTHON_WRAPPER:= $(shell cat ../../../version.txt)

build:
	s2i build -E environment . seldonio/seldon-core-s2i-python37-ubi8:${PYTHON_WRAPPER} ${IMAGE_BASE}:${VERSION}

push:
	docker push ${IMAGE_BASE}:${VERSION}

run_local:
	export PREDICTIVE_UNIT_SERVICE_PORT=9000 && TRACING=1 JAEGER_AGENT_HOST=localhost JAEGER_AGENT_PORT=6831 JAEGER_SAMPLER_TYPE=const JAEGER_SAMPLER_PARAM=1 seldon-core-microservice --service-type MODEL ModelWithMetrics

kind_load: build
	kind load -v 3 docker-image ${IMAGE_BASE}:${VERSION}

