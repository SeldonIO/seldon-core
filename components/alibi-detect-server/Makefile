SHELL := /bin/bash
VERSION := $(shell cat ../../version.txt)
IMAGE=alibi-detect-server

.PHONY: install_dev
install_dev:
	pip install -e . -r requirements-dev.txt

.PHONY: type_check
type_check:
	mypy --ignore-missing-imports adserver

.PHONY: test
test: type_check
	pytest -W ignore

.PHONY: lint
lint:
	black .

#
# Local Run
#

run-outlier-detector:
	python -m adserver --model_name cifar10od --http_port 8080 --protocol tensorflow.http --event_type org.kubeflow.serving.inference.outlier --storage_uri gs://seldon-models/alibi-detect/od/OutlierVAE/cifar10 --event_source http://localhost:8080 OutlierDetector

run-metrics-server:
	SELDON_DEPLOYMENT_ID="sdepname" PREDICTIVE_UNIT_ID="modelname" PREDICTIVE_UNIT_IMAGE="adserver:0.1" PREDICTOR_ID="pred" \
		 python -m adserver --model_name model --http_port 8080 --protocol seldonfeedback.http --event_type io.seldon.serving.feedback --storage_uri "adserver.cm_models.binary_metrics.BinaryMetrics" --event_source http://localhost:8080 MetricsServer

#
# Docker Run
#

docker-run-outlier-detector:
	docker run --name cifar10od -it --rm -p 8080:8080 seldonio/${IMAGE}:${VERSION} --model_name cifar10od --http_port 8080 --protocol tensorflow.http --event_type org.kubeflow.serving.inference.outlier --storage_uri gs://seldon-models/alibi-detect/od/OutlierVAE/cifar10 --event_source http://localhost:8080 OutlierDetector

docker-run-drift-detector:
	docker run --name cifar10cd  -it --rm -p 8080:8080 seldonio/${IMAGE}:${VERSION} --model_name cifar10cd --http_port 8080 --protocol tensorflow.http --event_type org.kubeflow.serving.inference.drift --storage_uri gs://seldon-models/alibi-detect/cd/ks/cifar10-0_4_3 --event_source http://localhost:8080 DriftDetector --drift_batch_size=2


#
# Test curls
#

curl-detector:
	curl -v localhost:8080/ -d @./input.json -H "ce-namespace: default" -H "ce-modelid: cifar10" -H "ce-type: io.seldon.serving.inference.request" -H "ce-id: 1234" -H "ce-source: localhost" -H "ce-specversion: 1.0"

curl-outlier-detector-scores:
	curl -v localhost:8080/ -d @./input.json -H "Alibi-Detect-Return-Feature-Score: true" -H "Alibi-Detect-Return-Instance-Score: true"

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

docker-build: 
	docker build --pull -f Dockerfile -t seldonio/${IMAGE}:${VERSION} .

docker-push:
	docker push seldonio/${IMAGE}:${VERSION} 

kind_load: docker-build
	kind load docker-image seldonio/${IMAGE}:${VERSION}

#
# RedHat
#

# password can be found at: https://connect.redhat.com/project/3993461/view
redhat-image-scan:
	docker pull seldonio/${IMAGE}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_alibi_detect} | docker login -u unused scan.connect.redhat.com --password-stdin
	docker tag seldonio/${IMAGE}:${VERSION} scan.connect.redhat.com/ospid-32ed6498-bce5-4c3b-9486-fe1c6e2582d3/${IMAGE}:${VERSION}
	docker push scan.connect.redhat.com/ospid-32ed6498-bce5-4c3b-9486-fe1c6e2582d3/${IMAGE}:${VERSION}
