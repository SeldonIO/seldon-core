IMAGE_VERSION=0.1
IMAGE_NAME = seldonio/output_transformer

build:
	s2i build . seldonio/seldon-core-s2i-python3:1.2.3-dev_20200817172210 $(IMAGE_NAME):$(IMAGE_VERSION)
