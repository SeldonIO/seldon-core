IMAGE_NAME=seldonio/tf-example-mnist
IMAGE_VERSION=0.1


build_image:
	s2i build . seldonio/seldon-core-s2i-python3:0.5 ${IMAGE_NAME}:${IMAGE_VERSION}

push_image:
	docker push $(IMAGE_NAME):$(IMAGE_VERSION)


.PHONY: train
train:
	mkdir -p model
	python create_model.py

.PHONY: clean
clean:
	rm -rf model
	rm -rf MNIST_data
