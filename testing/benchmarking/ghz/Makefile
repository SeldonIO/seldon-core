VERSION=v0.95.0
IMG=seldonio/ghz:${VERSION}

docker-build:
	docker build -f Dockerfile --build-arg GHZ_VERSION=${VERSION} -t ${IMG} .

docker-push:
	docker push ${IMG}
