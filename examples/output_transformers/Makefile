IMAGE_VERSION=0.1
IMAGE_NAME = seldonio/output_transformer

build:
	s2i build . seldonio/seldon-core-s2i-python37-ubi8:1.7.0-dev $(IMAGE_NAME):$(IMAGE_VERSION)
