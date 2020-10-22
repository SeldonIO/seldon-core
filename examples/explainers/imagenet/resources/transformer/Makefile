IMAGE_VERSION=0.1
IMAGE_NAME=docker.io/seldonio/imagenet-transformer

build_grpc:
	s2i build -E environment_grpc . seldonio/seldon-core-s2i-python36:1.4.0-dev_20201022145329 $(IMAGE_NAME):$(IMAGE_VERSION)

build_rest:
	s2i build -E environment_rest . seldonio/seldon-core-s2i-python36:1.4.0-dev_20201022145329 $(IMAGE_NAME):$(IMAGE_VERSION)

push_to_dockerhub:
	docker push $(IMAGE_NAME):$(IMAGE_VERSION)
