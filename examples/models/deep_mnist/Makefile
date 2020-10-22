IMAGE_NAME=seldonio/tf-example-mnist
IMAGE_VERSION=0.2

build_image:
	s2i build . seldonio/seldon-core-s2i-python3:1.4.0-dev_20201022145329 ${IMAGE_NAME}:${IMAGE_VERSION}

push_image:
	docker push $(IMAGE_NAME):$(IMAGE_VERSION)

kind_load:
	kind load -v 3 docker-image ${IMAGE_NAME}:${IMAGE_VERSION}

.PHONY: train
train:
	mkdir -p model
	python create_model.py

.PHONY: clean
clean:
	rm -rf model
	rm -rf MNIST_data
