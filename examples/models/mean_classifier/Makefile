SHELL := /bin/bash
VERSION := $(shell cat ../../../version.txt)
IMAGE_NAME_BASE=mock_classifier
IMAGE_BASE=seldonio/${IMAGE_NAME_BASE}

KIND_NAME ?= kind


build:
	s2i build -E environment . seldonio/seldon-core-s2i-python38:${VERSION} ${IMAGE_BASE}:${VERSION}

push:
	docker push ${IMAGE_BASE}:${VERSION}

run_local:
	export PREDICTIVE_UNIT_HTTP_SERVICE_PORT=9001 && export PREDICTIVE_UNIT_GRPC_SERVICE_PORT=5001 && export TRACING=1 && export JAEGER_AGENT_HOST=localhost && export JAEGER_AGENT_PORT=6831 && export JAEGER_SAMPLER_TYPE=const && export JAEGER_SAMPLER_PARAM=1 && export SELDON_DEBUG=0 && seldon-core-microservice --service-type MODEL MeanClassifier

kind_load: build
	kind load -v 3 docker-image ${IMAGE_BASE}:${VERSION} --name ${KIND_NAME}

#
# Redhat
#

# password can be found at: https://connect.redhat.com/projects/5ea15983ebed1415210b4b26/overview
project=5ea15983ebed1415210b4b26
redhat-image-scan:
	docker pull ${IMAGE_BASE}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_mock_model} | docker login -u redhat-isv-containers+${project}-robot quay.io --password-stdin
	docker tag ${IMAGE_BASE}:${VERSION} quay.io/redhat-isv-containers/${project}:${VERSION}
	docker push quay.io/redhat-isv-containers/${project}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		preflight check container quay.io/redhat-isv-containers/${project}:${VERSION} --docker-config=${HOME}/.docker/config.json --certification-project-id=${project} --pyxis-api-token=$${pyxis_api_token} --submit
