IMAGE_VERSION=0.1
IMAGE_NAME = seldonio/output_transformer

build:
	s2i build . seldonio/seldon-core-s2i-python3:1.3.0_20200928150014 $(IMAGE_NAME):$(IMAGE_VERSION)
