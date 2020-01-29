IMAGE_NAME=seldon-java-wrapper
VERSION := $(shell cat ../version.txt)

version:
	echo ${VERSION}

build_jar: update_proto
	@set -x && mvn clean verify package -B

download_protos_tensorflow:
	cd ../proto/tensorflow ; make create_protos

update_proto: download_protos_tensorflow
	cp -v ../proto/prediction.proto src/main/proto/
	cp -vr ../proto/tensorflow/tensorflow src/main/proto

clean:
	mvn clean -B
	rm -fr src/main/proto/*
