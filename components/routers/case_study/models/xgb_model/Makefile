IMAGE_VERSION=0.1
IMAGE_NAME=seldonio/credit_default_xgb_model

build:
	s2i build . seldonio/seldon-core-s2i-python3:0.4 $(IMAGE_NAME):$(IMAGE_VERSION)

push_to_dockerhub:
	docker push $(IMAGE_NAME):$(IMAGE_VERSION)
