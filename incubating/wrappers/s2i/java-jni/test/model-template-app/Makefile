VERSION := 0.1.0
IMAGE_NAME_BASE=jni-model-template
IMAGE=seldonio/${IMAGE_NAME_BASE}:${VERSION}

build_jar: 
	mvn clean package -B

build:
	s2i build . \
		seldonio/s2i-java-jni-build:0.5.0 \
		--runtime-image seldonio/s2i-java-jni-runtime:0.5.0 \
		${IMAGE}

run:
	docker run \
		--rm -it \
		-p 5000:5000 \
		${IMAGE}

clean:
	mvn clean


