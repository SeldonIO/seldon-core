SHELL := /bin/bash
VERSION := $(shell cat ../../../version.txt)
IMAGE_NAME_BASE=mock_classifier
IMAGE_BASE=seldonio/${IMAGE_NAME_BASE}


build_%:
	s2i build -E environment_$* . seldonio/seldon-core-s2i-python37-ubi8:${VERSION} ${IMAGE_BASE}_$*:${VERSION}

push_%:
	docker push ${IMAGE_BASE}_$*:${VERSION}

.PHONY: build_all
build_all: build_rest build_grpc
.PHONY: push_all
push_all:push_rest push_grpc

run_rest_local:
	export PREDICTIVE_UNIT_SERVICE_PORT=9000 && TRACING=1 JAEGER_AGENT_HOST=localhost JAEGER_AGENT_PORT=6831 JAEGER_SAMPLER_TYPE=const JAEGER_SAMPLER_PARAM=1 seldon-core-microservice --service-type MODEL MeanClassifier REST

run_grpc_local:
	export PREDICTIVE_UNIT_SERVICE_PORT=9000 && TRACING=1 JAEGER_AGENT_HOST=localhost JAEGER_AGENT_PORT=6831 JAEGER_SAMPLER_TYPE=const JAEGER_SAMPLER_PARAM=1 seldon-core-microservice --service-type MODEL MeanClassifier GRPC


kind_load_%:
	kind load -v 3 docker-image ${IMAGE_BASE}_$*:${VERSION}
.PHONY: kind_load_all
kind_load_all: kind_load_rest kind_load_grpc

#
# Redhat
#

# https://connect.redhat.com/project/4035711/view
scan_rest=ospid-c90fdfe6-d054-4598-baa8-7f7aac3ed63a
# no groc at present
scan_grpc=
redhat-image-scan-%:
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_mock_model_$*} | docker login -u unused scan.connect.redhat.com --password-stdin
	docker tag ${IMAGE_BASE}_$*:${VERSION} scan.connect.redhat.com/${scan_$*}/${IMAGE_NAME_BASE}_$*:${VERSION}
	docker push scan.connect.redhat.com/${scan_$*}/${IMAGE_NAME_BASE}_$*:${VERSION}

.PHONY: redhat-image-scan-all
redhat-image-scan: redhat-image-scan-rest 

