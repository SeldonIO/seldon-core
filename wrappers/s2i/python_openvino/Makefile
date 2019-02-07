IMAGE_VERSION=0.1
IMAGE_NAME = docker.io/seldonio/seldon-core-s2i-openvino

SELDON_CORE_DIR=../../..

s2i:
	cp -R ../python/s2i .

.PHONY: build
build: s2i
	docker build -f Dockerfile_openvino_base -t $(IMAGE_NAME):$(IMAGE_VERSION) .
	#docker build -f Dockerfile_openvino_base --build-arg http_proxy=$http_proxy --build-arg https_proxy=$https_proxy -t $(IMAGE_NAME):$(IMAGE_VERSION) .

push_to_dockerhub:
	docker push $(IMAGE_NAME):$(IMAGE_VERSION)
