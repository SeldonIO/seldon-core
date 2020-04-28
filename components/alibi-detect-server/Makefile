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

docker-build: 
	docker build -f Dockerfile -t seldonio/${IMAGE}:${VERSION} .

docker-push:
	docker push seldonio/${IMAGE}:${VERSION} 

#
# Local Run
#

run-outlier-detector:
	python -m adserver --model_name cifar10od --http_port 8080 --protocol tensorflow.http --event_type org.kubeflow.serving.inference.outlier --storage_uri gs://seldon-models/alibi-detect/od/OutlierVAE/cifar10 --event_source http://localhost:8080 OutlierDetector

#
# Docker Run
#

docker-run-outlier-detector:
	docker run --name cifar10od -it --rm -p 8080:8080 seldonio/${IMAGE}:${VERSION} --model_name cifar10od --http_port 8080 --protocol tensorflow.http --event_type org.kubeflow.serving.inference.outlier --storage_uri gs://seldon-models/alibi-detect/od/OutlierVAE/cifar10 --event_source http://localhost:8080 OutlierDetector

docker-run-drift-detector:
	docker run --name cifar10cd  -it --rm -p 8080:8080 seldonio/${IMAGE}:${VERSION} --model_name cifar10cd --http_port 8080 --protocol tensorflow.http --event_type org.kubeflow.serving.inference.drift --storage_uri gs://seldon-models/alibi-detect/cd/ks/cifar10 --event_source http://localhost:8080 DriftDetector --drift_batch_size=2


#
# Test curls
#

curl-detector:
	curl -v localhost:8080/ -d @./input.json -H "ce-namespace: default" -H "ce-modelid: cifar10"

curl-outlier-detector-scores:
	curl -v localhost:8080/ -d @./input.json -H "Alibi-Detect-Return-Feature-Score: true" -H "Alibi-Detect-Return-Instance-Score: true"


#
# RedHat
#

IMAGE_NAME_BASE=alibi-detect-server
IMG_VERSION_REDHAT ?= ${IMAGE_NAME_BASE}-ubi8:${VERSION}
IMG_REDHAT ?= seldonio/${IMG_VERSION_REDHAT}

# Build the docker image for Redhat
docker-build-redhat: 
	docker build . -f Dockerfile.redhat -t ${IMG_REDHAT}

# Push the docker image
docker-push-redhat:
	docker push ${IMG_REDHAT}

# password can be found at: https://connect.redhat.com/project/3993461/view
redhat-image-scan:
	docker login -u unused scan.connect.redhat.com
	docker tag ${IMG_REDHAT} scan.connect.redhat.com/ospid-32ed6498-bce5-4c3b-9486-fe1c6e2582d3/${IMG_VERSION_REDHAT}
	docker push scan.connect.redhat.com/ospid-32ed6498-bce5-4c3b-9486-fe1c6e2582d3/${IMG_VERSION_REDHAT}
