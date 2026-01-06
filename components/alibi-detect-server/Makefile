SHELL := /bin/bash

VERSION ?= $(shell cat ../../version.txt)
DOCKER_REGISTRY ?= seldonio

BASE_IMAGE ?= ${DOCKER_REGISTRY}/conda-ubi9
IMAGE ?= ${DOCKER_REGISTRY}/alibi-detect-server
KIND_NAME ?= kind

SELDON_CORE_DIR=../..


get_local_repo: clean
	cp $(SELDON_CORE_DIR)/version.txt version.txt
	cp -R "$(SELDON_CORE_DIR)/python/." _seldon_core/

clean:
	rm version.txt || true
	rm -rf _seldon_core || true

dev_install: get_local_repo
	poetry install


.PHONY: type_check
type_check:
	poetry run mypy --ignore-missing-imports adserver

.PHONY: test
test: type_check
	poetry run pytest -W ignore adserver

.PHONY: lint
lint:
	poetry run black .

#
# Local Run
#

run-outlier-detector-tensorflow:
	TF_USE_LEGACY_KERAS=1 python -m adserver --model_name cifar10od --http_port 8080 --protocol tensorflow.http --event_type org.kubeflow.serving.inference.outlier --storage_uri gs://seldon-models/alibi-detect/od/OutlierVAE/cifar10 --event_source http://localhost:8080 OutlierDetector

run-outlier-detector-v2:
	TF_USE_LEGACY_KERAS=1 python -m adserver --model_name cifar10od --http_port 8080 --protocol kfserving.http --event_type org.kubeflow.serving.inference.outlier --storage_uri gs://seldon-models/alibi-detect/od/OutlierVAE/cifar10 --event_source http://localhost:8080 OutlierDetector

run-metrics-server:
	TF_USE_LEGACY_KERAS=1 SELDON_DEPLOYMENT_ID="sdepname" PREDICTIVE_UNIT_ID="modelname" PREDICTIVE_UNIT_IMAGE="adserver:0.1" PREDICTOR_ID="pred" \
		 python -m adserver --model_name model --http_port 8080 --protocol seldonfeedback.http --event_type io.seldon.serving.feedback --storage_uri "adserver.cm_models.binary_metrics.BinaryMetrics" --event_source http://localhost:8080 MetricsServer

#
# Docker Run
#

docker-run-outlier-detector-tensorflow:
	docker run --name cifar10od -e TF_USE_LEGACY_KERAS=1 -it --rm -p 8080:8080 ${IMAGE}:${VERSION} --model_name cifar10od --http_port 8080 --protocol tensorflow.http --event_type org.kubeflow.serving.inference.outlier --storage_uri gs://seldon-models/alibi-detect/od/OutlierVAE/cifar10 --event_source http://localhost:8080 OutlierDetector

docker-run-drift-detector-tensorflow:
	docker run --name cifar10cd -e TF_USE_LEGACY_KERAS=1 -it --rm -p 8080:8080 ${IMAGE}:${VERSION} --model_name cifar10cd --http_port 8080 --protocol tensorflow.http --event_type org.kubeflow.serving.inference.drift --storage_uri gs://seldon-models/alibi-detect/cd/ks/cifar10-0_6_2 --event_source http://localhost:8080 DriftDetector --drift_batch_size=2

#
# Test curls
#

curl-detector-tensorflow:
	curl -v localhost:8080/ -d @./cifar10.json -H "ce-namespace: default" -H "ce-modelid: cifar10" -H "ce-type: io.seldon.serving.inference.request" -H "ce-id: 1234" -H "ce-source: localhost" -H "ce-specversion: 1.0"

curl-detector-v2:
	curl -v localhost:8080/ -d @./cifar10-v2.json -H "ce-namespace: default" -H "ce-modelid: cifar10" -H "ce-type: io.seldon.serving.inference.request" -H "ce-id: 1234" -H "ce-source: localhost" -H "ce-specversion: 1.0"

curl-detector-v2-outlier:
	curl -v localhost:8080/ -d @./cifar10-v2-outlier.json -H "ce-namespace: default" -H "ce-modelid: cifar10" -H "ce-type: io.seldon.serving.inference.request" -H "ce-id: 1234" -H "ce-source: localhost" -H "ce-specversion: 1.0"

curl-tensorflow-outlier-detector-scores:
	curl -v localhost:8080/ -d @./cifar10.json -H "ce-source: localhost" -H "ce-type: io.seldon.serving.inference.request" -H "ce-id: 1234" -H "ce-specversion: 1.0" -H "Alibi-Detect-Return-Feature-Score: true" -H "Alibi-Detect-Return-Instance-Score: true"


curl-metrics-server-feedback:
	curl -v -X POST -H 'Content-Type: application/json' \
	 	-H "ce-source: localhost" \
		-H "ce-type: io.seldon.serving.feedback" \
		-H "ce-id: 1234" \
		-H "ce-specversion: 1.0" \
		-d '{"truth": {"data": {"ndarray": [0]}}, "response": {"data": {"ndarray": [1]}}}' \
		http://localhost:8080/


curl-metrics-server-get:
	curl http://localhost:8080/v1/metrics



# image building logic

docker-build: get_local_repo
	docker build --platform linux/amd64 -f Dockerfile --build-arg BASE_IMAGE=${BASE_IMAGE} --build-arg VERSION=${VERSION} -t ${IMAGE}:${VERSION} .

docker-build-and-push-prod:
	docker buildx build --provenance=true --attest type=sbom,generator=docker/scout-sbom-indexer:latest --platform linux/amd64 -f Dockerfile --build-arg BASE_IMAGE=${BASE_IMAGE} --build-arg VERSION=${VERSION} -t ${IMAGE}:${VERSION} . --push

docker-push:
	docker push ${IMAGE}:${VERSION}

kind_load: docker-build
	kind load docker-image ${IMAGE}:${VERSION} --name ${KIND_NAME}

#
# RedHat
#

# password can be found at: https://connect.redhat.com/projects/5e9d53076c2dde3913c2bb66/overview
project=5e9d53076c2dde3913c2bb66
redhat-image-scan:
	docker pull ${IMAGE}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_alibi_detect} | docker login -u redhat-isv-containers+${project}-robot quay.io --password-stdin
	docker tag ${IMAGE}:${VERSION} quay.io/redhat-isv-containers/${project}:${VERSION}
	docker push quay.io/redhat-isv-containers/${project}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		preflight check container quay.io/redhat-isv-containers/${project}:${VERSION} --docker-config=${HOME}/.docker/config.json --certification-project-id=${project} --pyxis-api-token=$${pyxis_api_token} --submit
