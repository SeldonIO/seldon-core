IMAGE_VERSION=0.1
IMAGE_NAME = docker.io/seldonio/sagemaker-proxy

SELDON_CORE_DIR=../../..

.PHONY: build
build: 
	s2i build -E environment  . seldonio/seldon-core-s2i-python3:0.18 $(IMAGE_NAME):$(IMAGE_VERSION)

push_to_dockerhub:
	docker push $(IMAGE_NAME):$(IMAGE_VERSION)

