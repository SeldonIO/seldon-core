VERSION := $(shell cat ../version.txt)
IMAGE_NAME_BASE=engine
IMG ?= seldonio/${IMAGE_NAME_BASE}:${VERSION}
IMG_VERSION_REDHAT ?= ${IMAGE_NAME_BASE}-ubi8:1.1.0
IMG_REDHAT ?= seldonio/${IMG_VERSION_REDHAT}

version:
	echo ${VERSION}

build_jar: update_proto update_swagger
	mvn clean verify -Dlicense.useMissingFile -B

build_image: update_proto update_swagger
	docker build  -f Dockerfile -t ${IMG} .

docker-build-redhat: update_proto update_swagger
	docker build  -f Dockerfile.redhat -t ${IMG_REDHAT} .

# password can be found at: https://connect.redhat.com/project/1366491/view
redhat-image-scan:
	docker login -u unused scan.connect.redhat.com
	docker tag ${IMG_REDHAT} scan.connect.redhat.com/ospid-0d5322f9-b490-461e-a93f-1a88f69c1251/${IMG_VERSION_REDHAT}
	docker push scan.connect.redhat.com/ospid-0d5322f9-b490-461e-a93f-1a88f69c1251/${IMG_VERSION_REDHAT}

push_to_registry:
	docker push ${IMG}

clean:
	@mvn clean
	@rm -fv engine.json
	@rm -fr src/main/proto/*

cache_dependencies:
	mvn -Dmaven.repo.local=./.m2 dependency:resolve
	mvn -Dmaven.repo.local=./.m2 verify

download_protos_k8s:
	cd ../proto/k8s ; make create_protos

download_protos_tensorflow:
	cd ../proto/tensorflow ; make create_protos

update_proto: download_protos_k8s download_protos_tensorflow
	cp -v ../proto/seldon_deployment.proto src/main/proto/
	cp -v ../proto/prediction.proto src/main/proto/
	cp -vr ../proto/k8s/k8s.io src/main/proto
	cp -vr ../proto/tensorflow/tensorflow src/main/proto

update_swagger:
	cp -v ../openapi/engine.oas3.json src/main/resources/static/seldon.json

code_coverage: update_proto
	mvn clean verify
	cd target/site && aws s3 cp jacoco s3://seldon-core-docs/master/engine/jacoco --recursive --acl public-read


dependencies: update_proto
	mvn clean project-info-reports:dependencies
	cd target/site && aws s3 cp dependencies.html s3://seldon-core-docs/master/engine/dependencies/dependencies.html --acl public-read
	cd target/site && aws s3 cp css s3://seldon-core-docs/master/engine/dependencies/css --recursive --acl public-read
	cd target/site && aws s3 cp images s3://seldon-core-docs/master/engine/dependencies/images --recursive --acl public-read

kind-image-install: build_image
	kind load -v 3 docker-image ${IMG}

kind-image-install-redhat: build_image_redhat
	kind load -v 3 docker-image ${IMG_REDHAT}
