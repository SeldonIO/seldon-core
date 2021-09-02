SHELL := /bin/bash
VERSION := $(shell cat ../../../version.txt)
IMAGE_NAME_BASE=seldontest_
IMAGE_BASE=seldonio/${IMAGE_NAME_BASE}


build_%:
	s2i build -E environment_$* . seldonio/seldon-core-s2i-python37-ubi8:${VERSION} ${IMAGE_BASE}$*:${VERSION}


push_%:
	docker push ${IMAGE_BASE}$*:${VERSION}


run_local_%:
	export PREDICTIVE_UNIT_HTTP_SERVICE_PORT=9001 && export PREDICTIVE_UNIT_GRPC_SERVICE_PORT=5001 && export TRACING=1 && export JAEGER_AGENT_HOST=localhost && export JAEGER_AGENT_PORT=6831 && export JAEGER_SAMPLER_TYPE=const && export JAEGER_SAMPLER_PARAM=1 && export SELDON_DEBUG=0 && seldon-core-microservice --service-type MODEL Fixed$*

kind_load_%: 
	kind load -v 3 docker-image ${IMAGE_BASE}$*:${VERSION}


.PHONY: run_build_all
run_build_all: build_predict build_predict_raw

.PHONY: run_push_all
run_push_all: push_predict push_predict_raw

.PHONY: run_kind_load_all
run_kind_load_all: run_build_all kind_load_predict kind_load_predict_raw


