VERSION := 0.2.0
IMAGE_NAME_BASE=jni-model-template
IMAGE=seldonio/${IMAGE_NAME_BASE}:${VERSION}

build_jar: 
	mvn clean package -B

build:
	s2i build . \
		seldonio/s2i-java-jni-build:0.5.1 \
		--runtime-image seldonio/s2i-java-jni-runtime:0.5.1 \
		${IMAGE}

run:
	docker run \
		--rm -it \
		-p 9000:9000 \
		${IMAGE}

clean:
	mvn clean


