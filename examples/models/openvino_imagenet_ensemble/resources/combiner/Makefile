IMAGE_VERSION=0.1
IMAGE_NAME=docker.io/seldonio/openvino-demo-combiner

build:
	s2i build -E environment_grpc . seldonio/seldon-core-s2i-openvino:0.1 $(IMAGE_NAME):$(IMAGE_VERSION)

push_to_dockerhub:
	docker push $(IMAGE_NAME):$(IMAGE_VERSION)

