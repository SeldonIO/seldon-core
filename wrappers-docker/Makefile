IMAGE_NAME=docker.io/seldonio/core-python-wrapper
IMAGE_VERSION=0.6

SELDON_CORE_DIR=..

get_wrappers_and_protos:
	cp -R $(SELDON_CORE_DIR)/wrappers _wrappers
	cp $(SELDON_CORE_DIR)/proto/prediction.proto _wrappers/python/proto

build_image: clean get_wrappers_and_protos
	docker build --force-rm=true -t $(IMAGE_NAME):$(IMAGE_VERSION) .
	@docker tag $(IMAGE_NAME):$(IMAGE_VERSION) $(IMAGE_NAME):latest

push_to_dockerhub:
	@ \
		docker push $(IMAGE_NAME):$(IMAGE_VERSION) && \
		docker push $(IMAGE_NAME):latest

clean:
	rm -rfv _wrappers

