IMAGE_NAME=seldonio/sklearn-iris
SELDON_VERSION=$(shell cat ../../../version.txt)
IMAGE_VERSION=0.3
KIND_CLUSTER=ansible

build_image: train
	s2i build -E environment . seldonio/seldon-core-s2i-python37-ubi8:${SELDON_VERSION} ${IMAGE_NAME}:${IMAGE_VERSION}

push_image:
	docker push $(IMAGE_NAME):$(IMAGE_VERSION)

kind_load:
	kind load -v 3 docker-image ${IMAGE_NAME}:${IMAGE_VERSION} --name ${KIND_CLUSTER}

.PHONY: train
train:
	python train_iris.py

.PHONY: clean
clean:
	rm -rf IrisClassifier.sav
