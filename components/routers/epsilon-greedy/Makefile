IMAGE_VERSION=1.3
IMAGE_NAME=seldonio/mab_epsilon_greedy

build:
	s2i build . seldonio/seldon-core-s2i-python3:0.4 $(IMAGE_NAME):$(IMAGE_VERSION)

push_to_dockerhub:
	docker push $(IMAGE_NAME):$(IMAGE_VERSION)
