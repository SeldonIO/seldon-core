IMAGE_NAME=seldonio/gpu-mnist-example
IMAGE_VERSION=0.1


build_image:
	s2i build . joelh1996/gpu-base:0.1 ${IMAGE_NAME}:${IMAGE_VERSION}

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
