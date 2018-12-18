IMAGE_VERSION=0.3
IMAGE_NAME = seldonio/outlier_mahalanobis

build:
	s2i build . seldonio/seldon-core-s2i-python3:0.4 $(IMAGE_NAME):$(IMAGE_VERSION)

push_to_dockerhub:
	docker push $(IMAGE_NAME):$(IMAGE_VERSION)


