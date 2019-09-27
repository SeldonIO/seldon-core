VERSION=0.12-SNAPSHOT

#TODO
# Need to pass version into Dockerfile to get matching seldon-core version
# Need to ensure tensorflow version installed is the same as that for python library
# or wait until tensorflow is removed from python library

docker-build:
	docker build -t seldonio/seldon-core-s2i-python3-tf-gpu:${VERSION} .

docker-push:
	docker push seldonio/seldon-core-s2i-python3-tf-gpu:${VERSION}
