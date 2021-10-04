IMAGE_VERSION=0.3
IMAGE_NAME=docker.io/seldonio/imagenet-transformer

build:
	s2i build -E environment . seldonio/seldon-core-s2i-python36:1.11.1 $(IMAGE_NAME):$(IMAGE_VERSION)

push_to_dockerhub:
	docker push $(IMAGE_NAME):$(IMAGE_VERSION)
