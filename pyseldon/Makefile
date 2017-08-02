DOCKER_IMAGE_NAME=seldonio/pyseldon2
DOCKER_IMAGE_VERSION=3.1

install:
	python setup.py install

test:
	nosetests -s -v seldon

build_docker_image:
	docker build --force-rm=true -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION) .

push_to_registry:
	docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION)


