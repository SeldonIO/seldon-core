IMAGE_VERSION=0.2
IMAGE_NAME = docker.io/seldonio/seldon-core-s2i-openvino

SELDON_CORE_DIR=../../..

s2i:
	cp -R ../python/s2i .

.PHONY: build
build: s2i
	docker build -f Dockerfile_openvino_base -t $(IMAGE_NAME):$(IMAGE_VERSION) .

push_to_dockerhub:
	docker push $(IMAGE_NAME):$(IMAGE_VERSION)
