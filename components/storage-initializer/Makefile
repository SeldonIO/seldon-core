SHELL := /bin/bash
VERSION := $(shell cat ../../version.txt)
IMAGE=storage-initializer

kfserving:
	git clone -b seldon_grpc_explainer https://github.com/seldonio/kfserving.git

docker-build: kfserving
	docker build --file=Dockerfile --force-rm=true -t seldonio/${IMAGE}:${VERSION} .

docker-push:
	docker push ${IMAGE}:${TAG}

kind-load:
	kind load docker-image ${IMAGE}:${TAG}

# password can be found at: https://connect.redhat.com/project/3986991/view
redhat-image-scan:
	docker pull seldonio/${IMAGE}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_storage_initializer} | docker login -u unused scan.connect.redhat.com --password-stdin
	docker tag seldonio/${IMAGE}:${VERSION} scan.connect.redhat.com/ospid-dc132f00-5370-46ed-b506-737bb66d3f34/${IMAGE}:${VERSION}
	docker push scan.connect.redhat.com/ospid-dc132f00-5370-46ed-b506-737bb66d3f34/${IMAGE}:${VERSION}

clean:
	rm -rf kfserving
