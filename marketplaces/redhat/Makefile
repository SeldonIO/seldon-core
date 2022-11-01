SHELL := /bin/bash

VERSION ?= $(shell cat ../../version.txt)
MLSERVER_VERSION ?= 1.2.0.dev12

IMG_MLSERVER ?= docker.io/seldonio/mlserver:${MLSERVER_VERSION}
IMG_MLSERVER_SC ?= docker.io/seldonio/mlserver-sc:${VERSION}
IMG_MLSERVER_SC_SLIM ?= docker.io/seldonio/mlserver-sc-slim:${VERSION}

redhat-image-scan: scan-mlserver scan-mlserver-sc scan-mlserver-sc-slim

# password can be found at: https://connect.redhat.com/projects/63566bb9822ce8cef9ba27fc/overview
project-mlserver=63566bb9822ce8cef9ba27fc
scan-mlserver:
	docker pull ${IMG_MLSERVER}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_mlserver} | docker login -u redhat-isv-containers+${project-mlserver}-robot quay.io --password-stdin
	docker tag ${IMG_MLSERVER} quay.io/redhat-isv-containers/${project-mlserver}:${MLSERVER_VERSION}
	docker push quay.io/redhat-isv-containers/${project-mlserver}:${MLSERVER_VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		preflight check container quay.io/redhat-isv-containers/${project-mlserver}:${MLSERVER_VERSION} --docker-config=${HOME}/.docker/config.json --certification-project-id=${project-mlserver} --pyxis-api-token=$${pyxis_api_token} --submit


# password can be found at: https://connect.redhat.com/projects/635670d3624969b495b6936f/overview
project-mlserver-sc=635670d3624969b495b6936f
scan-mlserver-sc:
	docker pull ${IMG_MLSERVER_SC}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_mlserver_sc} | docker login -u redhat-isv-containers+${project-mlserver-sc}-robot quay.io --password-stdin
	docker tag ${IMG_MLSERVER_SC} quay.io/redhat-isv-containers/${project-mlserver-sc}:${VERSION}
	docker push quay.io/redhat-isv-containers/${project-mlserver-sc}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		preflight check container quay.io/redhat-isv-containers/${project-mlserver-sc}:${VERSION} --docker-config=${HOME}/.docker/config.json --certification-project-id=${project-mlserver-sc} --pyxis-api-token=$${pyxis_api_token} --submit


# password can be found at: https://connect.redhat.com/projects/63567143624969b495b69370/overview
project-mlserver-sc-slim=63567143624969b495b69370
scan-mlserver-sc-slim:
	docker pull ${IMG_MLSERVER_SC_SLIM}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_mlserver_sc_slim} | docker login -u redhat-isv-containers+${project-mlserver-sc-slim}-robot quay.io --password-stdin
	docker tag ${IMG_MLSERVER_SC_SLIM} quay.io/redhat-isv-containers/${project-mlserver-sc-slim}:${VERSION}
	docker push quay.io/redhat-isv-containers/${project-mlserver-sc-slim}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		rm -rf ./artifacts && \
		preflight check container quay.io/redhat-isv-containers/${project-mlserver-sc-slim}:${VERSION} --docker-config=${HOME}/.docker/config.json --certification-project-id=${project-mlserver-sc-slim} --pyxis-api-token=$${pyxis_api_token} --submit
