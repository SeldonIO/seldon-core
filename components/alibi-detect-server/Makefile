SHELL := /bin/bash
VERSION := $(shell cat ../../version.txt)
REPO=seldonio
IMAGE=alibi-detect-server

SELDON_CORE_DIR=../..

BASE_IMAGE ?= seldonio/conda-ubi8
PYTHON_VERSION ?= 3.7.10
CONDA_VERSION ?= 4.7.12



get_local_repo: clean
	cp $(SELDON_CORE_DIR)/version.txt version.txt
	cp -rT $(SELDON_CORE_DIR)/python/ _seldon_core/

clean:
	rm version.txt || true
	rm -rf _seldon_core || true

dev_install: get_local_repo
	poetry install


.PHONY: type_check
type_check:
	mypy --ignore-missing-imports adserver

.PHONY: test
test: type_check
	pytest -W ignore adserver

.PHONY: lint
lint:
	black .

#
# Local Run
#

run-outlier-detector-tensorflow:
	python -m adserver --model_name cifar10od --http_port 8080 --protocol tensorflow.http --event_type org.kubeflow.serving.inference.outlier --storage_uri gs://seldon-models/alibi-detect/od/OutlierVAE/cifar10 --event_source http://localhost:8080 OutlierDetector

run-outlier-detector-v2:
	python -m adserver --model_name cifar10od --http_port 8080 --protocol kfserving.http --event_type org.kubeflow.serving.inference.outlier --storage_uri gs://seldon-models/alibi-detect/od/OutlierVAE/cifar10 --event_source http://localhost:8080 OutlierDetector

run-metrics-server:
	SELDON_DEPLOYMENT_ID="sdepname" PREDICTIVE_UNIT_ID="modelname" PREDICTIVE_UNIT_IMAGE="adserver:0.1" PREDICTOR_ID="pred" \
		 python -m adserver --model_name model --http_port 8080 --protocol seldonfeedback.http --event_type io.seldon.serving.feedback --storage_uri "adserver.cm_models.binary_metrics.BinaryMetrics" --event_source http://localhost:8080 MetricsServer

#
# Docker Run
#

docker-run-outlier-detector-tensorflow:
	docker run --name cifar10od -it --rm -p 8080:8080 ${REPO}/${IMAGE}:${VERSION} --model_name cifar10od --http_port 8080 --protocol tensorflow.http --event_type org.kubeflow.serving.inference.outlier --storage_uri gs://seldon-models/alibi-detect/od/OutlierVAE/cifar10 --event_source http://localhost:8080 OutlierDetector

docker-run-drift-detector-tensorflow:
	docker run --name cifar10cd  -it --rm -p 8080:8080 ${REPO}/${IMAGE}:${VERSION} --model_name cifar10cd --http_port 8080 --protocol tensorflow.http --event_type org.kubeflow.serving.inference.drift --storage_uri gs://seldon-models/alibi-detect/cd/ks/cifar10-0_4_3 --event_source http://localhost:8080 DriftDetector --drift_batch_size=2

docker-run-drift-detector-torch:
	docker run --name cifar10cd  -it --rm -p 8080:8080 ${REPO}/${IMAGE}-gpu:${VERSION} --model_name cifar10cd --http_port 8080 --protocol tensorflow.http --event_type org.kubeflow.serving.inference.drift --storage_uri gs://seldon-models/alibi-detect/cd/mmd/cifar10_mmd_torch --event_source http://localhost:8080 DriftDetector --drift_batch_size=2


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
	curl -v localhost:8080/ -d @./cifar10.json -H "Alibi-Detect-Return-Feature-Score: true" -H "Alibi-Detect-Return-Instance-Score: true"

curl-metrics-server:
	curl -v -X POST -H 'Content-Type: application/json' \
	   -d '{"truth": {"data": {"ndarray": [0]}}, "response": {"data": {"ndarray": [1]}}}' \
		http://localhost:8080/

curl-metrics-server-elasticsearch:
	curl -v -X POST -H 'Content-Type: application/json' \
		-H 'Ce-Requestid: 7983e38c-29dc-45ff-8ffb-77252e7ac86d' \
		-H 'Ce-Namespace: seldon' \
		-H 'Ce-Inferenceservicename: seldon' \
		-d '{"truth": {"data": {"ndarray": [0]}}}' \
		http://localhost:8080/

curl-metrics-server-metrics:
	curl http://localhost:8080/v1/metrics



# image building logic

docker-build: get_local_repo
	docker build -f Dockerfile --build-arg BASE_IMAGE=${BASE_IMAGE} --build-arg PYTHON_VERSION=${PYTHON_VERSION} --build-arg CONDA_VERSION=${CONDA_VERSION} --build-arg VERSION=${VERSION} -t ${REPO}/${IMAGE}:${VERSION} .

docker-build-gpu:
	docker build -f Dockerfile.gpu -t ${REPO}/${IMAGE}-gpu:${VERSION} .

docker-push:
	docker push ${REPO}/${IMAGE}:${VERSION}

docker-push-gpu:
	docker push ${REPO}/${IMAGE}-gpu:${VERSION}

kind_load: docker-build
	kind load docker-image ${REPO}/${IMAGE}:${VERSION}

#
# RedHat
#

# password can be found at: https://connect.redhat.com/project/3993461/view
redhat-image-scan:
	docker pull ${REPO}/${IMAGE}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_alibi_detect} | docker login -u unused scan.connect.redhat.com --password-stdin
	docker tag ${REPO}/${IMAGE}:${VERSION} scan.connect.redhat.com/ospid-32ed6498-bce5-4c3b-9486-fe1c6e2582d3/${IMAGE}:${VERSION}
	docker push scan.connect.redhat.com/ospid-32ed6498-bce5-4c3b-9486-fe1c6e2582d3/${IMAGE}:${VERSION}
