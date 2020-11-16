VERSION := $(shell cat ../../../version.txt)

build_images:
	s2i build -E environment_v1 . seldonio/seldon-core-s2i-python3:${VERSION} seldonio/fixed-model:0.1
	s2i build -E environment_v2 . seldonio/seldon-core-s2i-python3:${VERSION}  seldonio/fixed-model:0.2

push_images:
	docker push seldonio/fixed-model:0.1
	docker push seldonio/fixed-model:0.2


kind_load_images: build_images
	kind load -v 3 docker-image seldonio/fixed-model:0.1
	kind load -v 3 docker-image seldonio/fixed-model:0.2
