IMAGE_VERSION=0.3
IMAGE_NAME=docker.io/seldonio/openvino-demo-prediction

build:
	s2i build -E environment_grpc . seldonio/seldon-core-s2i-openvino:0.2 $(IMAGE_NAME):$(IMAGE_VERSION)

push_to_dockerhub:
	docker push $(IMAGE_NAME):$(IMAGE_VERSION)
