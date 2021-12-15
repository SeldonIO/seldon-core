# Makefile for building and pushing fixed models docker images using for testing

build_image:
	mlserver build . --tag seldonio/mock-preprocessor:0.1.0

push_image:
	docker push seldonio/mock-preprocessor:0.1.0

kind_load_image: build_image
	kind load -v 3 docker-image seldonio/mock-preprocessor:0.1.0
