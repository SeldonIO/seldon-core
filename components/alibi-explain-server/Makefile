SHELL := /bin/bash
VERSION := $(shell cat ../../version.txt)
IMAGE=alibiexplainer

kfserving:
	git clone -b seldon_grpc_explainer https://github.com/seldonio/kfserving.git

docker-build: kfserving
	docker build --file=Dockerfile --force-rm=true -t seldonio/${IMAGE}:${VERSION} .

docker-push:
	docker push ${IMAGE}:${TAG}

kind-load:
	kind load docker-image ${IMAGE}:${TAG}

# password can be found at: https://connect.redhat.com/project/3987291/view
redhat-image-scan:
	docker pull seldonio/${IMAGE}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_alibi_explain} | docker login -u unused scan.connect.redhat.com --password-stdin
	docker tag seldonio/${IMAGE}:${VERSION} scan.connect.redhat.com/ospid-02f3e15b-c16f-4353-affa-61d5f3c6408b/${IMAGE}:${VERSION}
	docker push scan.connect.redhat.com/ospid-02f3e15b-c16f-4353-affa-61d5f3c6408b/${IMAGE}:${VERSION}

clean:
	rm -rf kfserving
