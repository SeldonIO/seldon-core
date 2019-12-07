IMAGE_VERSION=0.3
IMAGE_NAME = docker.io/seldonio/seldon-core-s2i-python3-ngraph-onnx

SELDON_CORE_DIR=../../..

.PHONY: build
build: 
	docker build  -t $(IMAGE_NAME):$(IMAGE_VERSION) .

push_to_dockerhub:
	docker push $(IMAGE_NAME):$(IMAGE_VERSION)
