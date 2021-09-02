SHELL := /bin/bash
VERSION := $(shell cat ../../../version.txt)
IMAGE_NAME_BASE=mock_classifier
IMAGE_BASE=seldonio/${IMAGE_NAME_BASE}


build:
	s2i build -E environment . seldonio/seldon-core-s2i-python37-ubi8:${VERSION} ${IMAGE_BASE}:${VERSION}

push:
	docker push ${IMAGE_BASE}:${VERSION}

run_local:
	export PREDICTIVE_UNIT_HTTP_SERVICE_PORT=9001 && export PREDICTIVE_UNIT_GRPC_SERVICE_PORT=5001 && export TRACING=1 && export JAEGER_AGENT_HOST=localhost && export JAEGER_AGENT_PORT=6831 && export JAEGER_SAMPLER_TYPE=const && export JAEGER_SAMPLER_PARAM=1 && export SELDON_DEBUG=0 && seldon-core-microservice --service-type MODEL MeanClassifier

kind_load: build
	kind load -v 3 docker-image ${IMAGE_BASE}:${VERSION}

#
# Redhat
#

# https://connect.redhat.com/project/4035711/view
scan=ospid-c90fdfe6-d054-4598-baa8-7f7aac3ed63a
redhat-image-scan:
	docker pull ${IMAGE_BASE}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_mock_model} | docker login -u unused scan.connect.redhat.com --password-stdin
	docker tag ${IMAGE_BASE}:${VERSION} scan.connect.redhat.com/${scan}/${IMAGE_NAME_BASE}:${VERSION}
	docker push scan.connect.redhat.com/${scan}/${IMAGE_NAME_BASE}:${VERSION}



