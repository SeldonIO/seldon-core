IMAGE_VERSION=0.7
IMAGE_NAME = docker.io/seldonio/tfserving-proxy

SELDON_CORE_DIR=../../..

.PHONY: build_rest
build_rest: 
	s2i build -E environment_rest . seldonio/seldon-core-s2i-python3:0.13 $(IMAGE_NAME)_rest:$(IMAGE_VERSION)

.PHONY: build_grpc
build_grpc: 
	s2i build -E environment_grpc . seldonio/seldon-core-s2i-python3:0.13 $(IMAGE_NAME)_grpc:$(IMAGE_VERSION)


push_to_dockerhub_rest:
	docker push $(IMAGE_NAME)_rest:$(IMAGE_VERSION)

push_to_dockerhub_grpc:
	docker push $(IMAGE_NAME)_grpc:$(IMAGE_VERSION)

