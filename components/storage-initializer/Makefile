SHELL := /bin/bash
VERSION := $(shell cat ../../version.txt)
IMAGE=storage-initializer

kfserving:
	git clone -b seldon_grpc_explainer https://github.com/seldonio/kfserving.git

docker-build: kfserving
	docker build --file=Dockerfile --force-rm=true -t seldonio/${IMAGE}:${VERSION} .

docker-push:
	docker push seldonio/${IMAGE}:${VERSION}

kind_load: docker-build
	kind load docker-image seldonio/${IMAGE}:${VERSION}

# password can be found at: https://connect.redhat.com/project/3986991/view
scan=ospid-dc132f00-5370-46ed-b506-737bb66d3f34
project=5e9c7eea929b3d289e60977a
redhat-image-scan:
	docker pull seldonio/${IMAGE}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_storage_initializer} | docker login -u unused scan.connect.redhat.com --password-stdin
	docker tag seldonio/${IMAGE}:${VERSION} scan.connect.redhat.com/${scan}/${IMAGE}:${VERSION}
	docker push scan.connect.redhat.com/${scan}/${IMAGE}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		preflight check container scan.connect.redhat.com/${scan}/${IMAGE}:${VERSION} --docker-config=${HOME}/.docker/config.json --certification-project-id=${project} --pyxis-api-token=$${pyxis_api_token} --submit

clean:
	rm -rf kfserving
