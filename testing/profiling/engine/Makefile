VERSION=0.2.6-SNAPSHOT
IMAGE_NAME=engine-debug

build_image: 
	docker build --build-arg SELDON_ENGINE_IMAGE=seldonio/engine:${VERSION} -t seldonio/${IMAGE_NAME}:latest .
	docker tag seldonio/$(IMAGE_NAME):latest seldonio/$(IMAGE_NAME):${VERSION}
