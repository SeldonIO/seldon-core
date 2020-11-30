BASE_VERSION := $(shell cat ../../../../../version.txt)
IMAGE_VERSION=0.2
IMAGE_NAME=seldonio/credit_default_rf_model

build:
	s2i build . seldonio/seldon-core-s2i-python37-ubi8:$(BASE_VERSION) $(IMAGE_NAME):$(IMAGE_VERSION)

push_to_dockerhub:
	docker push $(IMAGE_NAME):$(IMAGE_VERSION)
